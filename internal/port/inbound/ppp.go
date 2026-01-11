package inbound_port

import "prabogo/internal/model"

type PPPPort interface {
	CreateSecret(a any) error
	GetSecret(a any) error
	UpdateSecret(a any) error
	DeleteSecret(a any) error
	ListSecrets(a any) error

	CreateProfile(a any) error
	GetProfile(a any) error
	UpdateProfile(a any) error
	DeleteProfile(a any) error
	ListProfiles(a any) error
}

type PPPDomain interface {
	// Secret
	CreateSecret(ctx any, input model.PPPSecretInput) (*model.PPPSecret, error)
	GetSecret(ctx any, id string) (*model.PPPSecret, error)
	UpdateSecret(ctx any, id string, input model.PPPSecretUpdateInput) (*model.PPPSecret, error)
	DeleteSecret(ctx any, id string) error
	ListSecrets(ctx any) ([]model.PPPSecret, error)

	// Profile
	CreateProfile(ctx any, input model.PPPProfileInput) (*model.PPPProfile, error)
	GetProfile(ctx any, id string) (*model.PPPProfile, error)
	UpdateProfile(ctx any, id string, input model.PPPProfileInput) (*model.PPPProfile, error)
	DeleteProfile(ctx any, id string) error
	ListProfiles(ctx any) ([]model.PPPProfile, error)
}
