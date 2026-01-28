package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type MikrotikLogPort interface {
	StreamLogs(a any) error
}

type MikrotikLogDomain interface {
	StreamLogs(ctx context.Context) (<-chan model.LogStreamData, error)
	StopLogStream()
}
