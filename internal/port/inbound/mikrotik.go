package inbound_port

import (
	"context"

	"github.com/google/uuid"

	"prabogo/internal/model"
)

type MikrotikHttpPort interface {
	Create(a any) error
	GetByID(a any) error
	List(a any) error
	Update(a any) error
	Delete(a any) error
	UpdateStatus(a any) error
	GetActiveMikrotik(a any) error
	SetActive(a any) error
}

type MikrotikDomain interface {
	Create(ctx context.Context, input model.MikrotikInput) (*model.Mikrotik, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Mikrotik, error)
	List(ctx context.Context, filter model.MikrotikFilter) ([]model.Mikrotik, error)
	Update(ctx context.Context, id uuid.UUID, input model.MikrotikUpdateInput) (*model.Mikrotik, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.MikrotikStatus) error
	UpdateLastSync(ctx context.Context, id uuid.UUID) error
	GetActiveMikrotik(ctx context.Context) (*model.Mikrotik, error)
	SetActive(ctx context.Context, id uuid.UUID) error
	DeactivateAll(ctx context.Context) error
}
