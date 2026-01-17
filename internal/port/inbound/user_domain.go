package inbound_port

import (
	"context"

	"MikrOps/internal/model"
)

type UserDomain interface {
	CreateUser(ctx context.Context, input model.CreateUserRequest, createdBy string) (*model.User, error)
	GetUserByID(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) (*model.User, error)
	ListUsers(ctx context.Context, tenantID *string, requestingUserID string, isSuperAdmin bool, limit, offset int) ([]model.User, int64, error)
	UpdateUser(ctx context.Context, userID string, input model.UpdateUserRequest, updatedBy string, isSuperAdmin bool) (*model.User, error)
	DeleteUser(ctx context.Context, userID, requestingUserID string, isSuperAdmin bool) error
	AssignRole(ctx context.Context, userID, roleID, requestingUserID string, isSuperAdmin bool) error
	AssignToTenant(ctx context.Context, userID, tenantID string, roleID *string, isPrimary bool, requestingUserID string) error
}
