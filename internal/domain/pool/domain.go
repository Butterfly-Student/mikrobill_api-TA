package pool

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	outbound_port "MikrOps/internal/port/outbound"
)

type PoolDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewPoolDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) *PoolDomain {
	return &PoolDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *PoolDomain) getActiveClient(ctx context.Context) (outbound_port.MikrotikClientPort, error) {
	// Get active mikrotik from DB
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Create client
	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}
	return client, nil
}
