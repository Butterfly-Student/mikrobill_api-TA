package outbound_port

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go

import (
	"context"
	"MikrOps/internal/model"
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
}

