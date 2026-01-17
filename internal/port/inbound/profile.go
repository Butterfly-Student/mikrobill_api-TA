package inbound_port

import (
	"context"
	"MikrOps/internal/model"
)

type ProfilePort interface {
	CreateProfile(a any) error
	GetProfile(a any) error
	ListProfiles(a any) error
	UpdateProfile(a any) error
	DeleteProfile(a any) error
}

type ProfileDomain interface {
	CreateProfile(ctx context.Context, input model.CreateProfileRequest) (*model.MikrotikProfile, error)
	GetProfile(ctx context.Context, id string) (*model.MikrotikProfile, error)
	ListProfiles(ctx context.Context) ([]model.MikrotikProfile, error)
	UpdateProfile(ctx context.Context, id string, input model.CreateProfileRequest) (*model.MikrotikProfile, error)
	DeleteProfile(ctx context.Context, id string) error
}

