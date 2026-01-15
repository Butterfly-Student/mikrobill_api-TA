package outbound_port

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockgen -source=tenant_user.go -destination=./../../../tests/mocks/port/mock_tenant_user.go
type TenantUserDatabasePort interface {
	// HasAccess checks if user has access to the specified tenant
	HasAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)

	// GetPrimaryTenant retrieves the primary tenant for a user
	GetPrimaryTenant(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)

	// GetTenantsForUser retrieves all active tenants for a user
	GetTenantsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}
