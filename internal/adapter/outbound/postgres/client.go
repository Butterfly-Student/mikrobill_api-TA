package postgres_outbound_adapter

import (
	"context"

	"github.com/palantir/stacktrace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
)

const tableClients = "clients"

type clientAdapter struct {
	db *gorm.DB
}

func NewClientAdapter(db *gorm.DB) outbound_port.ClientDatabasePort {
	return &clientAdapter{db: db}
}

func (a *clientAdapter) Upsert(ctx context.Context, datas []model.ClientInput) error {
	if len(datas) == 0 {
		return nil
	}

	var clients []model.Client
	for _, data := range datas {
		model.ClientPrepare(&data)
		clients = append(clients, model.Client{
			ClientInput: data,
		})
	}

	// Upsert using GORM's Clauses with OnConflict
	err := a.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "bearer_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
		}).
		Create(&clients).Error

	if err != nil {
		return stacktrace.Propagate(err, "failed to upsert clients")
	}

	return nil
}

func (a *clientAdapter) FindByFilter(ctx context.Context, filter model.ClientFilter, lock bool) ([]model.Client, error) {
	var clients []model.Client
	query := a.db.WithContext(ctx).Model(&model.Client{})

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.Names) > 0 {
		query = query.Where("name IN ?", filter.Names)
	}
	if len(filter.BearerKeys) > 0 {
		query = query.Where("bearer_key IN ?", filter.BearerKeys)
	}

	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	if err := query.Find(&clients).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to find clients")
	}

	return clients, nil
}

func (a *clientAdapter) DeleteByFilter(ctx context.Context, filter model.ClientFilter) error {
	query := a.db.WithContext(ctx).Model(&model.Client{})

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.Names) > 0 {
		query = query.Where("name IN ?", filter.Names)
	}
	if len(filter.BearerKeys) > 0 {
		query = query.Where("bearer_key IN ?", filter.BearerKeys)
	}

	if err := query.Delete(&model.Client{}).Error; err != nil {
		return stacktrace.Propagate(err, "failed to delete clients")
	}

	return nil
}

func (a *clientAdapter) IsExists(ctx context.Context, bearerKey string) (bool, error) {
	var count int64

	err := a.db.WithContext(ctx).
		Model(&model.Client{}).
		Where("bearer_key = ?", bearerKey).
		Count(&count).Error

	if err != nil {
		return false, stacktrace.Propagate(err, "failed to count client")
	}

	return count > 0, nil
}

