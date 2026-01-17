package inbound_port

import (
	"context"
	"MikrOps/internal/model"
)

type MonitorPort interface {
	StreamTraffic(a any) error
	PingCustomer(a any) error
	StreamPing(a any) error
}

type MonitorDomain interface {
	StreamTraffic(ctx context.Context, customerID string) (<-chan model.CustomerTrafficData, error)
	PingCustomer(ctx context.Context, customerID string) (map[string]interface{}, error)
	StreamPing(ctx context.Context, customerID string) (<-chan model.PingResponse, error)
}

