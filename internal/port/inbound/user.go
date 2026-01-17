package inbound_port

type UserHttpPort interface {
	CreateUser(a any)
	ListUsers(a any)
	GetUser(a any)
	UpdateUser(a any)
	DeleteUser(a any)
	AssignRole(a any)
	AssignToTenant(a any)
}
