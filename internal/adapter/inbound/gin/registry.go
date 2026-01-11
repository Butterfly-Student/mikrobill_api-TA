package gin_inbound_adapter

import (
	"prabogo/internal/domain"
	inbound_port "prabogo/internal/port/inbound"
)

type adapter struct {
	domain domain.Domain
}

func NewAdapter(
	domain domain.Domain,
) inbound_port.HttpPort {
	return &adapter{
		domain: domain,
	}
}

func (s *adapter) Ping() inbound_port.PingHttpPort {
	return NewPingAdapter()
}

func (s *adapter) Middleware() inbound_port.MiddlewareHttpPort {
	return NewMiddlewareAdapter(s.domain)
}

func (s *adapter) Client() inbound_port.ClientHttpPort {
	return NewClientAdapter(s.domain)
}

func (s *adapter) Testing() inbound_port.TestingHttpPort {
	return NewTestingAdapter(s.domain)
}

func (s *adapter) Mikrotik() inbound_port.MikrotikHttpPort {
	return NewMikrotikAdapter(s.domain)
}

func (s *adapter) Auth() inbound_port.AuthHttpPort {
	return NewAuthAdapter(s.domain)
}

func (s *adapter) PPP() inbound_port.PPPPort {
	return NewPPPAdapter(s.domain)
}

func (s *adapter) Monitor() inbound_port.MonitorPort {
	return NewMonitorAdapter(s.domain)
}

func (s *adapter) Profile() inbound_port.ProfilePort {
	return NewProfileAdapter(s.domain)
}

func (s *adapter) Customer() inbound_port.CustomerPort {
	return NewCustomerAdapter(s.domain)
}

func (s *adapter) Callback() inbound_port.CallbackHttpPort {
	return NewCallbackAdapter(s.domain)
}
