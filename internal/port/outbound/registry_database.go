package outbound_port

import (
	"context"
)

//go:generate mockgen -source=registry_database.go -destination=./../../../tests/mocks/port/mock_registry_database.go
type InTransaction func(repoRegistry DatabasePort) (interface{}, error)

type DatabasePort interface {
	Client() ClientDatabasePort
	Auth() AuthDatabasePort
	Mikrotik() MikrotikDatabasePort
	Profile() ProfileDatabasePort
	Customer() CustomerDatabasePort
	Tenant() TenantDatabasePort
	TenantUser() TenantUserDatabasePort
	User() UserDatabasePort
	DoInTransaction(ctx context.Context, txFunc InTransaction) (out interface{}, err error)
}
