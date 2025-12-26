package hotspot

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik/model"
	"strconv"
)

// ========== USER MANAGEMENT ==========

// AddUser menambahkan user baru
func (s *Service) AddUser(config model.UserRequest) (*model.UserResponse, error) {
	// Determine comment prefix
	commentPrefix := "up-"
	if config.Name == config.Password {
		commentPrefix = "vc-"
	}
	fullComment := commentPrefix + config.Comment

	// Calculate data limit
	dataLimitBytes, err := calculateDataLimit(config.DataLimit)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Prepare user data
	args := []string{
		"=name=" + config.Name,
		"=password=" + config.Password,
		"=profile=" + config.Profile,
		"=disabled=" + boolToYesNo(config.Disabled),
		"=comment=" + fullComment,
	}

	if config.Server != "" {
		args = append(args, "=server="+config.Server)
	}
	if config.MacAddress != "" {
		args = append(args, "=mac-address="+config.MacAddress)
	}
	if config.TimeLimit != "" {
		args = append(args, "=limit-uptime="+config.TimeLimit)
	}
	if dataLimitBytes > 0 {
		args = append(args, "=limit-bytes-total="+strconv.FormatInt(dataLimitBytes, 10))
	}

	// Add user
	sentences := append([]string{"/ip/hotspot/user/add"}, args...)
	reply, err := s.client.RunArgs(sentences)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Get created user
	userID := reply.Done.Map["ret"]
	if userID != "" {
		user, err := s.GetUser(config.Name)
		if err == nil {
			return user, nil
		}
	}

	return &model.UserResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			Name:    config.Name,
		},
	}, nil
}

// GetUser mendapatkan detail user
func (s *Service) GetUser(username string) (*model.UserResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/user/print", "?name="+username)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "User not found"},
		}, fmt.Errorf("user not found")
	}

	return &model.UserResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// GetAllUsers mendapatkan semua user
func (s *Service) GetAllUsers() (*model.UserResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/user/print")
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	users := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		users[i] = re.Map
	}

	return &model.UserResponse{
		Message: "success",
		Data:    users,
	}, nil
}

// UpdateUser mengupdate user
func (s *Service) UpdateUser(userID string, updates model.UserUpdateRequest) (*model.UserResponse, error) {
	args := []string{"=.id=" + userID}

	if updates.Name != "" {
		args = append(args, "=name="+updates.Name)
	}
	if updates.Password != "" {
		args = append(args, "=password="+updates.Password)
	}
	if updates.Profile != "" {
		args = append(args, "=profile="+updates.Profile)
	}
	if updates.MacAddress != "" {
		args = append(args, "=mac-address="+updates.MacAddress)
	}
	if updates.TimeLimit != "" {
		args = append(args, "=limit-uptime="+updates.TimeLimit)
	}
	if updates.DataLimit != "" {
		dataLimitBytes, err := calculateDataLimit(updates.DataLimit)
		if err == nil && dataLimitBytes > 0 {
			args = append(args, "=limit-bytes-total="+strconv.FormatInt(dataLimitBytes, 10))
		}
	}
	if updates.Comment != "" {
		commentPrefix := "up-"
		if updates.Name == updates.Password {
			commentPrefix = "vc-"
		}
		args = append(args, "=comment="+commentPrefix+updates.Comment)
	}
	if updates.Disabled != nil {
		args = append(args, "=disabled="+boolToYesNo(*updates.Disabled))
	}

	sentences := append([]string{"/ip/hotspot/user/SET"}, args...)
	_, err := s.client.RunArgs(sentences)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.UserResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteUser menghapus user
func (s *Service) DeleteUser(userID string) (*model.UserResponse, error) {
	_, err := s.client.Run("/ip/hotspot/user/remove", "=.id="+userID)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.UserResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// ListUsers mendapatkan user dengan filter
func (s *Service) GetUSerByComment(commentFilter string) (*model.UserResponse, error) {
	args := []string{}
	if commentFilter != "" {
		args = append(args, "?comment="+commentFilter)
	}

	sentences := append([]string{"/ip/hotspot/user/print"}, args...)
	reply, err := s.client.RunArgs(sentences)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	users := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		users[i] = re.Map
	}

	return &model.UserResponse{
		Message: "success",
		Data:    users,
	}, nil
}
