package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type MikrotikQueuePort interface {
	MikrotikCreateQueue(a any) error
	MikrotikGetQueue(a any) error
	MikrotikUpdateQueue(a any) error
	MikrotikDeleteQueue(a any) error
	MikrotikListQueues(a any) error
}

type MikrotikQueueDomain interface {
	MikrotikCreateQueue(ctx context.Context, input model.QueueSimpleInput) (*model.QueueSimple, error)
	MikrotikGetQueue(ctx context.Context, id string) (*model.QueueSimple, error)
	MikrotikUpdateQueue(ctx context.Context, id string, input model.QueueSimpleUpdateInput) (*model.QueueSimple, error)
	MikrotikDeleteQueue(ctx context.Context, id string) error
	MikrotikListQueues(ctx context.Context) ([]model.QueueSimple, error)
}
