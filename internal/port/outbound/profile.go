package outbound_port

//go:generate mockgen -source=profile.go -destination=./../../../tests/mocks/port/mock_profile.go

import (
	"context"
	"MikrOps/internal/model"

	"github.com/google/uuid"
)

type ProfileDatabasePort interface {
	// CreateProfile inserts a new profile to mikrotik_profiles table
	CreateProfile(ctx context.Context, input model.CreateProfileRequest, mikrotikID uuid.UUID) (*model.MikrotikProfile, error)

	// CreateProfilePPPoE inserts PPPoE settings to mikrotik_profile_pppoe table
	CreateProfilePPPoE(ctx context.Context, profileID uuid.UUID, input model.CreateProfileRequest) error

	// UpdateMikrotikObjectID updates the mikrotik_object_id field
	UpdateMikrotikObjectID(ctx context.Context, profileID uuid.UUID, objectID string) error

	// GetByID retrieves a profile with PPPoE settings by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.MikrotikProfile, error)

	// GetByMikrotikID retrieves a profile by mikrotik_id
	GetByMikrotikID(ctx context.Context, mikrotikID, profileID uuid.UUID) (*model.MikrotikProfile, error)

	// List retrieves all profiles for a MikroTik
	List(ctx context.Context, mikrotikID uuid.UUID) ([]model.MikrotikProfile, error)

	// Update updates profile details
	Update(ctx context.Context, id uuid.UUID, input model.CreateProfileRequest) error

	// Delete removes a profile
	Delete(ctx context.Context, id uuid.UUID) error
}

