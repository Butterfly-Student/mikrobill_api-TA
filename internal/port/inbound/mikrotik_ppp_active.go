package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type MikrotikPPPActivePort interface {
	MikrotikListActive(a any) error
	MikrotikGetActive(a any) error
}

type MikrotikPPPActiveDomain interface {
	MikrotikListActive(ctx context.Context) ([]model.PPPActive, error)
	MikrotikGetActive(ctx context.Context, id string) (*model.PPPActive, error)
}
