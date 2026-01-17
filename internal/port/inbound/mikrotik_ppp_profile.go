package inbound_port

import (
	"context"
	"MikrOps/internal/model"
)

type MikrotikPPPProfilePort interface {
	MikrotikCreateProfile(a any) error
	MikrotikGetProfile(a any) error
	MikrotikUpdateProfile(a any) error
	MikrotikDeleteProfile(a any) error
	MikrotikListProfiles(a any) error
}

type MikrotikPPPProfileDomain interface {
	// Profile
	MikrotikCreateProfile(ctx context.Context, input model.PPPProfileInput) (*model.PPPProfile, error)
	MikrotikGetProfile(ctx context.Context, id string) (*model.PPPProfile, error)
	MikrotikUpdateProfile(ctx context.Context, id string, input model.PPPProfileInput) (*model.PPPProfile, error)
	MikrotikDeleteProfile(ctx context.Context, id string) error
	MikrotikListProfiles(ctx context.Context) ([]model.PPPProfile, error)
}

