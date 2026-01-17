package postgres_outbound_adapter

import (
	"context"

	"gorm.io/gorm"

	outbound_port "MikrOps/internal/port/outbound"
)

type adapter struct {
	db *gorm.DB
}

func NewAdapter(db *gorm.DB) outbound_port.DatabasePort {
	return &adapter{
		db: db,
	}
}

// DoInTransaction executes a function within a database transaction
func (s *adapter) DoInTransaction(ctx context.Context, txFunc outbound_port.InTransaction) (out interface{}, err error) {
	// Check if we're already in a transaction
	if tx := s.db.Statement.DB; tx != s.db {
		// Already in a transaction, just execute the function
		return txFunc(s)
	}

	// Start a new transaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new adapter with the transaction
		txAdapter := &adapter{db: tx}

		// Execute the function
		result, txErr := txFunc(txAdapter)
		out = result

		return txErr
	})

	return out, err
}

// Client returns the client database port
func (s *adapter) Client() outbound_port.ClientDatabasePort {
	return NewClientAdapter(s.db)
}

// Auth returns the auth database port
func (s *adapter) Auth() outbound_port.AuthDatabasePort {
	return NewAuthAdapter(s.db)
}

// Mikrotik returns the mikrotik database port
func (s *adapter) Mikrotik() outbound_port.MikrotikDatabasePort {
	return NewMikrotikAdapter(s.db)
}

// Profile returns the profile database port
func (s *adapter) Profile() outbound_port.ProfileDatabasePort {
	return NewProfileAdapter(s.db)
}

// Customer returns the customer database port
func (s *adapter) Customer() outbound_port.CustomerDatabasePort {
	return NewCustomerAdapter(s.db)
}

// Tenant returns the tenant database port
func (s *adapter) Tenant() outbound_port.TenantDatabasePort {
	return NewTenantAdapter(s.db)
}

// TenantUser returns the tenant user database port
func (s *adapter) TenantUser() outbound_port.TenantUserDatabasePort {
	return NewTenantUserAdapter(s.db)
}

// User returns the user database port
func (s *adapter) User() outbound_port.UserDatabasePort {
	return NewUserAdapter(s.db)
}
