package log

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	outbound_port "MikrOps/internal/port/outbound"
)

type LogDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
	cachePort             outbound_port.CachePort
}

func NewLogDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
	cachePort outbound_port.CachePort,
) *LogDomain {
	return &LogDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
		cachePort:             cachePort,
	}
}

func (d *LogDomain) getActiveClient(ctx context.Context) (outbound_port.MikrotikClientPort, error) {
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}
	return client, nil
}
