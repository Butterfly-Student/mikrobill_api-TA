package ppp

import "mikrobill/internal/infrastructure/mikrotik/model"

// ========== PPP ACTIVE CONNECTIONS ==========

// GetActiveConnections mendapatkan semua koneksi PPP yang aktif
func (s *Service) GetActiveConnections() (*model.PPPActiveResponse, error) {
	reply, err := s.client.Run("/ppp/active/print")
	if err != nil {
		return &model.PPPActiveResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	actives := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		actives[i] = re.Map
	}

	return &model.PPPActiveResponse{
		Message: "success",
		Data:    actives,
	}, nil
}

// GetActiveConnectionByName mendapatkan koneksi aktif berdasarkan username
func (s *Service) GetActiveConnectionByName(name string) (*model.PPPActiveResponse, error) {
	reply, err := s.client.Run("/ppp/active/print", "?name="+name)
	if err != nil {
		return &model.PPPActiveResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	actives := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		actives[i] = re.Map
	}

	return &model.PPPActiveResponse{
		Message: "success",
		Data:    actives,
	}, nil
}

// DisconnectActiveConnection memutus koneksi PPP yang aktif
func (s *Service) DisconnectActiveConnection(activeID string) (*model.PPPActiveResponse, error) {
	_, err := s.client.Run("/ppp/active/remove", "=.id="+activeID)
	if err != nil {
		return &model.PPPActiveResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPActiveResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// DisconnectByUsername memutus semua koneksi dari username tertentu
func (s *Service) DisconnectByUsername(username string) (*model.PPPActiveResponse, error) {
	// Get active connections for this user
	activeResp, err := s.GetActiveConnectionByName(username)
	if err != nil {
		return activeResp, err
	}

	// Convert to slice of maps
	actives, ok := activeResp.Data.([]map[string]string)
	if !ok {
		return &model.PPPActiveResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "Invalid data format"},
		}, nil
	}

	// Disconnect each active connection
	disconnected := 0
	for _, active := range actives {
		if id, exists := active[".id"]; exists {
			_, err := s.client.Run("/ppp/active/remove", "=.id="+id)
			if err == nil {
				disconnected++
			}
		}
	}

	return &model.PPPActiveResponse{
		Message: "success",
		Data: map[string]interface{}{
			"disconnected": disconnected,
			"username":     username,
		},
	}, nil
}

// GetActiveConnectionsByService mendapatkan koneksi aktif berdasarkan service
func (s *Service) GetActiveConnectionsByService(service string) (*model.PPPActiveResponse, error) {
	reply, err := s.client.Run("/ppp/active/print", "?service="+service)
	if err != nil {
		return &model.PPPActiveResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	actives := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		actives[i] = re.Map
	}

	return &model.PPPActiveResponse{
		Message: "success",
		Data:    actives,
	}, nil
}

// GetActiveConnectionStats mendapatkan statistik koneksi aktif
func (s *Service) GetActiveConnectionStats() (map[string]interface{}, error) {
	reply, err := s.client.Run("/ppp/active/print")
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	total := len(reply.Re)

	for _, re := range reply.Re {
		service := re.Map["service"]
		if service != "" {
			stats[service]++
		}
	}

	return map[string]interface{}{
		"total":      total,
		"by_service": stats,
	}, nil
}
