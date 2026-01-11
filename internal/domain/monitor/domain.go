package monitor

import (
	"fmt"

	"github.com/palantir/stacktrace"

	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type monitorDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewMonitorDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.MonitorDomain {
	return &monitorDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *monitorDomain) StreamTraffic(ctx any, interfaceName string) (<-chan map[string]interface{}, func(), error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, nil, fmt.Errorf("no active mikrotik found")
	}

	// Create client - we need a dedicated connection for streaming
	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}

	// Start listening
	replyChan, err := client.ListenArgs("/interface/monitor-traffic", map[string]string{
		"interface": interfaceName,
		"once":      "false", // continuous
	})
	if err != nil {
		client.Close()
		return nil, nil, stacktrace.Propagate(err, "failed to start monitor-traffic")
	}

	// Convert channel
	outChan := make(chan map[string]interface{})

	// Cancellation/Cleanup
	cancel := func() {
		client.Close() // This should break the Listen loop if library supports it or strictly close conn
		close(outChan) // We close our output channel
	}

	go func() {
		defer client.Close() // Ensure closed on exit
		for re := range replyChan {
			// re is *proto.Sentence
			if len(re.Map) > 0 {
				outChan <- map[string]interface{}{
					"rx-bits-per-second": re.Map["rx-bits-per-second"],
					"tx-bits-per-second": re.Map["tx-bits-per-second"],
				}
			}
		}
		// If replyChan closes, we are done
	}()

	return outChan, cancel, nil
}
