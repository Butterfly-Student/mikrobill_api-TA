package repository

import (
	"context"
	"mikrobill/internal/entity"
	"time"

	"gorm.io/gorm"
)

type MikrotikRepository interface {
	Create(ctx context.Context, mk *entity.Mikrotik) error
	GetByID(ctx context.Context, id string) (*entity.Mikrotik, error)
	List(ctx context.Context, page, pageSize int, search string) ([]entity.Mikrotik, int64, error)
	Update(ctx context.Context, mk *entity.Mikrotik) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateLastSync(ctx context.Context, id string) error

	// Active Mikrotik Management
	GetActiveMikrotik(ctx context.Context) (*entity.Mikrotik, error)
	SetActive(ctx context.Context, id string, active bool) error
	DeactivateAll(ctx context.Context) error
}

type mikrotikRepository struct {
	db *gorm.DB
}

func NewMikrotikRepository(db *gorm.DB) MikrotikRepository {
	return &mikrotikRepository{db: db}
}

func (r *mikrotikRepository) Create(ctx context.Context, mk *entity.Mikrotik) error {
	return r.db.WithContext(ctx).Create(mk).Error
}

func (r *mikrotikRepository) GetByID(ctx context.Context, id string) (*entity.Mikrotik, error) {
	var mk entity.Mikrotik
	err := r.db.WithContext(ctx).First(&mk, "id = ?", id).Error
	return &mk, err
}

func (r *mikrotikRepository) List(ctx context.Context, page, pageSize int, search string) ([]entity.Mikrotik, int64, error) {
	var mks []entity.Mikrotik
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Mikrotik{})

	if search != "" {
		query = query.Where("name ILIKE ? OR location ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&mks).Error

	return mks, total, err
}

func (r *mikrotikRepository) Update(ctx context.Context, mk *entity.Mikrotik) error {
	return r.db.WithContext(ctx).Save(mk).Error
}

func (r *mikrotikRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Mikrotik{}, "id = ?", id).Error
}

func (r *mikrotikRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&entity.Mikrotik{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *mikrotikRepository) UpdateLastSync(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Mikrotik{}).
		Where("id = ?", id).
		Update("last_sync", &now).Error
}

// GetActiveMikrotik retrieves the currently active mikrotik
func (r *mikrotikRepository) GetActiveMikrotik(ctx context.Context) (*entity.Mikrotik, error) {
	var mk entity.Mikrotik
	err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&mk).Error
	return &mk, err
}

// SetActive sets a mikrotik as active or inactive
func (r *mikrotikRepository) SetActive(ctx context.Context, id string, active bool) error {
	return r.db.WithContext(ctx).Model(&entity.Mikrotik{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}

// DeactivateAll sets all mikrotiks as inactive
func (r *mikrotikRepository) DeactivateAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&entity.Mikrotik{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error
}
