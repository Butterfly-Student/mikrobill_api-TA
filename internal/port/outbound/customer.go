package outbound_port

import (
	"prabogo/internal/model"
	"time"

	"github.com/google/uuid"
)

type CustomerDatabasePort interface {
	// CreateCustomer inserts a new customer to customers table
	CreateCustomer(input model.CustomerInput, mikrotikID uuid.UUID, mikrotikObjectID string) (*model.Customer, error)

	// CreateCustomerService inserts a new service to customer_services table
	CreateCustomerService(customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error)

	// UpdateMikrotikObjectID updates the mikrotik_object_id field in customers table
	UpdateMikrotikObjectID(customerID uuid.UUID, objectID string) error

	// GetByID retrieves a customer with service details by ID
	GetByID(id uuid.UUID) (*model.CustomerWithService, error)

	// GetByUsername retrieves a customer by username and mikrotik_id
	GetByUsername(mikrotikID uuid.UUID, username string) (*model.Customer, error)

	// GetByPPPoEUsername retrieves a customer by username across all mikrotiks or just matching username
	// Assuming PPPoE usernames are unique per Mikrotik, but callbacks usually don't send Mikrotik ID explicitly unless configured.
	// We'll search by username.
	GetByPPPoEUsername(username string) (*model.Customer, error)

	// List retrieves all customers for a MikroTik
	List(mikrotikID uuid.UUID) ([]model.CustomerWithService, error)

	// Update updates customer details
	Update(id uuid.UUID, input model.CustomerInput) error

	// UpdateStatus updates customer status and network info
	UpdateStatus(id uuid.UUID, status model.CustomerStatus, ip, mac, interfaceName *string) error

	// Delete removes a customer
	Delete(id uuid.UUID) error
}
