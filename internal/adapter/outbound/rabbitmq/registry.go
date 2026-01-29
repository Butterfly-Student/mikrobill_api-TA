package rabbitmq_outbound_adapter

import (
	outbound_port "MikrOps/internal/port/outbound"
)

type adapter struct {
	client       outbound_port.ClientMessagePort
	provisioning outbound_port.ProvisioningMessagePort
}

func NewAdapter() outbound_port.MessagePort {
	return &adapter{
		client:       NewClientAdapter(),
		provisioning: NewProvisioningAdapter(),
	}
}

func (s *adapter) Client() outbound_port.ClientMessagePort {
	return s.client
}

func (s *adapter) Provisioning() outbound_port.ProvisioningMessagePort {
	return s.provisioning
}
