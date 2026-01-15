package outbound_port

import (
	"context"

	"github.com/google/uuid"

	"prabogo/internal/model"
)

type TenantDatabasePort interface {
	CreateTenant(ctx context.Context, input model.TenantInput) (*model.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	List(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input model.TenantInput) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStats, error)
}
