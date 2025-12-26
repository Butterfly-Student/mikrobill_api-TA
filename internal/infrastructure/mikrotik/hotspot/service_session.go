package hotspot

import "mikrobill/internal/infrastructure/mikrotik/model"

// ========== ACTIVE SESSION MANAGEMENT ==========

// GetActiveUsers mendapatkan user yang sedang aktif
func (s *Service) GetActiveUsers() (*model.UserResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/active/print")
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	activeUsers := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		activeUsers[i] = re.Map
	}

	return &model.UserResponse{
		Message: "success",
		Data:    activeUsers,
	}, nil
}

// RemoveActiveUser disconnect user yang sedang aktif
func (s *Service) RemoveActiveUser(activeID string) (*model.UserResponse, error) {
	_, err := s.client.Run("/ip/hotspot/active/remove", "=.id="+activeID)
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

// ========== HOST MANAGEMENT ==========

// GetHosts mendapatkan daftar host
func (s *Service) GetHosts() (*model.UserResponse, error) {
	reply, err := s.client.Run("/ip/hotspot/host/print")
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	hosts := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		hosts[i] = re.Map
	}

	return &model.UserResponse{
		Message: "success",
		Data:    hosts,
	}, nil
}

// RemoveHost menghapus host
func (s *Service) RemoveHost(hostID string) (*model.UserResponse, error) {
	_, err := s.client.Run("/ip/hotspot/host/remove", "=.id="+hostID)
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
