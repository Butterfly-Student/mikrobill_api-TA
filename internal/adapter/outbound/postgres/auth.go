package postgres_outbound_adapter

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	contextutil "MikrOps/utils/context"
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
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("email = ?", email).First(&user).Error

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
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("username = ?", username).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find user by username")
	}
	return &user, nil
}

// Di file: internal/adapter/outbound/postgres/auth.go
func (a *authAdapter) FindUserByEmailOrUsername(ctx context.Context, identifier string) (*model.User, error) {
	var user model.User
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("email = ? OR username = ?", identifier, identifier).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil jika tidak ditemukan, bukan error
		}
		return nil, stacktrace.Propagate(err, "failed to find user")
	}

	// Load role if needed - with tenant filtering
	if user.RoleID != nil && *user.RoleID != "" {
		var role model.Role
		roleQuery := a.db.WithContext(ctx)

		if !isSuperAdmin {
			roleQuery = roleQuery.Where("tenant_id = ?", tenantID.String())
		}

		if err := roleQuery.First(&role, "id = ?", *user.RoleID).Error; err == nil {
			user.UserRole = model.UserRole(role.Name)
		}
	}

	return &user, nil
}

func (a *authAdapter) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("id = ?", id.String()).First(&user).Error

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
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("name = ?", name).First(&role).Error

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
	query := a.db.WithContext(ctx)

	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil && err != contextutil.ErrTenantIDNotFound {
		return nil, stacktrace.Propagate(err, "failed to get tenant id from context")
	}

	isSuperAdmin := contextutil.IsSuperAdmin(ctx)

	if !isSuperAdmin {
		query = query.Where("tenant_id = ?", tenantID.String())
	}

	err = query.Where("id = ?", id.String()).First(&role).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find role by id")
	}

	return &role, nil
}
