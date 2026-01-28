package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type MikrotikPPPInactivePort interface {
	MikrotikListInactive(a any) error
}

type MikrotikPPPInactiveDomain interface {
	MikrotikListInactive(ctx context.Context) ([]model.PPPSecret, error)
}
