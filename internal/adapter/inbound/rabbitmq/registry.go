package rabbitmq_inbound_adapter

import (
	"MikrOps/internal/domain"
	inbound_port "MikrOps/internal/port/inbound"
	outbound_port "MikrOps/internal/port/outbound"
)

type adapter struct {
	domain                domain.Domain
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewAdapter(
	domain domain.Domain,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.MessagePort {
	return &adapter{
		domain:                domain,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (a *adapter) Client() inbound_port.ClientMessagePort {
	return NewClientAdapter(a.domain)
}

func (a *adapter) Provisioning() inbound_port.ProvisioningMessagePort {
	return NewProvisioningAdapter(a.domain, a.mikrotikClientFactory)
}
