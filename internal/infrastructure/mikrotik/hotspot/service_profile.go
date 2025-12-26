package hotspot

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik/model"
	"strconv"
)

// ========== PROFILE MANAGEMENT ==========

// CreateProfile membuat profile baru
func (s *Service) CreateProfile(config model.ProfileRequest) (*model.ProfileResponse, error) {
	// Validate config
	if err := validateProfileConfig(config); err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Sanitize name
	config.Name = sanitizeName(config.Name)

	// Set defaults
	if config.Price == "" {
		config.Price = "0"
	}
	if config.SellingPrice == "" {
		config.SellingPrice = "0"
	}
	if config.ExpMode == "" {
		config.ExpMode = model.ExpireModeNone
	}

	// Generate on-login script
	onLoginScript := generateOnLoginScript(config)

	// Prepare profile data
	args := []string{
		"=name=" + config.Name,
		"=status-autorefresh=1m",
	}

	if config.AddressPool != "" {
		args = append(args, "=address-pool="+config.AddressPool)
	}
	if config.RateLimit != "" {
		args = append(args, "=rate-limit="+config.RateLimit)
	}
	if config.SharedUsers != nil {
		args = append(args, "=shared-users="+strconv.Itoa(*config.SharedUsers))
	}
	if config.ParentQueue != "" {
		args = append(args, "=parent-queue="+config.ParentQueue)
	}
	if onLoginScript != "" {
		args = append(args, "=on-login="+onLoginScript)
	}

	// Add profile to MikroTik
	sentences := append([]string{"/ip/hotspot/user/profile/add"}, args...)

	reply, err := s.client.RunArgs(sentences)
	if err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Check for trap (error)
	if len(reply.Re) > 0 {
		if trapMsg, ok := reply.Re[0].Map["message"]; ok {
			return &model.ProfileResponse{
				Message: "error",
				Data:    model.ErrorData{Error: trapMsg},
			}, fmt.Errorf("error: %s", trapMsg)
		}
	}

	// Get created profile
	profileID := reply.Done.Map["ret"]
	if profileID != "" {
		profile, err := s.GetProfile(config.Name)
		if err == nil {
			return profile, nil
		}
	}

	return &model.ProfileResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			Name:    config.Name,
		},
	}, nil
}

// GetProfile mendapatkan detail profile
func (s *Service) GetProfile(name string) (*model.ProfileResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/user/profile/print", "?name="+name)
	if err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "Profile not found"},
		}, fmt.Errorf("profile not found")
	}

	return &model.ProfileResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// ListProfiles mendapatkan semua profile
func (s *Service) ListProfiles() (*model.ProfileResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/user/profile/print")
	if err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	profiles := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		profiles[i] = re.Map
	}

	return &model.ProfileResponse{
		Message: "success",
		Data:    profiles,
	}, nil
}

// UpdateProfile mengupdate profile
func (s *Service) UpdateProfile(profileID string, updates model.ProfileUpdateRequest) (*model.ProfileResponse, error) {
	args := []string{"=.id=" + profileID}

	if updates.RateLimit != "" {
		args = append(args, "=rate-limit="+updates.RateLimit)
	}
	if updates.SharedUsers != nil {
		args = append(args, "=shared-users="+strconv.Itoa(*updates.SharedUsers))
	}
	if updates.AddressPool != "" {
		args = append(args, "=address-pool="+updates.AddressPool)
	}
	if updates.ParentQueue != "" {
		args = append(args, "=parent-queue="+updates.ParentQueue)
	}

	// Generate new on-login script if needed
	if updates.ExpMode != "" || updates.Price != "" || updates.LockUser != "" {
		config := model.ProfileRequest{
			Name:       updates.Name,
			ExpMode:    updates.ExpMode,
			Price:      updates.Price,
			LockUser:   updates.LockUser,
			LockServer: updates.LockServer,
			Validity:   updates.Validity,
		}
		onLoginScript := generateOnLoginScript(config)
		if onLoginScript != "" {
			args = append(args, "=on-login="+onLoginScript)
		}
	}

	// Add profile to MikroTik
	sentences := append([]string{"/ip/hotspot/user/profile/set"}, args...)

	_, err := s.client.RunArgs(sentences)
	if err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.ProfileResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteProfile menghapus profile
func (s *Service) DeleteProfile(profileID string) (*model.ProfileResponse, error) {
	_, err := s.client.Run("/ip/hotspot/user/profile/remove", "=.id="+profileID)
	if err != nil {
		return &model.ProfileResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.ProfileResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// ListPools mendapatkan semua IP pool
func (s *Service) ListPools() (*model.Response, error) {
	reply, err := s.client.Run("/ip/pool/print")
	if err != nil {
		return &model.Response{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	pools := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		pools[i] = re.Map
	}

	return &model.Response{
		Message: "success",
		Data:    pools,
	}, nil
}

// ListServers mendapatkan semua hotspot server
func (s *Service) ListServers() (*model.Response, error) {
	reply, err := s.client.Run("/ip/hotspot/print")
	if err != nil {
		return &model.Response{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	servers := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		servers[i] = re.Map
	}

	return &model.Response{
		Message: "success",
		Data:    servers,
	}, nil
}
