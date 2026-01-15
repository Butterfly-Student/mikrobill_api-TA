package inbound_port

import (
	"context"
	"prabogo/internal/model"
)

type MikrotikPPPSecretPort interface {
	MikrotikCreateSecret(a any) error
	MikrotikGetSecret(a any) error
	MikrotikUpdateSecret(a any) error
	MikrotikDeleteSecret(a any) error
	MikrotikListSecrets(a any) error
}

type MikrotikPPPSecretDomain interface {
	// Secret
	MikrotikCreateSecret(ctx context.Context, input model.PPPSecretInput) (*model.PPPSecret, error)
	MikrotikGetSecret(ctx context.Context, id string) (*model.PPPSecret, error)
	MikrotikUpdateSecret(ctx context.Context, id string, input model.PPPSecretUpdateInput) (*model.PPPSecret, error)
	MikrotikDeleteSecret(ctx context.Context, id string) error
	MikrotikListSecrets(ctx context.Context) ([]model.PPPSecret, error)
}
