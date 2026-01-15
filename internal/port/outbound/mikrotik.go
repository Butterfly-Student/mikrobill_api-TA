package outbound_port

//go:generate mockgen -source=mikrotik.go -destination=./../../../tests/mocks/port/mock_mikrotik.go

import (
	"context"
	"prabogo/internal/model"

	"github.com/google/uuid"
)

type MikrotikDatabasePort interface {
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
