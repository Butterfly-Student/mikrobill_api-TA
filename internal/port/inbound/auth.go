package inbound_port

type AuthHttpPort interface {
	Register(a any)
	Login(a any)
	Logout(a any)
	RefreshToken(a any)
	GetProfile(a any)
}
