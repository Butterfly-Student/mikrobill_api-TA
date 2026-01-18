package domain

import (
	"MikrOps/internal/domain/auth"
	"MikrOps/internal/domain/client"
	"MikrOps/internal/domain/customer"
	"MikrOps/internal/domain/mikrotik"
	"MikrOps/internal/domain/monitor"
	"MikrOps/internal/domain/ppp"
	"MikrOps/internal/domain/profile"
	"MikrOps/internal/domain/tenant"
	"MikrOps/internal/domain/testing"
	"MikrOps/internal/domain/user"
	inbound_port "MikrOps/internal/port/inbound"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/logger"
)

type Domain interface {
	Testing() testing.TestingDomain
	Client() client.ClientDomain
	Auth() auth.AuthDomain
	Mikrotik() mikrotik.MikrotikDomain
	MikrotikPPPSecret() inbound_port.MikrotikPPPSecretDomain
	MikrotikPPPProfile() inbound_port.MikrotikPPPProfileDomain
	Monitor() inbound_port.MonitorDomain
	DirectMonitor() inbound_port.DirectMonitorDomain
	Profile() inbound_port.ProfileDomain
	Customer() inbound_port.CustomerDomain
	Tenant() inbound_port.TenantDomain
	User() inbound_port.UserDomain
	Database() outbound_port.DatabasePort // For direct queries in middleware
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
	return auth.NewAuthDomain(d.databasePort, d.cachePort)
}

func (d *domain) Testing() testing.TestingDomain {
	return testing.NewTestingDomain()
}

func (d *domain) Mikrotik() mikrotik.MikrotikDomain {
	return mikrotik.NewMikrotikDomain(d.databasePort)
}

func (d *domain) MikrotikPPPSecret() inbound_port.MikrotikPPPSecretDomain {
	return ppp.NewPPPDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) MikrotikPPPProfile() inbound_port.MikrotikPPPProfileDomain {
	return ppp.NewPPPDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) Monitor() inbound_port.MonitorDomain {
	return monitor.NewMonitorDomain(d.databasePort, d.mikrotikClientFactory, d.cachePort)
}

func (d *domain) DirectMonitor() inbound_port.DirectMonitorDomain {
	return monitor.NewMonitorDomain(d.databasePort, d.mikrotikClientFactory, d.cachePort).(inbound_port.DirectMonitorDomain)
}

func (d *domain) Profile() inbound_port.ProfileDomain {
	return profile.NewProfileDomain(d.databasePort, d.mikrotikClientFactory)
}

func (d *domain) Customer() inbound_port.CustomerDomain {
	return customer.NewCustomerDomain(d.databasePort, d.mikrotikClientFactory, d.cachePort)
}

func (d *domain) Tenant() inbound_port.TenantDomain {
	return tenant.NewTenantDomain(d.databasePort, logger.GetLogger())
}

func (d *domain) User() inbound_port.UserDomain {
	return user.NewUserDomain(d.databasePort)
}

func (d *domain) Database() outbound_port.DatabasePort {
	return d.databasePort
}
