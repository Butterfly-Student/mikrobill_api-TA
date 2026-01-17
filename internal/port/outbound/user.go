package outbound_port

import (
	"context"

	"MikrOps/internal/model"

	"github.com/google/uuid"
)

//go:generate mockgen -source=user.go -destination=./../../../tests/mocks/port/mock_user.go

type UserDatabasePort interface {
	// CreateUser creates a new user in the database
	CreateUser(ctx context.Context, user *model.User) error

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// ListUsers retrieves users with pagination, optionally filtered by tenant
	ListUsers(ctx context.Context, tenantID *uuid.UUID, limit, offset int) ([]model.User, int64, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *model.User) error

	// DeleteUser soft deletes a user by ID
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// AssignUserRole assigns a role to a user
	AssignUserRole(ctx context.Context, userID, roleID uuid.UUID) error

	// GetUsersByTenantID retrieves all users for a specific tenant
	GetUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]model.User, error)
}
