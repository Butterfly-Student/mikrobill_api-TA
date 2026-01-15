package outbound_port

//go:generate mockgen -source=auth.go -destination=./../../../tests/mocks/port/mock_auth.go

import (
	"context"
	"prabogo/internal/model"

	"github.com/google/uuid"
)

type AuthDatabasePort interface {
	SaveUser(ctx context.Context, user model.User) error
	FindUserByFilter(ctx context.Context, filter model.UserFilter, lock bool) ([]model.User, error)
	FindRoleByName(ctx context.Context, name string) (*model.Role, error)
	FindRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	// Add other methods as needed
}
