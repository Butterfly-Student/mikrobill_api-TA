package context

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"MikrOps/internal/model"
)

// Context keys for storing tenant-related data
type contextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"

	// UserKey is the context key for authenticated user
	UserKey contextKey = "user"

	// IsSuperAdminKey is the context key for super admin flag
	IsSuperAdminKey contextKey = "is_superadmin"
)

var (
	// ErrTenantIDNotFound is returned when tenant ID is not in context
	ErrTenantIDNotFound = errors.New("tenant ID not found in context")

	// ErrUserNotFound is returned when user is not in context
	ErrUserNotFound = errors.New("user not found in context")

	// ErrInvalidTenantID is returned when tenant ID has invalid type
	ErrInvalidTenantID = errors.New("invalid tenant ID type in context")

	// ErrInvalidUser is returned when user has invalid type
	ErrInvalidUser = errors.New("invalid user type in context")
)

// GetTenantID retrieves tenant ID from context
// Returns error if not found or invalid type
func GetTenantID(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(TenantIDKey)
	if value == nil {
		return uuid.Nil, ErrTenantIDNotFound
	}

	tenantID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidTenantID
	}

	return tenantID, nil
}

// MustGetTenantID retrieves tenant ID from context
// Panics if not found (use only when tenant ID is guaranteed to exist)
func MustGetTenantID(ctx context.Context) uuid.UUID {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		panic(err)
	}
	return tenantID
}

// GetUser retrieves authenticated user from context
// Returns error if not found or invalid type
func GetUser(ctx context.Context) (*model.User, error) {
	value := ctx.Value(UserKey)
	if value == nil {
		return nil, ErrUserNotFound
	}

	user, ok := value.(*model.User)
	if !ok {
		return nil, ErrInvalidUser
	}

	return user, nil
}

// MustGetUser retrieves user from context
// Panics if not found (use only when user is guaranteed to exist)
func MustGetUser(ctx context.Context) *model.User {
	user, err := GetUser(ctx)
	if err != nil {
		panic(err)
	}
	return user
}

// IsSuperAdmin checks if current user is super admin
// Returns false if flag not found in context
func IsSuperAdmin(ctx context.Context) bool {
	value := ctx.Value(IsSuperAdminKey)
	if value == nil {
		return false
	}

	isSuperAdmin, ok := value.(bool)
	if !ok {
		return false
	}

	return isSuperAdmin
}

// SetTenantID sets tenant ID in context
func SetTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// SetUser sets user in context
func SetUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

// SetSuperAdmin sets super admin flag in context
func SetSuperAdmin(ctx context.Context, isSuperAdmin bool) context.Context {
	return context.WithValue(ctx, IsSuperAdminKey, isSuperAdmin)
}

// WithTenantContext creates a new context with tenant information
func WithTenantContext(ctx context.Context, tenantID uuid.UUID, user *model.User, isSuperAdmin bool) context.Context {
	ctx = SetTenantID(ctx, tenantID)
	ctx = SetUser(ctx, user)
	ctx = SetSuperAdmin(ctx, isSuperAdmin)
	return ctx
}

