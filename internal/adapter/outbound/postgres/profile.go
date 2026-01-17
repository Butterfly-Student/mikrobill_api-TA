package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	contextutil "MikrOps/utils/context"
)

const (
	tableProfiles      = "mikrotik_profiles"
	tableProfilesPPPoE = "mikrotik_profile_pppoe"
)

type profileAdapter struct {
	db *gorm.DB
}

func NewProfileAdapter(db *gorm.DB) outbound_port.ProfileDatabasePort {
	return &profileAdapter{db: db}
}

func (a *profileAdapter) CreateProfile(ctx context.Context, input model.CreateProfileRequest, mikrotikID uuid.UUID) (*model.MikrotikProfile, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	profile := &model.MikrotikProfile{
		TenantID:   tenantID.String(),
		MikrotikID: mikrotikID.String(),
		Name:       input.Name,
		Type:       input.Type,
		RateLimit:  input.RateLimit,
		Price:      input.Price,
		TaxRate:    0.0,
	}

	if input.TaxRate != nil {
		profile.TaxRate = *input.TaxRate
	}

	if err := a.db.WithContext(ctx).Create(profile).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to create profile")
	}

	return profile, nil
}

func (a *profileAdapter) CreateProfilePPPoE(ctx context.Context, profileID uuid.UUID, input model.CreateProfileRequest) error {
	pppoe := &model.MikrotikProfilePPPoE{
		ProfileID:      profileID.String(),
		LocalAddress:   input.LocalAddress,
		RemoteAddress:  input.RemoteAddress,
		AddressPool:    input.AddressPool,
		MTU:            1480,
		MRU:            1480,
		UseMPLS:        false,
		UseCompression: false,
		UseEncryption:  false,
	}

	if input.MTU != nil {
		pppoe.MTU = *input.MTU
	}
	if input.MRU != nil {
		pppoe.MRU = *input.MRU
	}

	if err := a.db.WithContext(ctx).Create(pppoe).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create profile pppoe")
	}

	return nil
}

func (a *profileAdapter) UpdateMikrotikObjectID(ctx context.Context, profileID uuid.UUID, objectID string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Model(&model.MikrotikProfile{}).
		Where("id = ? AND tenant_id = ?", profileID.String(), tenantID.String()).
		Update("mikrotik_object_id", objectID)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update mikrotik object id")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	return nil
}

func (a *profileAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.MikrotikProfile, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var profile model.MikrotikProfile
	if err := a.db.WithContext(ctx).
		Preload("PPPoEDetails").
		Preload("HotspotDetails").
		Preload("StaticIPDetails").
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	return &profile, nil
}

func (a *profileAdapter) GetByMikrotikID(ctx context.Context, mikrotikID, profileID uuid.UUID) (*model.MikrotikProfile, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var profile model.MikrotikProfile
	if err := a.db.WithContext(ctx).
		Preload("PPPoEDetails").
		Preload("HotspotDetails").
		Preload("StaticIPDetails").
		Where("id = ? AND mikrotik_id = ? AND tenant_id = ?", profileID.String(), mikrotikID.String(), tenantID.String()).
		First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, stacktrace.NewError("profile not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	return &profile, nil
}

func (a *profileAdapter) List(ctx context.Context, mikrotikID uuid.UUID) ([]model.MikrotikProfile, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	var profiles []model.MikrotikProfile
	if err := a.db.WithContext(ctx).
		Preload("PPPoEDetails").
		Preload("HotspotDetails").
		Preload("StaticIPDetails").
		Where("mikrotik_id = ? AND tenant_id = ?", mikrotikID.String(), tenantID.String()).
		Order("created_at DESC").
		Find(&profiles).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to list profiles")
	}

	return profiles, nil
}

func (a *profileAdapter) Update(ctx context.Context, id uuid.UUID, input model.CreateProfileRequest) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Update profile
	profileUpdates := map[string]interface{}{
		"name":       input.Name,
		"rate_limit": input.RateLimit,
		"price":      input.Price,
	}

	if input.TaxRate != nil {
		profileUpdates["tax_rate"] = *input.TaxRate
	}

	result := a.db.WithContext(ctx).
		Model(&model.MikrotikProfile{}).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Updates(profileUpdates)

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to update profile")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	// Update PPPoE settings if provided
	if input.Type == model.ProfileTypePPPoE {
		pppoeUpdates := map[string]interface{}{
			"local_address":  input.LocalAddress,
			"remote_address": input.RemoteAddress,
			"address_pool":   input.AddressPool,
		}

		if input.MTU != nil {
			pppoeUpdates["mtu"] = *input.MTU
		}
		if input.MRU != nil {
			pppoeUpdates["mru"] = *input.MRU
		}

		if err := a.db.WithContext(ctx).
			Model(&model.MikrotikProfilePPPoE{}).
			Where("profile_id = ?", id.String()).
			Updates(pppoeUpdates).Error; err != nil {
			return stacktrace.Propagate(err, "failed to update pppoe settings")
		}
	}

	return nil
}

func (a *profileAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	result := a.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id.String(), tenantID.String()).
		Delete(&model.MikrotikProfile{})

	if result.Error != nil {
		return stacktrace.Propagate(result.Error, "failed to delete profile")
	}

	if result.RowsAffected == 0 {
		return stacktrace.NewError("profile not found")
	}

	return nil
}

