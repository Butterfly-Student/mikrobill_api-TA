package queue

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	outbound_port "MikrOps/internal/port/outbound"
)

type QueueDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewQueueDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) *QueueDomain {
	return &QueueDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *QueueDomain) getActiveClient(ctx context.Context) (outbound_port.MikrotikClientPort, error) {
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
