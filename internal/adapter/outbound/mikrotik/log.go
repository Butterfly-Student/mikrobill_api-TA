package mikrotik_outbound_adapter

import (
	"context"

	"MikrOps/internal/model"
)

// MonitorLogs streams real-time logs from MikroTik using ListenArgsContext
func MonitorLogs(ctx context.Context, client *Client) (<-chan model.MikrotikLog, error) {
	reply, err := client.ListenArgsContext(ctx, []string{
		"/log/print",
		"=follow=yes",
	})
	if err != nil {
		if IsConnectionError(err) {
			// Try to reconnect once
			if recErr := client.Reconnect(); recErr == nil {
				reply, err = client.ListenArgsContext(ctx, []string{
					"/log/print",
					"=follow=yes",
				})
			}
		}
	}
	if err != nil {
		return nil, err
	}

	logChan := make(chan model.MikrotikLog, 10)

	go func() {
		defer close(logChan)
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

				logEntry := model.MikrotikLog{
					Time:    r.Map["time"],
					Topics:  r.Map["topics"],
					Message: r.Map["message"],
				}

				select {
				case logChan <- logEntry:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return logChan, nil
}
