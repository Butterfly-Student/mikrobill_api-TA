package outbound_port

//go:generate mockgen -source=auth.go -destination=./../../../tests/mocks/port/mock_auth.go

import (
	"context"
	"MikrOps/internal/model"

	"github.com/google/uuid"
)

type AuthDatabasePort interface {
	SaveUser(ctx context.Context, user model.User) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindRoleByName(ctx context.Context, name string) (*model.Role, error)
	FindRoleByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
}

