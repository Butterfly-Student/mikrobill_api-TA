package mikrotik

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
)

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

type mikrotikDomain struct {
	databasePort outbound_port.DatabasePort
}

func NewMikrotikDomain(
	databasePort outbound_port.DatabasePort,
) MikrotikDomain {
	return &mikrotikDomain{
		databasePort: databasePort,
	}
}

func (d *mikrotikDomain) Create(ctx context.Context, input model.MikrotikInput) (*model.Mikrotik, error) {
	mikrotikPort := d.databasePort.Mikrotik()
	result, err := mikrotikPort.Create(input)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik")
	}
	return result, nil
}

func (d *mikrotikDomain) GetByID(ctx context.Context, id uuid.UUID) (*model.Mikrotik, error) {
	mikrotikPort := d.databasePort.Mikrotik()
	result, err := mikrotikPort.GetByID(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mikrotik by id")
	}
	return result, nil
}

func (d *mikrotikDomain) List(ctx context.Context, filter model.MikrotikFilter) ([]model.Mikrotik, error) {
	mikrotikPort := d.databasePort.Mikrotik()
	results, err := mikrotikPort.List(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list mikrotik")
	}
	return results, nil
}

func (d *mikrotikDomain) Update(ctx context.Context, id uuid.UUID, input model.MikrotikUpdateInput) (*model.Mikrotik, error) {
	mikrotikPort := d.databasePort.Mikrotik()
	result, err := mikrotikPort.Update(id, input)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update mikrotik")
	}
	return result, nil
}

func (d *mikrotikDomain) Delete(ctx context.Context, id uuid.UUID) error {
	mikrotikPort := d.databasePort.Mikrotik()
	err := mikrotikPort.Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete mikrotik")
	}
	return nil
}

func (d *mikrotikDomain) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MikrotikStatus) error {
	mikrotikPort := d.databasePort.Mikrotik()
	err := mikrotikPort.UpdateStatus(id, status)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik status")
	}
	return nil
}

func (d *mikrotikDomain) UpdateLastSync(ctx context.Context, id uuid.UUID) error {
	mikrotikPort := d.databasePort.Mikrotik()
	err := mikrotikPort.UpdateLastSync(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik last sync")
	}
	return nil
}

func (d *mikrotikDomain) GetActiveMikrotik(ctx context.Context) (*model.Mikrotik, error) {
	mikrotikPort := d.databasePort.Mikrotik()
	result, err := mikrotikPort.GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	return result, nil
}

func (d *mikrotikDomain) SetActive(ctx context.Context, id uuid.UUID) error {
	mikrotikPort := d.databasePort.Mikrotik()
	err := mikrotikPort.SetActive(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to set mikrotik active")
	}
	return nil
}

func (d *mikrotikDomain) DeactivateAll(ctx context.Context) error {
	mikrotikPort := d.databasePort.Mikrotik()
	err := mikrotikPort.DeactivateAll()
	if err != nil {
		return stacktrace.Propagate(err, "failed to deactivate all mikrotik")
	}
	return nil
}
