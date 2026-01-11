package outbound_port

import "prabogo/internal/model"

type AuthDatabasePort interface {
	SaveUser(user model.User) error
	FindUserByFilter(filter model.UserFilter, lock bool) ([]model.User, error)
	FindRoleByName(name string) (*model.Role, error)
	// Add other methods as needed
}
