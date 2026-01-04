package repository

import (
	"fmt"
	"log"
	"mikrobill/internal/entity"
	"time"

	"gorm.io/gorm"
)

// DatabaseProfileRepository implements entity.ProfileRepository
type DatabaseProfileRepository struct {
	db *gorm.DB
}

// NewDatabaseProfileRepository creates a new profile repository
func NewDatabaseProfileRepository(db *gorm.DB) *DatabaseProfileRepository {
	return &DatabaseProfileRepository{
		db: db,
	}
}

// CreateProfile creates a new profile with PPPoE details in a transaction
func (r *DatabaseProfileRepository) CreateProfile(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	log.Printf("[ProfileRepo] CreateProfile - Creating profile: %s (Type: %s)", profile.Name, profile.ProfileType)

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create base profile
		if err := tx.Create(profile).Error; err != nil {
			log.Printf("[ProfileRepo] CreateProfile - ERROR creating profile: %v", err)
			return fmt.Errorf("failed to create profile: %w", err)
		}

		// Create PPPoE details if provided
		if pppoeDetails != nil && profile.ProfileType == "pppoe" {
			pppoeDetails.ProfileID = profile.ID
			if err := tx.Create(pppoeDetails).Error; err != nil {
				log.Printf("[ProfileRepo] CreateProfile - ERROR creating PPPoE details: %v", err)
				return fmt.Errorf("failed to create pppoe details: %w", err)
			}
			log.Printf("[ProfileRepo] CreateProfile - PPPoE details created for profile %s", profile.ID)
		}

		log.Printf("[ProfileRepo] CreateProfile - SUCCESS: Created profile %s (ID: %s)", profile.Name, profile.ID)
		return nil
	})
}

// UpdateProfile updates an existing profile and its PPPoE details
func (r *DatabaseProfileRepository) UpdateProfile(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	log.Printf("[ProfileRepo] UpdateProfile - Updating profile: %s (ID: %s)", profile.Name, profile.ID)

	return r.db.Transaction(func(tx *gorm.DB) error {
		profile.UpdatedAt = time.Now()

		// Update base profile
		result := tx.Model(&entity.MikrotikProfile{}).
			Where("id = ?", profile.ID).
			Updates(profile)

		if result.Error != nil {
			log.Printf("[ProfileRepo] UpdateProfile - ERROR updating profile: %v", result.Error)
			return fmt.Errorf("failed to update profile: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			log.Printf("[ProfileRepo] UpdateProfile - Profile not found: %s", profile.ID)
			return fmt.Errorf("profile not found: %s", profile.ID)
		}

		// Update or create PPPoE details if provided
		if pppoeDetails != nil && profile.ProfileType == "pppoe" {
			pppoeDetails.ProfileID = profile.ID

			// Check if PPPoE details exist
			var existing entity.MikrotikProfilePPPoE
			err := tx.Where("profile_id = ?", profile.ID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// Create new PPPoE details
				if err := tx.Create(pppoeDetails).Error; err != nil {
					log.Printf("[ProfileRepo] UpdateProfile - ERROR creating PPPoE details: %v", err)
					return fmt.Errorf("failed to create pppoe details: %w", err)
				}
				log.Printf("[ProfileRepo] UpdateProfile - Created new PPPoE details for profile %s", profile.ID)
			} else if err != nil {
				return fmt.Errorf("failed to check pppoe details: %w", err)
			} else {
				// Update existing PPPoE details
				if err := tx.Model(&entity.MikrotikProfilePPPoE{}).
					Where("profile_id = ?", profile.ID).
					Updates(pppoeDetails).Error; err != nil {
					log.Printf("[ProfileRepo] UpdateProfile - ERROR updating PPPoE details: %v", err)
					return fmt.Errorf("failed to update pppoe details: %w", err)
				}
				log.Printf("[ProfileRepo] UpdateProfile - Updated PPPoE details for profile %s", profile.ID)
			}
		}

		log.Printf("[ProfileRepo] UpdateProfile - SUCCESS: Updated profile %s", profile.ID)
		return nil
	})
}

// DeleteProfile deletes a profile and its PPPoE details (cascade)
func (r *DatabaseProfileRepository) DeleteProfile(id string) error {
	log.Printf("[ProfileRepo] DeleteProfile - Deleting profile ID: %s", id)

	result := r.db.Where("id = ?", id).Delete(&entity.MikrotikProfile{})

	if result.Error != nil {
		log.Printf("[ProfileRepo] DeleteProfile - ERROR: %v", result.Error)
		return fmt.Errorf("failed to delete profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("[ProfileRepo] DeleteProfile - Profile not found: %s", id)
		return fmt.Errorf("profile not found: %s", id)
	}

	log.Printf("[ProfileRepo] DeleteProfile - SUCCESS: Deleted profile %s (rows affected: %d)", id, result.RowsAffected)
	return nil
}

// GetProfileByID retrieves a profile with PPPoE details by ID
func (r *DatabaseProfileRepository) GetProfileByID(id string) (*entity.ProfileWithPPPoE, error) {
	log.Printf("[ProfileRepo] GetProfileByID - Searching for profile ID: %s", id)

	var profile entity.MikrotikProfile
	err := r.db.Where("id = ?", id).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		log.Printf("[ProfileRepo] GetProfileByID - Profile not found: %s", id)
		return nil, fmt.Errorf("profile not found: %s", id)
	}
	if err != nil {
		log.Printf("[ProfileRepo] GetProfileByID - ERROR: %v", err)
		return nil, fmt.Errorf("failed to query profile: %w", err)
	}

	result := &entity.ProfileWithPPPoE{
		MikrotikProfile: profile,
	}

	// Load PPPoE details if profile type is pppoe
	if profile.ProfileType == "pppoe" {
		var pppoeDetails entity.MikrotikProfilePPPoE
		err := r.db.Where("profile_id = ?", id).First(&pppoeDetails).Error
		if err == nil {
			result.PPPoEDetails = &pppoeDetails
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("[ProfileRepo] GetProfileByID - WARNING: Failed to load PPPoE details: %v", err)
		}
	}

	log.Printf("[ProfileRepo] GetProfileByID - SUCCESS: Found profile %s (%s)", profile.Name, profile.ID)
	return result, nil
}

// ListProfiles returns paginated profiles with optional mikrotik_id filter
func (r *DatabaseProfileRepository) ListProfiles(mikrotikID *string, page, limit int) ([]*entity.ProfileWithPPPoE, int64, error) {
	log.Printf("[ProfileRepo] ListProfiles - Page: %d, Limit: %d, MikrotikID: %v", page, limit, mikrotikID)

	var profiles []entity.MikrotikProfile
	var total int64

	query := r.db.Model(&entity.MikrotikProfile{})

	// Apply mikrotik_id filter if provided
	if mikrotikID != nil && *mikrotikID != "" {
		query = query.Where("mikrotik_id = ?", *mikrotikID)
	}

	// Count total
	err := query.Count(&total).Error
	if err != nil {
		log.Printf("[ProfileRepo] ListProfiles - ERROR counting: %v", err)
		return nil, 0, fmt.Errorf("failed to count profiles: %w", err)
	}

	// Fetch profiles
	offset := (page - 1) * limit
	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		log.Printf("[ProfileRepo] ListProfiles - ERROR querying: %v", err)
		return nil, 0, fmt.Errorf("failed to query profiles: %w", err)
	}

	// Load PPPoE details for pppoe profiles
	result := make([]*entity.ProfileWithPPPoE, len(profiles))
	for i, profile := range profiles {
		result[i] = &entity.ProfileWithPPPoE{
			MikrotikProfile: profile,
		}

		if profile.ProfileType == "pppoe" {
			var pppoeDetails entity.MikrotikProfilePPPoE
			err := r.db.Where("profile_id = ?", profile.ID).First(&pppoeDetails).Error
			if err == nil {
				result[i].PPPoEDetails = &pppoeDetails
			}
		}
	}

	log.Printf("[ProfileRepo] ListProfiles - SUCCESS: Found %d profiles (total: %d)", len(profiles), total)
	return result, total, nil
}

// GetProfilesByMikrotikID retrieves all profiles for a specific MikroTik device
func (r *DatabaseProfileRepository) GetProfilesByMikrotikID(mikrotikID string) ([]*entity.ProfileWithPPPoE, error) {
	log.Printf("[ProfileRepo] GetProfilesByMikrotikID - Searching for profiles with MikrotikID: %s", mikrotikID)

	var profiles []entity.MikrotikProfile
	err := r.db.Where("mikrotik_id = ?", mikrotikID).
		Order("name").
		Find(&profiles).Error

	if err != nil {
		log.Printf("[ProfileRepo] GetProfilesByMikrotikID - ERROR: %v", err)
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}

	result := make([]*entity.ProfileWithPPPoE, len(profiles))
	for i, profile := range profiles {
		result[i] = &entity.ProfileWithPPPoE{
			MikrotikProfile: profile,
		}

		if profile.ProfileType == "pppoe" {
			var pppoeDetails entity.MikrotikProfilePPPoE
			err := r.db.Where("profile_id = ?", profile.ID).First(&pppoeDetails).Error
			if err == nil {
				result[i].PPPoEDetails = &pppoeDetails
			}
		}
	}

	log.Printf("[ProfileRepo] GetProfilesByMikrotikID - SUCCESS: Found %d profiles", len(profiles))
	return result, nil
}

// GetProfileByName retrieves a profile by mikrotik_id and name
func (r *DatabaseProfileRepository) GetProfileByName(mikrotikID string, name string) (*entity.ProfileWithPPPoE, error) {
	log.Printf("[ProfileRepo] GetProfileByName - Searching for profile: %s (MikrotikID: %s)", name, mikrotikID)

	var profile entity.MikrotikProfile
	err := r.db.Where("mikrotik_id = ? AND name = ?", mikrotikID, name).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		log.Printf("[ProfileRepo] GetProfileByName - Profile not found: %s", name)
		return nil, fmt.Errorf("profile not found: %s", name)
	}
	if err != nil {
		log.Printf("[ProfileRepo] GetProfileByName - ERROR: %v", err)
		return nil, fmt.Errorf("failed to query profile: %w", err)
	}

	result := &entity.ProfileWithPPPoE{
		MikrotikProfile: profile,
	}

	if profile.ProfileType == "pppoe" {
		var pppoeDetails entity.MikrotikProfilePPPoE
		err := r.db.Where("profile_id = ?", profile.ID).First(&pppoeDetails).Error
		if err == nil {
			result.PPPoEDetails = &pppoeDetails
		}
	}

	log.Printf("[ProfileRepo] GetProfileByName - SUCCESS: Found profile %s", profile.ID)
	return result, nil
}

// UpdateSyncStatus updates the last_sync timestamp for a profile
func (r *DatabaseProfileRepository) UpdateSyncStatus(id string, lastSync time.Time) error {
	log.Printf("[ProfileRepo] UpdateSyncStatus - Updating sync status for profile: %s", id)

	result := r.db.Model(&entity.MikrotikProfile{}).
		Where("id = ?", id).
		Update("last_sync", lastSync)

	if result.Error != nil {
		log.Printf("[ProfileRepo] UpdateSyncStatus - ERROR: %v", result.Error)
		return fmt.Errorf("failed to update sync status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("[ProfileRepo] UpdateSyncStatus - Profile not found: %s", id)
		return fmt.Errorf("profile not found: %s", id)
	}

	log.Printf("[ProfileRepo] UpdateSyncStatus - SUCCESS: Updated sync status for %s", id)
	return nil
}
