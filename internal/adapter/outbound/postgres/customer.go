package postgres_outbound_adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	contextutil "MikrOps/utils/context"
)

const (
	tableCustomers        = "customers"
	tableCustomerServices = "customer_services"
)

type customerAdapter struct {
	db *gorm.DB
}

func NewCustomerAdapter(db *gorm.DB) outbound_port.CustomerDatabasePort {
	return &customerAdapter{db: db}
}

func (a *customerAdapter) CreateCustomer(ctx context.Context, input model.CreateCustomerRequest, mikrotikID uuid.UUID, mikrotikObjectID string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	objectIDStr := mikrotikObjectID
	customer := &model.Customer{
		TenantID:         tenantID.String(),
		MikrotikID:       mikrotikID.String(),
		Username:         input.Username,
		Name:             input.Name,
		Phone:            input.Phone,
		Email:            input.Email,
		Address:          input.Address,
		MikrotikObjectID: &objectIDStr,
		ServiceType:      input.ServiceType,
		Status:           model.CustomerStatusInactive,
		AutoSuspension:   input.AutoSuspension != nil && *input.AutoSuspension,
		BillingDay:       1,
		JoinDate:         time.Now(),
		CustomerNotes:    input.CustomerNotes,
	}

	if input.BillingDay != nil {
		customer.BillingDay = *input.BillingDay
	}

	if err := a.db.WithContext(ctx).Create(customer).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer")
	}

	return customer, nil
}

func (a *customerAdapter) CreateCustomerService(ctx context.Context, customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	service := &model.CustomerService{
		TenantID:   tenantID.String(),
		CustomerID: customerID.String(),
		ProfileID:  profileID.String(),
		Price:      price,
		TaxRate:    taxRate,
		StartDate:  startDate,
		Status:     model.ServiceStatusActive,
	}

	if err := a.db.WithContext(ctx).Create(service).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer service")
	}

	return service, nil
}

func (a *customerAdapter) UpdateMikrotikObjectID(ctx context.Context, customerID uuid.UUID, objectID string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ? AND tenant_id = ?", customerID.String(), tenantID.String()).
		Update("mikrotik_object_id", objectID)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update mikrotik object id")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}

func (a *customerAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var customer model.Customer
	if err := a.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		First(&customer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	return &customer, nil
}

func (a *customerAdapter) GetByUsername(ctx context.Context, mikrotikID uuid.UUID, username string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var customer model.Customer
	err = a.db.WithContext(ctx).
		Where("mikrotik_id = ? AND username = ? AND tenant_id = ?", mikrotikID.String(), username, tenantID.String()).
		First(&customer).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // Not found is valid
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer by username")
	}

	return &customer, nil
}

func (a *customerAdapter) List(ctx context.Context, mikrotikID uuid.UUID) ([]model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var customers []model.Customer
	if err := a.db.WithContext(ctx).
		Where("mikrotik_id = ? AND tenant_id = ?", mikrotikID.String(), tenantID.String()).
		Order("created_at DESC").
		Find(&customers).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list customers")
	}

	return customers, nil
}

func (a *customerAdapter) Update(ctx context.Context, id uuid.UUID, input model.CreateCustomerRequest, price, taxRate float64) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Update customer
	customerUpdates := map[string]interface{}{
		"username":       input.Username,
		"name":           input.Name,
		"phone":          input.Phone,
		"email":          input.Email,
		"address":        input.Address,
		"service_type":   input.ServiceType,
		"customer_notes": input.CustomerNotes,
	}

	if input.AutoSuspension != nil {
		customerUpdates["auto_suspension"] = *input.AutoSuspension
	}
	if input.BillingDay != nil {
		customerUpdates["billing_day"] = *input.BillingDay
	}

	result := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Updates(customerUpdates)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update customer")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	// Update service
	serviceUpdates := map[string]interface{}{
		"profile_id": input.ProfileID,
		"price":      price,
		"tax_rate":   taxRate,
	}

	if input.StartDate != nil {
		serviceUpdates["start_date"] = *input.StartDate
	}

	if err := a.db.WithContext(ctx).
		Model(&model.CustomerService{}).
		Where("customer_id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Updates(serviceUpdates).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update service")
	}

	return nil
}

func (a *customerAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Delete(&model.Customer{})

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to delete customer")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}

func (a *customerAdapter) GetByPPPoEUsername(ctx context.Context, username string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var customer model.Customer
	err = a.db.WithContext(ctx).
		Where("username = ? AND tenant_id = ?", username, tenantID.String()).
		First(&customer).Error

	if err == gorm.ErrRecordNotFound {
		return nil, stacktrace.NewError("customer not found")
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer by pppoe username")
	}

	return &customer, nil
}

func (a *customerAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status model.CustomerStatus, ip, mac, interfaceName *string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if ip != nil {
		updates["assigned_ip"] = *ip
		updates["last_ip"] = *ip
		now := time.Now()
		updates["last_online"] = &now
	}
	if mac != nil {
		updates["mac_address"] = *mac
	}
	if interfaceName != nil {
		updates["interface"] = *interfaceName
	}

	result := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Updates(updates)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update customer status")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}

// CreateProspect creates a prospect (customer without MikroTik provisioning)
func (a *customerAdapter) CreateProspect(ctx context.Context, input model.PublicRegistrationRequest, tenantID, mikrotikID uuid.UUID) (*model.Customer, error) {
	customer := &model.Customer{
		TenantID:       tenantID.String(),
		MikrotikID:     mikrotikID.String(),
		Username:       input.Username,
		Name:           input.Name,
		Phone:          input.Phone,
		Email:          input.Email,
		Address:        input.Address,
		ServiceType:    input.ServiceType,
		Status:         model.CustomerStatusProspect, // PROSPECT status
		AutoSuspension: true,                         // Default, can be updated on approval
		BillingDay:     1,                            // Default, can be updated on approval
		JoinDate:       time.Now(),
		CustomerNotes:  input.CustomerNotes,
		// mikrotik_object_id is NULL - no provisioning yet
	}

	if err := a.db.WithContext(ctx).Create(customer).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to create prospect")
	}

	return customer, nil
}

// ListProspects retrieves all prospects for a MikroTik
func (a *customerAdapter) ListProspects(ctx context.Context, mikrotikID uuid.UUID) ([]model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var prospects []model.Customer
	if err := a.db.WithContext(ctx).
		Where("mikrotik_id = ? AND tenant_id = ? AND status = ?",
			mikrotikID.String(), tenantID.String(), model.CustomerStatusProspect).
		Preload("Services.Profile").
		Preload("Mikrotik").
		Order("created_at DESC").
		Find(&prospects).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list prospects")
	}

	return prospects, nil
}

// UpdateProspectToActive updates a prospect to active with MikroTik object ID
func (a *customerAdapter) UpdateProspectToActive(ctx context.Context, customerID uuid.UUID, mikrotikObjectID string, billingDay *int, autoSuspension *bool) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	updates := map[string]interface{}{
		"status":             model.CustomerStatusActive,
		"mikrotik_object_id": mikrotikObjectID,
		"updated_at":         time.Now(),
	}

	if billingDay != nil {
		updates["billing_day"] = *billingDay
	}
	if autoSuspension != nil {
		updates["auto_suspension"] = *autoSuspension
	}

	result := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("id = ? AND tenant_id = ?", customerID.String(), tenantID.String()).
		Updates(updates)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update prospect to active")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("prospect not found")
	}

	return nil
}

// UpdateServiceStartDate updates the start date of a customer's active service
func (a *customerAdapter) UpdateServiceStartDate(ctx context.Context, customerID uuid.UUID, startDate time.Time) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Model(&model.CustomerService{}).
		Where("customer_id = ? AND tenant_id = ? AND status = ?",
			customerID.String(), tenantID.String(), model.ServiceStatusActive).
		Update("start_date", startDate)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update service start date")
	}

	return nil
}
