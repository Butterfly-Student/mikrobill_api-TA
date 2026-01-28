package outbound_port

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go

import (
	"MikrOps/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type CustomerDatabasePort interface {
	// CreateCustomer creates a new customer
	CreateCustomer(ctx context.Context, input model.CreateCustomerRequest, mikrotikID uuid.UUID, mikrotikObjectID string) (*model.Customer, error)

	// CreateCustomerService creates a customer service subscription
	CreateCustomerService(ctx context.Context, customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error)

	// UpdateMikrotikObjectID updates the mikrotik object ID
	UpdateMikrotikObjectID(ctx context.Context, customerID uuid.UUID, objectID string) error

	// GetByID retrieves a customer by ID with service
	GetByID(ctx context.Context, id uuid.UUID) (*model.Customer, error)

	// GetByUsername retrieves a customer by username
	GetByUsername(ctx context.Context, mikrotikID uuid.UUID, username string) (*model.Customer, error)

	// List retrieves all customers for a MikroTik
	List(ctx context.Context, mikrotikID uuid.UUID) ([]model.Customer, error)

	// Update updates customer details
	Update(ctx context.Context, id uuid.UUID, input model.CreateCustomerRequest, price, taxRate float64) error

	// Delete removes a customer
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByPPPoEUsername retrieves a customer by PPPoE username
	GetByPPPoEUsername(ctx context.Context, username string) (*model.Customer, error)

	// UpdateStatus updates customer status and connection info
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.CustomerStatus, ip, mac, interfaceName *string) error

	// CreateProspect creates a prospect (customer without MikroTik provisioning)
	CreateProspect(ctx context.Context, input model.PublicRegistrationRequest, tenantID, mikrotikID uuid.UUID) (*model.Customer, error)

	// ListProspects retrieves all prospects for a MikroTik
	ListProspects(ctx context.Context, mikrotikID uuid.UUID) ([]model.Customer, error)

	// UpdateProspectToActive updates a prospect to active with MikroTik object ID
	UpdateProspectToActive(ctx context.Context, customerID uuid.UUID, mikrotikObjectID string, billingDay *int, autoSuspension *bool) error

	// UpdateServiceStartDate updates the start date of a customer's active service
	UpdateServiceStartDate(ctx context.Context, customerID uuid.UUID, startDate time.Time) error

	// ===========================================================================
	// CUSTOMER PORTAL & CREDENTIALS
	// ===========================================================================

	// GetByPortalEmail retrieves a customer by portal_email (for portal login)
	GetByPortalEmail(ctx context.Context, tenantID uuid.UUID, email string) (*model.Customer, error)

	// GetByServiceUsername retrieves a customer by service_username (for MikroTik callbacks)
	GetByServiceUsername(ctx context.Context, tenantID uuid.UUID, username string) (*model.Customer, error)

	// UpdatePortalPassword updates the portal_password_hash only
	UpdatePortalPassword(ctx context.Context, customerID uuid.UUID, passwordHash string) error

	// UpdateServiceCredentials sets service_username, service_password_encrypted, service_password_visible
	UpdateServiceCredentials(ctx context.Context, customerID uuid.UUID, username, encryptedPassword string, visible bool) error

	// UpdateProvisioningStatus updates provisioning workflow status
	UpdateProvisioningStatus(ctx context.Context, customerID uuid.UUID, status, errorMsg string, provisionedAt *time.Time) error
}
