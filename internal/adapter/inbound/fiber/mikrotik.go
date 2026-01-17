package fiber_inbound_adapter

import (
	"MikrOps/internal/domain"
	inbound_port "MikrOps/internal/port/inbound"
)

type mikrotikAdapter struct {
	domain domain.Domain
}

func NewMikrotikAdapter(
	domain domain.Domain,
) inbound_port.MikrotikHttpPort {
	return &mikrotikAdapter{
		domain: domain,
	}
}

func (h *mikrotikAdapter) Create(a any) error            { return nil }
func (h *mikrotikAdapter) GetByID(a any) error           { return nil }
func (h *mikrotikAdapter) List(a any) error              { return nil }
func (h *mikrotikAdapter) Update(a any) error            { return nil }
func (h *mikrotikAdapter) Delete(a any) error            { return nil }
func (h *mikrotikAdapter) UpdateStatus(a any) error      { return nil }
func (h *mikrotikAdapter) GetActiveMikrotik(a any) error { return nil }
func (h *mikrotikAdapter) SetActive(a any) error         { return nil }

