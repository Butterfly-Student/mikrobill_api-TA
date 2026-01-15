package inbound_port

import (
	"context"

	"github.com/google/uuid"

	"prabogo/internal/model"
)

// TenantDomain defines the interface for tenant domain operations
type TenantDomain interface {
	CreateTenant(ctx context.Context, input model.TenantInput, createdBy uuid.UUID) (*model.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	ListTenants(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, input model.TenantInput, updatedBy uuid.UUID) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStats, error)
	CheckResourceLimit(ctx context.Context, tenantID uuid.UUID, resourceType string) error
}
