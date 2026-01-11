package outbound_port

import (
	"prabogo/internal/model"

	"github.com/google/uuid"
)

type MikrotikDatabasePort interface {
	Create(input model.MikrotikInput) (*model.Mikrotik, error)
	GetByID(id uuid.UUID) (*model.Mikrotik, error)
	List(filter model.MikrotikFilter) ([]model.Mikrotik, error)
	Update(id uuid.UUID, input model.MikrotikUpdateInput) (*model.Mikrotik, error)
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status model.MikrotikStatus) error
	UpdateLastSync(id uuid.UUID) error
	GetActiveMikrotik() (*model.Mikrotik, error)
	SetActive(id uuid.UUID) error
	DeactivateAll() error
}
