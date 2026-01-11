package outbound_port

import (
	"prabogo/internal/model"

	"github.com/google/uuid"
)

type ProfileDatabasePort interface {
	// CreateProfile inserts a new profile to mikrotik_profiles table
	CreateProfile(input model.ProfileInput, mikrotikID uuid.UUID) (*model.Profile, error)

	// CreateProfilePPPoE inserts PPPoE settings to mikrotik_profile_pppoe table
	CreateProfilePPPoE(profileID uuid.UUID, input model.ProfileInput) error

	// UpdateMikrotikObjectID updates the mikrotik_object_id field
	UpdateMikrotikObjectID(profileID uuid.UUID, objectID string) error

	// GetByID retrieves a profile with PPPoE settings by ID
	GetByID(id uuid.UUID) (*model.ProfileWithPPPoE, error)

	// GetByMikrotikID retrieves a profile by mikrotik_id
	GetByMikrotikID(mikrotikID uuid.UUID, profileID uuid.UUID) (*model.ProfileWithPPPoE, error)

	// List retrieves all profiles for a MikroTik
	List(mikrotikID uuid.UUID) ([]model.ProfileWithPPPoE, error)

	// Update updates profile details
	Update(id uuid.UUID, input model.ProfileInput) error

	// Delete removes a profile
	Delete(id uuid.UUID) error
}
