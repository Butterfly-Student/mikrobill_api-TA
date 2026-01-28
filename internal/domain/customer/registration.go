package customer

import (
	"MikrOps/internal/model"
	"MikrOps/utils/encryption"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"golang.org/x/crypto/bcrypt"
)

// ===========================================================================
// PUBLIC REGISTRATION & PROSPECT MANAGEMENT
// ===========================================================================

// RegisterProspect handles public self-registration (no MikroTik provisioning)
// Portal credentials are hashed and stored, service credentials are NULL
func (d *customerDomain) RegisterProspect(ctx context.Context, slug string, input model.PublicRegistrationRequest) (*model.Customer, error) {
	// 1. Get tenant by slug
	tenant, err := d.databasePort.Tenant().GetBySlug(ctx, slug)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant by slug")
	}
	if tenant == nil {
		return nil, fmt.Errorf("tenant not found with slug: %s", slug)
	}

	tenantID, err := uuid.Parse(tenant.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid tenant id")
	}

	// 2. Validate portal email uniqueness per tenant
	existingCustomer, err := d.databasePort.Customer().GetByPortalEmail(ctx, tenantID, input.PortalEmail)
	if err == nil && existingCustomer != nil {
		return nil, fmt.Errorf("customer with email '%s' already registered", input.PortalEmail)
	}

	// 3. Get active mikrotik for this tenant (prospects need to be assigned to a mikrotik)
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik configured for this tenant")
	}

	mikrotikID, err := uuid.Parse(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid mikrotik id")
	}

	// 4. Validate profile exists and belongs to this mikrotik
	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	profile, err := d.databasePort.Profile().GetByMikrotikID(ctx, mikrotikID, profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	// 5. Hash portal password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.PortalPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to hash password")
	}

	// 6. Create prospect in database (portal credentials set, service credentials NULL)
	prospect, err := d.databasePort.Customer().CreateProspect(ctx, input, tenantID, mikrotikID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create prospect")
	}

	// 7. Update portal password hash (CreateProspect doesn't set it)
	prospectID, err := uuid.Parse(prospect.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid prospect id")
	}

	if err := d.databasePort.Customer().UpdatePortalPassword(ctx, prospectID, string(hashedPassword)); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update portal password")
	}

	// 8. Create customer service subscription
	_, err = d.databasePort.Customer().CreateCustomerService(
		ctx,
		prospectID,
		profileID,
		profile.Price,
		profile.TaxRate,
		time.Now(), // Start date set to now, can be updated on approval
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create customer service")
	}

	// Reload prospect with services
	return d.databasePort.Customer().GetByID(ctx, prospectID)
}

// ApproveProspect approves a prospect and provisions to MikroTik
// Generates service credentials based on service_type and publishes to RabbitMQ for async provisioning
func (d *customerDomain) ApproveProspect(ctx context.Context, req model.ApproveProspectRequest) (*model.Customer, error) {
	// 1. Get prospect
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	if customer.Status != model.CustomerStatusProspect {
		return nil, fmt.Errorf("customer is not a prospect (status: %s)", customer.Status)
	}

	// 2. Generate service credentials based on service_type
	var serviceUsername, servicePassword string
	var servicePasswordVisible bool

	if customer.ServiceType == model.ServiceTypePPPoE {
		// PPPoE: Auto-generate strong credentials, HIDDEN from customer
		serviceUsername = generatePPPoEUsername(customer.ID, customer.Phone)
		servicePassword, err = encryption.GeneratePPPoEPassword() // 16-char strong
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate PPPoE password")
		}
		servicePasswordVisible = false
	} else if customer.ServiceType == model.ServiceTypeHotspot {
		// Hotspot: Use preferred or auto-generate, simpler password, VISIBLE to customer
		if req.PreferredServiceUsername != nil && *req.PreferredServiceUsername != "" {
			serviceUsername = *req.PreferredServiceUsername
		} else {
			serviceUsername = generateHotspotUsername(customer.Phone)
		}
		servicePassword, err = encryption.GenerateHotspotPassword() // 8-char simple
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate Hotspot password")
		}
		servicePasswordVisible = true
	} else {
		return nil, fmt.Errorf("unsupported service type: %s", customer.ServiceType)
	}

	// 3. Encrypt service password for storage
	encryptionService, err := encryption.NewService(os.Getenv("SERVICE_CREDENTIAL_KEY"))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize encryption service")
	}

	encryptedPassword, err := encryptionService.Encrypt(servicePassword)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encrypt service password")
	}

	// 4. Update customer with service credentials
	if err := d.databasePort.Customer().UpdateServiceCredentials(
		ctx,
		customerID,
		serviceUsername,
		encryptedPassword,
		servicePasswordVisible,
	); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update service credentials")
	}

	// 5. Update provisioning status to 'provisioning'
	if err := d.databasePort.Customer().UpdateProvisioningStatus(
		ctx,
		customerID,
		"provisioning",
		"",
		nil,
	); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update provisioning status")
	}

	// 6. Publish RabbitMQ message for async provisioning (PHASE 4)
	// TODO: Implement RabbitMQ publisher
	// Message should contain: {customer_id, tenant_id, profile_id, service_username, service_password_plain, service_type}
	// Worker will call MikroTik API, update customer status, and set mikrotik_object_id

	// 7. Reload customer
	return d.databasePort.Customer().GetByID(ctx, customerID)
}

// Helper functions for username generation
func generatePPPoEUsername(customerID, phone string) string {
	// Use short hash of customer ID + phone for uniqueness
	// Example: pppoe_abc123
	shortID := customerID[:8]
	return fmt.Sprintf("pppoe_%s", shortID)
}

func generateHotspotUsername(phone string) string {
	// Use phone number without leading zero
	// Example: hs_81234567890
	return fmt.Sprintf("hs_%s", phone)
}
