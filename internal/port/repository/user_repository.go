package repository

import (
	"context"
	"mikrobill/internal/entity"
	"time"

	"gorm.io/gorm"
)

// ==================== USER REPOSITORY ====================

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByAPIToken(ctx context.Context, token string) (*entity.User, error)
	List(ctx context.Context, page, pageSize int, search string, status *entity.UserStatus, role *entity.UserRole) ([]entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int64) error
	UpdateLastLogin(ctx context.Context, id int64, ip string) error
	IncrementFailedLogin(ctx context.Context, id int64) error
	ResetFailedLogin(ctx context.Context, id int64) error
	LockAccount(ctx context.Context, id int64, until time.Time) error
	UnlockAccount(ctx context.Context, id int64) error
	UpdatePassword(ctx context.Context, id int64, encryptedPassword string) error
	UpdateAPIToken(ctx context.Context, id int64, token string, expiresAt time.Time) error
	GetByRole(ctx context.Context, role entity.UserRole) ([]entity.User, error)
	GetByRoleID(ctx context.Context, roleID int64) ([]entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateStatus(ctx context.Context, id int64, status entity.UserStatus) error
	UpdateTwoFactor(ctx context.Context, id int64, enabled bool, secret string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		First(&user, id).Error
	return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("username = ?", username).
		First(&user).Error
	return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	return &user, err
}

func (r *userRepository) GetByAPIToken(ctx context.Context, token string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("api_token = ? AND api_token_expires_at > ?", token, time.Now()).
		First(&user).Error
	return &user, err
}

func (r *userRepository) List(ctx context.Context, page, pageSize int, search string, status *entity.UserStatus, role *entity.UserRole) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{}).Preload("Role")

	if search != "" {
		query = query.Where(
			"username ILIKE ? OR fullname ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if role != nil {
		query = query.Where("user_role = ?", *role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id int64, ip string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login": &now,
			"last_ip":    ip,
		}).Error
}

func (r *userRepository) IncrementFailedLogin(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + 1")).Error
}

func (r *userRepository) ResetFailedLogin(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("failed_login_attempts", 0).Error
}

func (r *userRepository) LockAccount(ctx context.Context, id int64, until time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       entity.UserStatusLocked,
			"locked_until": &until,
		}).Error
}

func (r *userRepository) UnlockAccount(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       entity.UserStatusActive,
			"locked_until": nil,
		}).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int64, encryptedPassword string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"encrypted_password":    encryptedPassword,
			"password_changed_at":   &now,
			"force_password_change": false,
			"failed_login_attempts": 0,
		}).Error
}

func (r *userRepository) UpdateAPIToken(ctx context.Context, id int64, token string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"api_token":             token,
			"api_token_expires_at":  &expiresAt,
		}).Error
}

func (r *userRepository) GetByRole(ctx context.Context, role entity.UserRole) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_role = ?", role).
		Find(&users).Error
	return users, err
}

func (r *userRepository) GetByRoleID(ctx context.Context, roleID int64) ([]entity.User, error) {
	var users []entity.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("role_id = ?", roleID).
		Find(&users).Error
	return users, err
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) UpdateStatus(ctx context.Context, id int64, status entity.UserStatus) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *userRepository) UpdateTwoFactor(ctx context.Context, id int64, enabled bool, secret string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"two_factor_enabled": enabled,
			"two_factor_secret":  secret,
		}).Error
}
