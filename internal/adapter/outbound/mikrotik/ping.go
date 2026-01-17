package mikrotik_outbound_adapter

import (
	"context"
	"MikrOps/internal/model"
)

func (c *Client) StreamPing(
	ctx context.Context,
	address string,
	size string,
	interval string,
) (<-chan model.PingResponse, error) {

	args := []string{
		"/ping",
		"=address=" + address,
	}

	if size != "" {
		args = append(args, "=size="+size)
	}
	if interval != "" {
		args = append(args, "=interval="+interval)
	}

	reply, err := c.ListenArgsContext(ctx, args)
	if err != nil {
		if isConnectionError(err) {
			if recErr := c.Reconnect(); recErr == nil {
				reply, err = c.ListenArgsContext(ctx, args)
			}
		}
	}
	if err != nil {
		return nil, err
	}

	out := make(chan model.PingResponse)

	go func() {
		defer close(out)

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
				out <- mapToPingResponse(r.Map)
			}
		}
	}()

	return out, nil
}

func mapToPingResponse(m map[string]string) model.PingResponse {
	isSummary := false
	if _, hasSeq := m["seq"]; !hasSeq {
		if _, hasSent := m["sent"]; hasSent {
			isSummary = true
		}
	}

	return model.PingResponse{
		Seq:        m["seq"],
		Host:       m["host"],
		Size:       m["size"],
		TTL:        m["ttl"],
		Time:       m["time"],
		Status:     m["status"],
		Sent:       m["sent"],
		Received:   m["received"],
		PacketLoss: m["packet-loss"],
		AvgRtt:     m["avg-rtt"],
		MinRtt:     m["min-rtt"],
		MaxRtt:     m["max-rtt"],
		IsSummary:  isSummary,
	}
}

