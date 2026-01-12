package monitor

import (
	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type monitorDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
	cachePort             outbound_port.CachePort
}

func NewMonitorDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
	cachePort outbound_port.CachePort,
) inbound_port.MonitorDomain {
	return &monitorDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
		cachePort:             cachePort,
	}
}
