package mikrotik_outbound_adapter

import (
	"context"
	"MikrOps/internal/model"
)

func MonitorTraffic(
	ctx context.Context,
	client *Client,
	iface string,
) (<-chan model.InterfaceTraffic, error) {

	reply, err := client.ListenArgsContext(ctx, []string{
		"/interface/monitor-traffic",
		"=interface=" + iface,
	})
	if err != nil {
		if isConnectionError(err) {
			// Try to reconnect once
			if recErr := client.Reconnect(); recErr == nil {
				reply, err = client.ListenArgsContext(ctx, []string{
					"/interface/monitor-traffic",
					"=interface=" + iface,
				})
			}
		}
	}
	if err != nil {
		return nil, err
	}

	out := make(chan model.InterfaceTraffic)

	go func() {
		defer close(out)
		defer client.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-reply.Chan():
				if !ok {
					return
				}
				if r == nil || r.Map == nil {
					continue
				}

				out <- mapToInterfaceTraffic(r.Map)
			}
		}
	}()

	return out, nil
}

func mapToInterfaceTraffic(m map[string]string) model.InterfaceTraffic {
	return model.InterfaceTraffic{
		Name: m["name"],

		RxPacketsPerSecond:   m["rx-packets-per-second"],
		RxBitsPerSecond:      m["rx-bits-per-second"],
		FpRxPacketsPerSecond: m["fp-rx-packets-per-second"],
		FpRxBitsPerSecond:    m["fp-rx-bits-per-second"],

		RxDropsPerSecond:  m["rx-drops-per-second"],
		RxErrorsPerSecond: m["rx-errors-per-second"],

		TxPacketsPerSecond:   m["tx-packets-per-second"],
		TxBitsPerSecond:      m["tx-bits-per-second"],
		FpTxPacketsPerSecond: m["fp-tx-packets-per-second"],
		FpTxBitsPerSecond:    m["fp-tx-bits-per-second"],

		TxDropsPerSecond:      m["tx-drops-per-second"],
		TxQueueDropsPerSecond: m["tx-queue-drops-per-second"],
		TxErrorsPerSecond:     m["tx-errors-per-second"],

		Section: m[".section"],
	}
}

