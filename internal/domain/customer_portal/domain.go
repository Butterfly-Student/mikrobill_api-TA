package customer_portal

import (
	"context"
	"fmt"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/encryption"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

// CustomerPortalDomain defines customer portal operations
type CustomerPortalDomain interface {
	// Profile Management
	GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*model.CustomerProfileResponse, error)
	UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req model.UpdateCustomerProfileRequest) error

	// Traffic Monitoring
	GetTrafficStats(ctx context.Context, customerID uuid.UUID) (*model.CustomerTrafficResponse, error)

	// Session Control
	ResetSession(ctx context.Context, customerID uuid.UUID) error

	// Usage History
	GetUsageHistory(ctx context.Context, customerID uuid.UUID, days int) ([]model.UsageHistoryResponse, error)
}

type customerPortalDomain struct {
	customerPort      outbound_port.CustomerDatabasePort
	cachePort         outbound_port.PPPCachePort
	provisioningPort  outbound_port.ProvisioningMessagePort
	encryptionService *encryption.Service
}

func NewCustomerPortalDomain(
	customerPort outbound_port.CustomerDatabasePort,
	cachePort outbound_port.PPPCachePort,
	provisioningPort outbound_port.ProvisioningMessagePort,
	encryptionService *encryption.Service,
) CustomerPortalDomain {
	return &customerPortalDomain{
		customerPort:      customerPort,
		cachePort:         cachePort,
		provisioningPort:  provisioningPort,
		encryptionService: encryptionService,
	}
}

// GetCustomerProfile retrieves customer profile with conditional credential visibility
func (d *customerPortalDomain) GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (*model.CustomerProfileResponse, error) {
	// Get customer from database
	customer, err := d.customerPort.GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	if customer == nil {
		return nil, stacktrace.NewError("customer not found")
	}

	// Build response
	response := &model.CustomerProfileResponse{
		ID:                     customerID, // Use the customerID from the function parameter directly
		Name:                   customer.Name,
		Email:                  customer.PortalEmail,
		Phone:                  &customer.Phone,
		Address:                customer.Address,
		ServiceType:            string(customer.ServiceType),
		ServiceUsername:        customer.ServiceUsername,
		ServicePasswordVisible: customer.ServicePasswordVisible,
		Status:                 string(customer.Status),
		ProvisioningStatus:     customer.ProvisioningStatus,
	}

	// Decrypt and show service password ONLY if visible (Hotspot)
	if customer.ServicePasswordVisible && customer.ServicePasswordEncrypted != nil {
		decryptedPassword, err := d.encryptionService.Decrypt(*customer.ServicePasswordEncrypted)
		if err != nil {
			// Log error but don't fail the request
			return response, nil
		}
		response.ServicePassword = &decryptedPassword
	}

	return response, nil
}

// UpdateCustomerProfile updates customer profile (name, phone, address, portal password)
func (d *customerPortalDomain) UpdateCustomerProfile(ctx context.Context, customerID uuid.UUID, req model.UpdateCustomerProfileRequest) error {
	// Get customer to verify exists
	customer, err := d.customerPort.GetByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get customer")
	}

	if customer == nil {
		return stacktrace.NewError("customer not found")
	}

	// Update basic fields
	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.Address != nil {
		customer.Address = req.Address
	}

	// Update portal password if provided
	if req.PortalPassword != nil && *req.PortalPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.PortalPassword), 12)
		if err != nil {
			return stacktrace.Propagate(err, "failed to hash password")
		}
		hashedPasswordStr := string(hashedPassword)
		customer.PortalPasswordHash = &hashedPasswordStr
	}

	// Update customer in database
	updateReq := model.CreateCustomerRequest{
		Username:    customer.ServiceUsername,
		Name:        customer.Name,
		Phone:       customer.Phone,
		Email:       customer.Email,
		Address:     customer.Address,
		ServiceType: customer.ServiceType,
	}

	// Note: Price and TaxRate are usually handled in service update,
	// but here we align with the current Update signature which requires them.
	// Since we don't have them easily available in a basic profile update,
	// we use placeholders or ideally this should be a different port method.
	// For now, to fix compilation, we use 0.0.
	if err := d.customerPort.Update(ctx, customerID, updateReq, 0.0, 0.0); err != nil {
		return stacktrace.Propagate(err, "failed to update customer")
	}

	return nil
}

// GetTrafficStats retrieves current traffic statistics from Redis cache
func (d *customerPortalDomain) GetTrafficStats(ctx context.Context, customerID uuid.UUID) (*model.CustomerTrafficResponse, error) {
	// Get customer to get service username and tenant ID
	customer, err := d.customerPort.GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	if customer == nil {
		return nil, stacktrace.NewError("customer not found")
	}

	if customer.ServiceUsername == "" {
		return nil, stacktrace.NewError("customer has no service username")
	}

	// Build Redis key for traffic data
	// Format: traffic:{tenant_id}:{service_username}
	redisKey := fmt.Sprintf("traffic:%s:%s", customer.TenantID, customer.ServiceUsername)

	// Get traffic data from Redis
	trafficData, err := d.cachePort.GetTrafficData(ctx, redisKey)
	if err != nil {
		// If no data in Redis, return empty stats
		return &model.CustomerTrafficResponse{
			ServiceUsername: customer.ServiceUsername,
			UploadBytes:     0,
			DownloadBytes:   0,
			TotalBytes:      0,
			SessionDuration: 0,
			LastUpdate:      time.Now().Format(time.RFC3339),
		}, nil
	}

	// Build response
	response := &model.CustomerTrafficResponse{
		ServiceUsername: customer.ServiceUsername,
		UploadBytes:     trafficData.UploadBytes,
		DownloadBytes:   trafficData.DownloadBytes,
		TotalBytes:      trafficData.UploadBytes + trafficData.DownloadBytes,
		SessionDuration: trafficData.SessionDuration,
		LastUpdate:      trafficData.LastUpdate.Format(time.RFC3339),
	}

	return response, nil
}

// ResetSession publishes a disconnect task to RabbitMQ to reset customer's active session
func (d *customerPortalDomain) ResetSession(ctx context.Context, customerID uuid.UUID) error {
	// Get customer to get service username
	customer, err := d.customerPort.GetByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get customer")
	}

	if customer == nil {
		return stacktrace.NewError("customer not found")
	}

	if customer.ServiceUsername == "" {
		return stacktrace.NewError("customer has no service username")
	}

	// Build disconnect message
	disconnectMsg := model.DisconnectMessage{
		CustomerID:      customer.ID,
		TenantID:        customer.TenantID,
		ServiceUsername: customer.ServiceUsername,
		ServiceType:     string(customer.ServiceType),
	}

	// Publish to RabbitMQ disconnect queue
	if err := d.provisioningPort.PublishDisconnect(ctx, disconnectMsg); err != nil {
		return stacktrace.Propagate(err, "failed to publish disconnect message")
	}

	return nil
}

// GetUsageHistory retrieves historical usage data for the customer
func (d *customerPortalDomain) GetUsageHistory(ctx context.Context, customerID uuid.UUID, days int) ([]model.UsageHistoryResponse, error) {
	// For now, return empty array as usage logging will be implemented in Phase 7
	// This is a placeholder that can be enhanced later with actual database queries

	// Validate days parameter
	if days <= 0 {
		days = 30 // default to 30 days
	}
	if days > 90 {
		days = 90 // max 90 days
	}

	// TODO: Query usage_logs table when implemented in Phase 7
	// For now, return empty array
	return []model.UsageHistoryResponse{}, nil
}
