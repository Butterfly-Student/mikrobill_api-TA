package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
)

const (
	tableUsers = "users"
	tableRoles = "roles"
)

type authAdapter struct {
	db *gorm.DB
}

func NewAuthAdapter(db *gorm.DB) outbound_port.AuthDatabasePort {
	return &authAdapter{db: db}
}

func (a *authAdapter) SaveUser(ctx context.Context, user model.User) error {
	if err := a.db.WithContext(ctx).Save(&user).Error; err != nil {
		return stacktrace.Propagate(err, "failed to save user")
	}
	return nil
}

func (a *authAdapter) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := a.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find user by email")
	}
	return &user, nil
}

func (a *authAdapter) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := a.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find user by username")
	}
	return &user, nil
}

func (a *authAdapter) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := a.db.WithContext(ctx).
		Where("id = ?", id.String()).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find user by id")
	}
	return &user, nil
}

func (a *authAdapter) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role

	err := a.db.WithContext(ctx).
		Where("name = ?", name).
		First(&role).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find role by name")
	}

	return &role, nil
}

func (a *authAdapter) FindRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role

	err := a.db.WithContext(ctx).
		Where("id = ?", id.String()).
		First(&role).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find role by id")
	}

	return &role, nil
}

