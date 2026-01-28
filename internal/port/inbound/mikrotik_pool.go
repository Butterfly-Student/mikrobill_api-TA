package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type MikrotikPoolPort interface {
	MikrotikCreatePool(a any) error
	MikrotikGetPool(a any) error
	MikrotikUpdatePool(a any) error
	MikrotikDeletePool(a any) error
	MikrotikListPools(a any) error
}

type MikrotikPoolDomain interface {
	MikrotikCreatePool(ctx context.Context, input model.IPPoolInput) (*model.IPPool, error)
	MikrotikGetPool(ctx context.Context, id string) (*model.IPPool, error)
	MikrotikUpdatePool(ctx context.Context, id string, input model.IPPoolUpdateInput) (*model.IPPool, error)
	MikrotikDeletePool(ctx context.Context, id string) error
	MikrotikListPools(ctx context.Context) ([]model.IPPool, error)
}
