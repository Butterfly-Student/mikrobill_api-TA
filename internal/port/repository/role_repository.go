package repository

import (
	"context"
	"mikrobill/internal/entity"

	"gorm.io/gorm"
)

// ==================== ROLE REPOSITORY ====================

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id int64) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context, page, pageSize int, search string, isActive *bool, isSystem *bool) ([]entity.Role, int64, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id int64) error
	GetActive(ctx context.Context) ([]entity.Role, error)
	GetSystemRoles(ctx context.Context) ([]entity.Role, error)
	UpdatePermissions(ctx context.Context, id int64, permissions []byte) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id int64) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	return &role, err
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *roleRepository) List(ctx context.Context, page, pageSize int, search string, isActive *bool, isSystem *bool) ([]entity.Role, int64, error) {
	var roles []entity.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Role{})

	if search != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if isSystem != nil {
		query = query.Where("is_system = ?", *isSystem)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id ASC").Find(&roles).Error
	return roles, total, err
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) GetActive(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) UpdatePermissions(ctx context.Context, id int64, permissions []byte) error {
	return r.db.WithContext(ctx).Model(&entity.Role{}).
		Where("id = ?", id).
		Update("permissions", permissions).Error
}

func (r *roleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Role{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}