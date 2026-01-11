package inbound_port

import "prabogo/internal/model"

type ProfilePort interface {
	CreateProfile(a any) error
	GetProfile(a any) error
	ListProfiles(a any) error
	UpdateProfile(a any) error
	DeleteProfile(a any) error
}

type ProfileDomain interface {
	CreateProfile(ctx any, input model.ProfileInput) (*model.ProfileWithPPPoE, error)
	GetProfile(ctx any, id string) (*model.ProfileWithPPPoE, error)
	ListProfiles(ctx any) ([]model.ProfileWithPPPoE, error)
	UpdateProfile(ctx any, id string, input model.ProfileInput) (*model.ProfileWithPPPoE, error)
	DeleteProfile(ctx any, id string) error
}
