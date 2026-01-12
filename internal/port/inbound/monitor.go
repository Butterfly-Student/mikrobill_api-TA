package inbound_port

import "prabogo/internal/model"

type MonitorPort interface {
	StreamTraffic(a any) error
	PingCustomer(a any) error
	StreamPing(a any) error
}

type MonitorDomain interface {
	StreamTraffic(ctx any, customerID string) (<-chan model.CustomerTrafficData, error)
	PingCustomer(ctx any, customerID string) (map[string]interface{}, error)
	StreamPing(ctx any, customerID string) (<-chan model.PingResponse, error)
}
