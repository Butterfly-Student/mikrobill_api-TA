package postgres_outbound_adapter

import (
	"context"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"
)

type userAdapter struct {
	db *gorm.DB
}

func NewUserAdapter(db *gorm.DB) outbound_port.UserDatabasePort {
	return &userAdapter{db: db}
}

// CreateUser creates a new user in the database
func (a *userAdapter) CreateUser(ctx context.Context, user *model.User) error {
	if err := a.db.WithContext(ctx).Create(user).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create user")
	}
	return nil
}

// GetUserByID retrieves a user by their ID
func (a *userAdapter) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := a.db.WithContext(ctx).
		Preload("Tenant").
		Preload("Role").
		Where("id = ?", id.String()).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get user")
	}
	return &user, nil
}

// ListUsers retrieves users with pagination, optionally filtered by tenant
func (a *userAdapter) ListUsers(ctx context.Context, tenantID *uuid.UUID, limit, offset int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := a.db.WithContext(ctx).Model(&model.User{})

	// Filter by tenant if specified
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, stacktrace.Propagate(err, "failed to count users")
	}

	// Get paginated results
	if err := query.
		Preload("Tenant").
		Preload("Role").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, stacktrace.Propagate(err, "failed to list users")
	}

	return users, total, nil
}

// UpdateUser updates an existing user
func (a *userAdapter) UpdateUser(ctx context.Context, user *model.User) error {
	if err := a.db.WithContext(ctx).Save(user).Error; err != nil {
		return stacktrace.Propagate(err, "failed to update user")
	}
	return nil
}

// DeleteUser soft deletes a user by ID
func (a *userAdapter) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := a.db.WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&model.User{}).Error; err != nil {
		return stacktrace.Propagate(err, "failed to delete user")
	}
	return nil
}

// AssignUserRole assigns a role to a user
func (a *userAdapter) AssignUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	roleIDStr := roleID.String()
	if err := a.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID.String()).
		Update("role_id", roleIDStr).Error; err != nil {
		return stacktrace.Propagate(err, "failed to assign role")
	}
	return nil
}

// GetUsersByTenantID retrieves all users for a specific tenant
func (a *userAdapter) GetUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]model.User, error) {
	var users []model.User
	if err := a.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		Find(&users).Error; err != nil {
		return nil, stacktrace.Propagate(err, "failed to get users by tenant")
	}
	return users, nil
}
