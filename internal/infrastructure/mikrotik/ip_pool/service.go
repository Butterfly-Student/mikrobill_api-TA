package ip_pool

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik"
	"mikrobill/internal/infrastructure/mikrotik/model"
)

// Service adalah service untuk operasi MikroTik IP Pool
type Service struct {
	client *mikrotik.Client
}

// NewService membuat instance baru dari IP Pool Service
func NewService(client *mikrotik.Client) *Service {
	return &Service{
		client: client,
	}
}

// ========== IP POOL MANAGEMENT ==========

// AddIPPool menambahkan IP Pool baru
func (s *Service) AddIPPool(config model.IPPoolRequest) (*model.IPPoolResponse, error) {
	// Prepare IP Pool data
	args := []string{
		"=name=" + config.Name,
		"=ranges=" + config.Ranges,
	}

	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}
	if config.NextPool != "" {
		args = append(args, "=next-pool="+config.NextPool)
	}

	// Add IP Pool
	sentence := append([]string{"/ip/pool/add"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Get created pool
	poolID := reply.Done.Map["ret"]
	if poolID != "" {
		pool, err := s.GetIPPool(poolID)
		if err == nil {
			return pool, nil
		}
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			ID:      poolID,
		},
	}, nil
}

// GetIPPool mendapatkan detail IP Pool
func (s *Service) GetIPPool(poolID string) (*model.IPPoolResponse, error) {
	reply, err := s.client.Run("/ip/pool/print", "?.id="+poolID)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "IP Pool not found"},
		}, fmt.Errorf("ip pool not found")
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// GetIPPoolByName mendapatkan IP Pool berdasarkan nama
func (s *Service) GetIPPoolByName(name string) (*model.IPPoolResponse, error) {
	reply, err := s.client.Run("/ip/pool/print", "?name="+name)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "IP Pool not found"},
		}, fmt.Errorf("ip pool not found")
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// ListIPPools mendapatkan semua IP Pools
func (s *Service) ListIPPools() (*model.IPPoolResponse, error) {
	reply, err := s.client.Run("/ip/pool/print")
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	pools := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		pools[i] = re.Map
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    pools,
	}, nil
}

// UpdateIPPool mengupdate IP Pool
func (s *Service) UpdateIPPool(poolID string, updates model.IPPoolUpdateRequest) (*model.IPPoolResponse, error) {
	args := []string{"=.id=" + poolID}

	if updates.Name != "" {
		args = append(args, "=name="+updates.Name)
	}
	if updates.Ranges != "" {
		args = append(args, "=ranges="+updates.Ranges)
	}
	if updates.Comment != "" {
		args = append(args, "=comment="+updates.Comment)
	}
	if updates.NextPool != "" {
		args = append(args, "=next-pool="+updates.NextPool)
	}

	sentence := append([]string{"/ip/pool/set"}, args...)
	_, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteIPPool menghapus IP Pool
func (s *Service) DeleteIPPool(poolID string) (*model.IPPoolResponse, error) {
	_, err := s.client.Run("/ip/pool/remove", "=.id="+poolID)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// GetIPPoolUsed mendapatkan IP yang sudah digunakan dalam pool
func (s *Service) GetIPPoolUsed(poolName string) (*model.IPPoolResponse, error) {
	reply, err := s.client.Run("/ip/pool/used/print", "?pool="+poolName)
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	usedIPs := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		usedIPs[i] = re.Map
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    usedIPs,
	}, nil
}

// ListAllUsedIPs mendapatkan semua IP yang sedang digunakan dari semua pool
func (s *Service) ListAllUsedIPs() (*model.IPPoolResponse, error) {
	reply, err := s.client.Run("/ip/pool/used/print")
	if err != nil {
		return &model.IPPoolResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	usedIPs := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		usedIPs[i] = re.Map
	}

	return &model.IPPoolResponse{
		Message: "success",
		Data:    usedIPs,
	}, nil
}
