package postgres_outbound_adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	contextutil "MikrOps/utils/context"
)

const tableMikrotik = "mikrotik"

type mikrotikAdapter struct {
	db *gorm.DB
}

func NewMikrotikAdapter(db *gorm.DB) outbound_port.MikrotikDatabasePort {
	return &mikrotikAdapter{db: db}
}

func (a *mikrotikAdapter) Create(ctx context.Context, input model.CreateMikrotikRequest) (*model.Mikrotik, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	mikrotik := &model.Mikrotik{
		TenantID:             tenantID.String(),
		Name:                 input.Name,
		Host:                 input.Host,
		Port:                 input.Port,
		APIUsername:          input.APIUsername,
		APIEncryptedPassword: &input.APIPassword, // Should be encrypted before calling this
		Keepalive:            input.Keepalive != nil && *input.Keepalive,
		Timeout:              300000, // default
		Location:             input.Location,
		Description:          input.Description,
		Status:               model.MikrotikStatusOffline,
		IsActive:             false,
	}

	if input.Keepalive != nil {
		mikrotik.Keepalive = *input.Keepalive
	}
	if input.Timeout != nil {
		mikrotik.Timeout = *input.Timeout
	}

	if err := a.db.WithContext(ctx).Create(mikrotik).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik")
	}

	return mikrotik, nil
}

func (a *mikrotikAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.Mikrotik, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var mikrotik model.Mikrotik
	err = a.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		First(&mikrotik).Error

	if err == gorm.ErrRecordNotFound {
		return nil, stacktrace.NewError("mikrotik not found")
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mikrotik")
	}

	return &mikrotik, nil
}

func (a *mikrotikAdapter) List(ctx context.Context) ([]model.Mikrotik, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var mikrotiks []model.Mikrotik
	if err := a.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Find(&mikrotiks).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list mikrotik")
	}

	return mikrotiks, nil
}

func (a *mikrotikAdapter) Update(ctx context.Context, id uuid.UUID, input model.UpdateMikrotikRequest) (*model.Mikrotik, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	updates := make(map[string]interface{})

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Host != nil {
		updates["host"] = *input.Host
	}
	if input.Port != nil {
		updates["port"] = *input.Port
	}
	if input.APIUsername != nil {
		updates["api_username"] = *input.APIUsername
	}
	if input.APIPassword != nil {
		updates["api_encrypted_password"] = *input.APIPassword // Should be encrypted
	}
	if input.Keepalive != nil {
		updates["keepalive"] = *input.Keepalive
	}
	if input.Timeout != nil {
		updates["timeout"] = *input.Timeout
	}
	if input.Location != nil {
		updates["location"] = *input.Location
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if len(updates) == 0 {
		return a.GetByID(ctx, id)
	}

	result := a.db.WithContext(ctx).
		Model(&model.Mikrotik{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Updates(updates)

	if result.Error != nil {
		return nil, stacktrace.Propagate(result.Error, "failed to update mikrotik")
	}

	if result.RowsAffected == 0 {
		return nil, stacktrace.NewError("mikrotik not found")
	}

	return a.GetByID(ctx, id)
}

func (a *mikrotikAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Delete(&model.Mikrotik{})

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to delete mikrotik")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MikrotikStatus) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Model(&model.Mikrotik{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Update("status", status)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update mikrotik status")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) UpdateLastSync(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	now := time.Now()
	result := a.db.WithContext(ctx).
		Model(&model.Mikrotik{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Update("last_sync", &now)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update mikrotik last sync")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) GetActiveMikrotik(ctx context.Context) (*model.Mikrotik, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var mikrotik model.Mikrotik
	err = a.db.WithContext(ctx).
		Where("is_active = ? AND tenant_id = ?", true, tenantID.String()).
		First(&mikrotik).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // No active mikrotik is valid
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	return &mikrotik, nil
}

func (a *mikrotikAdapter) SetActive(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Use transaction
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all for this tenant
		if err := tx.Model(&model.Mikrotik{}).
			Where("tenant_id = ?", tenantID.String()).
			Update("is_active", false).Error; err != nil {
			return stacktrace.Propagate(err, "failed to deactivate all mikrotik")
		}

		// Activate the specific one
		result := tx.Model(&model.Mikrotik{}).
			Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
			Update("is_active", true)

		if result.Error != nil {
			return stacktrace.Propagate(result.Error, "failed to activate mikrotik")
		}

		if result.RowsAffected == 0 {
			return stacktrace.NewError("mikrotik not found")
		}

		return nil
	})
}

func (a *mikrotikAdapter) DeactivateAll(ctx context.Context) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	if err := a.db.WithContext(ctx).
		Model(&model.Mikrotik{}).
		Where("tenant_id = ?", tenantID.String()).
		Update("is_active", false).Error; err != nil {
		return stacktrace.Propagate(err, "failed to deactivate all mikrotik")
	}

	return nil
}

