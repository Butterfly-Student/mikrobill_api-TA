package outbound_port

import (
	"context"

	"github.com/google/uuid"

	"MikrOps/internal/model"
)

type TenantDatabasePort interface {
	CreateTenant(ctx context.Context, input model.CreateTenantRequest) (*model.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	List(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input model.UpdateTenantRequest) (*model.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStatsResponse, error)
}

