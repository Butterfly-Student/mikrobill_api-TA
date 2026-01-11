package domain

import (
	"prabogo/internal/domain/auth"
	"prabogo/internal/domain/client"
	"prabogo/internal/domain/customer"
	"prabogo/internal/domain/mikrotik"
	"prabogo/internal/domain/monitor"
	"prabogo/internal/domain/ppp"
	"prabogo/internal/domain/profile"
	"prabogo/internal/domain/testing"
	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type Domain interface {
	Testing() testing.TestingDomain
	Client() client.ClientDomain
	Auth() auth.AuthDomain
	Mikrotik() mikrotik.MikrotikDomain
	PPP() inbound_port.PPPDomain
	Monitor() inbound_port.MonitorDomain
	Profile() inbound_port.ProfileDomain
	Customer() inbound_port.CustomerDomain
}

type domain struct {
	databasePort          outbound_port.DatabasePort
	messagePort           outbound_port.MessagePort
	cachePort             outbound_port.CachePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) Domain {
	return &domain{
		databasePort:          databasePort,
		messagePort:           messagePort,
		cachePort:             cachePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *domain) Client() client.ClientDomain {
	return client.NewClientDomain(d.databasePort, d.messagePort, d.cachePort)
}

func (d *domain) Auth() auth.AuthDomain {
	return auth.NewAuthDomain(d.databasePort)
}

func (d *domain) Testing() testing.TestingDomain {
	return testing.NewTestingDomain()
}

func (d *domain) Mikrotik() mikrotik.MikrotikDomain {
	return mikrotik.NewMikrotikDomain(d.databasePort)
}

func (d *domain) PPP() inbound_port.PPPDomain {
	return ppp.NewPPPDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) Monitor() inbound_port.MonitorDomain {
	return monitor.NewMonitorDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) Profile() inbound_port.ProfileDomain {
	return profile.NewProfileDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) Customer() inbound_port.CustomerDomain {
	return customer.NewCustomerDomain(d.databasePort, d.mikrotikClientFactory)
}
