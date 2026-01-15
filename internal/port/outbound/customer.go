package outbound_port

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go

import (
	"context"
	"prabogo/internal/model"
	"time"

	"github.com/google/uuid"
)

type CustomerDatabasePort interface {
	// CreateCustomer inserts a new customer to customers table
	CreateCustomer(ctx context.Context, input model.CustomerInput, mikrotikID uuid.UUID, mikrotikObjectID string) (*model.Customer, error)

	// CreateCustomerService inserts a new service to customer_services table
	CreateCustomerService(ctx context.Context, customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error)

	// UpdateMikrotikObjectID updates the mikrotik_object_id field in customers table
	UpdateMikrotikObjectID(ctx context.Context, customerID uuid.UUID, objectID string) error

	// GetByID retrieves a customer with service details by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.CustomerWithService, error)

	// GetByUsername retrieves a customer by username and mikrotik_id
	GetByUsername(ctx context.Context, mikrotikID uuid.UUID, username string) (*model.Customer, error)

	// GetByPPPoEUsername retrieves a customer by username across all mikrotiks within tenant
	GetByPPPoEUsername(ctx context.Context, username string) (*model.Customer, error)

	// List retrieves all customers for a MikroTik
	List(ctx context.Context, mikrotikID uuid.UUID) ([]model.CustomerWithService, error)

	// Update updates customer details
	Update(ctx context.Context, id uuid.UUID, input model.CustomerInput, price, taxRate float64) error

	// UpdateStatus updates customer status and network info
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.CustomerStatus, ip, mac, interfaceName *string) error

	// Delete removes a customer
	Delete(ctx context.Context, id uuid.UUID) error
}
