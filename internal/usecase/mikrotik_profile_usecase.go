package usecase

import (
	"context"
	"fmt"
	"log"
	"mikrobill/internal/entity"
	"mikrobill/internal/infrastructure/mikrotik"
	"strings"
	"time"
)

// ProfileService handles business logic for profile management
type ProfileService struct {
	repo            entity.ProfileRepository
	mikrotikUseCase MikrotikUseCase
}

// NewProfileService creates a new profile service
func NewProfileService(repo entity.ProfileRepository, mikrotikUseCase MikrotikUseCase) *ProfileService {
	return &ProfileService{
		repo:            repo,
		mikrotikUseCase: mikrotikUseCase,
	}
}

// CreateProfileWithSync creates a profile in database and syncs to MikroTik
func (s *ProfileService) CreateProfileWithSync(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	log.Printf("[ProfileService] CreateProfileWithSync - Creating profile: %s", profile.Name)

	// Validate profile
	if err := s.validateProfile(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileService] CreateProfileWithSync - Validation failed: %v", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create in database
	if err := s.repo.CreateProfile(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileService] CreateProfileWithSync - DB creation failed: %v", err)
		return err
	}

	// Sync to MikroTik if enabled
	if profile.SyncWithMikrotik {
		if err := s.syncProfileToMikrotik(profile, pppoeDetails); err != nil {
			log.Printf("[ProfileService] CreateProfileWithSync - WARNING: Sync failed: %v", err)
			// Don't fail the entire operation if sync fails
			// Just log the error and continue
		} else {
			// Update sync status
			now := time.Now()
			s.repo.UpdateSyncStatus(profile.ID, now)
		}
	}

	log.Printf("[ProfileService] CreateProfileWithSync - SUCCESS: Profile %s created", profile.ID)
	return nil
}

// UpdateProfileWithSync updates a profile in database and syncs to MikroTik
func (s *ProfileService) UpdateProfileWithSync(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	log.Printf("[ProfileService] UpdateProfileWithSync - Updating profile: %s", profile.ID)

	// Validate profile
	if err := s.validateProfile(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileService] UpdateProfileWithSync - Validation failed: %v", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.repo.UpdateProfile(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileService] UpdateProfileWithSync - DB update failed: %v", err)
		return err
	}

	// Sync to MikroTik if enabled
	if profile.SyncWithMikrotik {
		if err := s.syncProfileToMikrotik(profile, pppoeDetails); err != nil {
			log.Printf("[ProfileService] UpdateProfileWithSync - WARNING: Sync failed: %v", err)
		} else {
			now := time.Now()
			s.repo.UpdateSyncStatus(profile.ID, now)
		}
	}

	log.Printf("[ProfileService] UpdateProfileWithSync - SUCCESS: Profile %s updated", profile.ID)
	return nil
}

// DeleteProfileWithSync deletes a profile from database and MikroTik
func (s *ProfileService) DeleteProfileWithSync(id string) error {
	log.Printf("[ProfileService] DeleteProfileWithSync - Deleting profile: %s", id)

	// Get profile first to check sync status
	profile, err := s.repo.GetProfileByID(id)
	if err != nil {
		log.Printf("[ProfileService] DeleteProfileWithSync - Failed to get profile: %v", err)
		return err
	}

	// Delete from MikroTik if sync is enabled
	if profile.SyncWithMikrotik {
		client, err := s.mikrotikUseCase.GetClientByID(context.Background(), profile.MikrotikID)
		if err != nil {
			log.Printf("[ProfileService] DeleteProfileWithSync - WARNING: Failed to get MikroTik client: %v", err)
		} else {
			defer client.Close()
			if err := client.DeletePPPoEProfile(profile.Name); err != nil {
				log.Printf("[ProfileService] DeleteProfileWithSync - WARNING: Failed to delete from MikroTik: %v", err)
				// Continue with DB deletion even if MikroTik deletion fails
			}
		}
	}

	// Delete from database
	if err := s.repo.DeleteProfile(id); err != nil {
		log.Printf("[ProfileService] DeleteProfileWithSync - DB deletion failed: %v", err)
		return err
	}

	log.Printf("[ProfileService] DeleteProfileWithSync - SUCCESS: Profile %s deleted", id)
	return nil
}

// GetProfile retrieves a profile by ID
func (s *ProfileService) GetProfile(id string) (*entity.ProfileWithPPPoE, error) {
	return s.repo.GetProfileByID(id)
}

// ListProfiles returns paginated profiles
func (s *ProfileService) ListProfiles(mikrotikID *string, page, limit int) ([]*entity.ProfileWithPPPoE, int64, error) {
	return s.repo.ListProfiles(mikrotikID, page, limit)
}

// SyncProfileToMikrotik pushes a profile from database to MikroTik
func (s *ProfileService) SyncProfileToMikrotik(id string) error {
	log.Printf("[ProfileService] SyncProfileToMikrotik - Syncing profile: %s", id)

	profile, err := s.repo.GetProfileByID(id)
	if err != nil {
		return err
	}

	if err := s.syncProfileToMikrotik(&profile.MikrotikProfile, profile.PPPoEDetails); err != nil {
		return err
	}

	// Update sync timestamp
	now := time.Now()
	return s.repo.UpdateSyncStatus(id, now)
}

// SyncProfileFromMikrotik pulls a profile from MikroTik and updates database
func (s *ProfileService) SyncProfileFromMikrotik(mikrotikID string, profileName string) error {
	log.Printf("[ProfileService] SyncProfileFromMikrotik - Syncing profile: %s from MikroTik", profileName)

	// Get profile from MikroTik
	client, err := s.mikrotikUseCase.GetClientByID(context.Background(), mikrotikID)
	if err != nil {
		return fmt.Errorf("failed to get mikrotik client: %w", err)
	}
	defer client.Close()

	mtProfile, err := client.GetPPPoEProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to get profile from mikrotik: %w", err)
	}

	// Check if profile exists in database
	existingProfile, err := s.repo.GetProfileByName(mikrotikID, profileName)

	if err != nil {
		// Profile doesn't exist in DB, create it
		profile, pppoeDetails := s.convertMikrotikProfileToDB(mikrotikID, mtProfile)
		if err := s.repo.CreateProfile(profile, pppoeDetails); err != nil {
			return fmt.Errorf("failed to create profile in db: %w", err)
		}
		log.Printf("[ProfileService] SyncProfileFromMikrotik - Created new profile from MikroTik: %s", profileName)
	} else {
		// Profile exists, update it
		profile, pppoeDetails := s.convertMikrotikProfileToDB(mikrotikID, mtProfile)
		profile.ID = existingProfile.ID
		if err := s.repo.UpdateProfile(profile, pppoeDetails); err != nil {
			return fmt.Errorf("failed to update profile in db: %w", err)
		}
		log.Printf("[ProfileService] SyncProfileFromMikrotik - Updated profile from MikroTik: %s", profileName)
	}

	return nil
}

// SyncAllProfiles syncs all profiles from MikroTik to database
func (s *ProfileService) SyncAllProfiles(mikrotikID string) error {
	log.Printf("[ProfileService] SyncAllProfiles - Syncing all profiles from MikroTik")

	// Get all profiles from MikroTik
	client, err := s.mikrotikUseCase.GetClientByID(context.Background(), mikrotikID)
	if err != nil {
		return fmt.Errorf("failed to get mikrotik client: %w", err)
	}
	defer client.Close()

	mtProfiles, err := client.GetPPPoEProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles from mikrotik: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, mtProfile := range mtProfiles {
		profileName := mtProfile["name"]
		if err := s.SyncProfileFromMikrotik(mikrotikID, profileName); err != nil {
			log.Printf("[ProfileService] SyncAllProfiles - ERROR syncing %s: %v", profileName, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("[ProfileService] SyncAllProfiles - Completed: %d success, %d errors", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("sync completed with %d errors", errorCount)
	}

	return nil
}

// syncProfileToMikrotik is a helper to sync a profile to MikroTik
func (s *ProfileService) syncProfileToMikrotik(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	if profile.ProfileType != "pppoe" {
		return fmt.Errorf("only pppoe profiles are supported for sync")
	}

	if pppoeDetails == nil {
		return fmt.Errorf("pppoe details required for pppoe profile")
	}

	params := mikrotik.PPPoEProfileParams{
		Name:          profile.Name,
		LocalAddress:  pppoeDetails.LocalAddress,
		RemoteAddress: "",
	}

	if pppoeDetails.RemoteAddress != nil {
		params.RemoteAddress = *pppoeDetails.RemoteAddress
	}
	if profile.RateLimitUp != nil && profile.RateLimitDown != nil {
		params.RateLimitUp = *profile.RateLimitUp
		params.RateLimitDown = *profile.RateLimitDown
	}
	if profile.IdleTimeout != nil {
		params.IdleTimeout = *profile.IdleTimeout
	}
	if profile.SessionTimeout != nil {
		params.SessionTimeout = *profile.SessionTimeout
	}
	if profile.KeepaliveTimeout != nil {
		params.KeepaliveTimeout = *profile.KeepaliveTimeout
	}
	params.OnlyOne = profile.OnlyOne
	if profile.DNSServer != nil {
		params.DNSServer = *profile.DNSServer
	}

	// Check if profile exists on MikroTik
	client, err := s.mikrotikUseCase.GetClientByID(context.Background(), profile.MikrotikID)
	if err != nil {
		return fmt.Errorf("failed to get mikrotik client: %w", err)
	}
	defer client.Close()

	profileID, err := client.FindPPPoEProfileID(profile.Name)
	if err != nil {
		return fmt.Errorf("failed to check profile existence: %w", err)
	}

	if profileID == "" {
		// Create new profile
		return client.CreatePPPoEProfile(params)
	} else {
		// Update existing profile
		return client.UpdatePPPoEProfile(params)
	}
}

// convertMikrotikProfileToDB converts MikroTik API response to domain models
func (s *ProfileService) convertMikrotikProfileToDB(mikrotikID string, mtProfile map[string]string) (*entity.MikrotikProfile, *entity.MikrotikProfilePPPoE) {
	profile := &entity.MikrotikProfile{
		MikrotikID:       mikrotikID,
		Name:             mtProfile["name"],
		ProfileType:      "pppoe",
		SyncWithMikrotik: true,
		IsActive:         true,
	}

	// Parse rate-limit
	if rateLimit, ok := mtProfile["rate-limit"]; ok && rateLimit != "" {
		parts := strings.Split(rateLimit, "/")
		if len(parts) == 2 {
			profile.RateLimitUp = &parts[0]
			profile.RateLimitDown = &parts[1]
		}
	}

	// Parse timeouts
	if idleTimeout, ok := mtProfile["idle-timeout"]; ok && idleTimeout != "" {
		profile.IdleTimeout = &idleTimeout
	}
	if sessionTimeout, ok := mtProfile["session-timeout"]; ok && sessionTimeout != "" {
		profile.SessionTimeout = &sessionTimeout
	}
	if keepaliveTimeout, ok := mtProfile["keepalive-timeout"]; ok && keepaliveTimeout != "" {
		profile.KeepaliveTimeout = &keepaliveTimeout
	}

	// Parse only-one
	if onlyOne, ok := mtProfile["only-one"]; ok {
		profile.OnlyOne = onlyOne == "yes" || onlyOne == "true"
	}

	// Parse DNS server
	if dnsServer, ok := mtProfile["dns-server"]; ok && dnsServer != "" {
		profile.DNSServer = &dnsServer
	}

	// PPPoE details
	pppoeDetails := &entity.MikrotikProfilePPPoE{
		LocalAddress: mtProfile["local-address"],
	}

	if remoteAddress, ok := mtProfile["remote-address"]; ok && remoteAddress != "" {
		pppoeDetails.RemoteAddress = &remoteAddress
	}

	return profile, pppoeDetails
}

// validateProfile validates profile data
func (s *ProfileService) validateProfile(profile *entity.MikrotikProfile, pppoeDetails *entity.MikrotikProfilePPPoE) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.ProfileType == "" {
		return fmt.Errorf("profile type is required")
	}

	if profile.ProfileType == "pppoe" {
		if pppoeDetails == nil {
			return fmt.Errorf("pppoe details required for pppoe profile")
		}
		if pppoeDetails.LocalAddress == "" {
			return fmt.Errorf("local address is required for pppoe profile")
		}
	}

	return nil
}
