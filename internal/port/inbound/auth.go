package inbound_port

type AuthHttpPort interface {
	Register(a any)
	Login(a any)
}
