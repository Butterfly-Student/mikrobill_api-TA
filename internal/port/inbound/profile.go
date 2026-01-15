package inbound_port

import (
	"context"
	"prabogo/internal/model"
)

type ProfilePort interface {
	CreateProfile(a any) error
	GetProfile(a any) error
	ListProfiles(a any) error
	UpdateProfile(a any) error
	DeleteProfile(a any) error
}

type ProfileDomain interface {
	CreateProfile(ctx context.Context, input model.ProfileInput) (*model.ProfileWithPPPoE, error)
	GetProfile(ctx context.Context, id string) (*model.ProfileWithPPPoE, error)
	ListProfiles(ctx context.Context) ([]model.ProfileWithPPPoE, error)
	UpdateProfile(ctx context.Context, id string, input model.ProfileInput) (*model.ProfileWithPPPoE, error)
	DeleteProfile(ctx context.Context, id string) error
}
