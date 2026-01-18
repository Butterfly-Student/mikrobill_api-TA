package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type DirectMonitorDomain interface {
	StreamTrafficByInterface(ctx context.Context, interfaceName string) (<-chan model.CustomerTrafficData, error)
	PingHost(ctx context.Context, targetIP string) (map[string]interface{}, error)
	StreamPingHost(ctx context.Context, targetIP string) (<-chan model.PingResponse, error)
}

type DirectMonitorPort interface {
	StreamTrafficByInterface(ctx any)
	PingHost(ctx any)
	StreamPingHost(ctx any)
}
