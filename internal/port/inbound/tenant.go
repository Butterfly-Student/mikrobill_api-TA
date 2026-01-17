package inbound_port

import (
	"context"

	"github.com/google/uuid"

	"MikrOps/internal/model"
)

// TenantDomain defines the interface for tenant domain operations
type TenantDomain interface {
	CreateTenant(ctx context.Context, input model.CreateTenantRequest, createdBy uuid.UUID) (*model.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	ListTenants(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, input model.UpdateTenantRequest, updatedBy uuid.UUID) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStatsResponse, error)
	CheckResourceLimit(ctx context.Context, tenantID uuid.UUID, resourceType string) error
}

