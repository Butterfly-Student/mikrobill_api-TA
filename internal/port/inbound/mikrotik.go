package inbound_port

type MikrotikHttpPort interface {
	Create(a any) error
	GetByID(a any) error
	List(a any) error
	Update(a any) error
	Delete(a any) error
	UpdateStatus(a any) error
	GetActiveMikrotik(a any) error
	SetActive(a any) error
}
