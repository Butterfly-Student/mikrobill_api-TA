package outbound_port

//go:generate mockgen -source=profile.go -destination=./../../../tests/mocks/port/mock_profile.go

import (
	"context"
	"prabogo/internal/model"

	"github.com/google/uuid"
)

type ProfileDatabasePort interface {
	// CreateProfile inserts a new profile to mikrotik_profiles table
	CreateProfile(ctx context.Context, input model.ProfileInput, mikrotikID uuid.UUID) (*model.Profile, error)

	// CreateProfilePPPoE inserts PPPoE settings to mikrotik_profile_pppoe table
	CreateProfilePPPoE(ctx context.Context, profileID uuid.UUID, input model.ProfileInput) error

	// UpdateMikrotikObjectID updates the mikrotik_object_id field
	UpdateMikrotikObjectID(ctx context.Context, profileID uuid.UUID, objectID string) error

	// GetByID retrieves a profile with PPPoE settings by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProfileWithPPPoE, error)

	// GetByMikrotikID retrieves a profile by mikrotik_id
	GetByMikrotikID(ctx context.Context, mikrotikID, profileID uuid.UUID) (*model.ProfileWithPPPoE, error)

	// List retrieves all profiles for a MikroTik
	List(ctx context.Context, mikrotikID uuid.UUID) ([]model.ProfileWithPPPoE, error)

	// Update updates profile details
	Update(ctx context.Context, id uuid.UUID, input model.ProfileInput) error

	// Delete removes a profile
	Delete(ctx context.Context, id uuid.UUID) error
}
