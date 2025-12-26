package ppp

import (
	"fmt"

	"mikrobill/internal/infrastructure/mikrotik/model"
)

// ========== PPP SECRET MANAGEMENT ==========

// AddSecret menambahkan PPP secret baru
func (s *Service) AddSecret(config model.PPPSecretRequest) (*model.PPPSecretResponse, error) {
	// Prepare secret data
	args := []string{
		"=name=" + config.Name,
		"=password=" + config.Password,
		"=disabled=" + boolToYesNo(config.Disabled),
	}

	if config.Service != "" {
		args = append(args, "=service="+config.Service)
	}
	if config.Profile != "" {
		args = append(args, "=profile="+config.Profile)
	}
	if config.LocalAddress != "" {
		args = append(args, "=local-address="+config.LocalAddress)
	}
	if config.RemoteAddress != "" {
		args = append(args, "=remote-address="+config.RemoteAddress)
	}
	if config.CallerID != "" {
		args = append(args, "=caller-id="+config.CallerID)
	}
	if config.Routes != "" {
		args = append(args, "=routes="+config.Routes)
	}
	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}
	if config.LimitBytesIn != "" {
		args = append(args, "=limit-bytes-in="+config.LimitBytesIn)
	}
	if config.LimitBytesOut != "" {
		args = append(args, "=limit-bytes-out="+config.LimitBytesOut)
	}

	// Add secret
	sentence := append([]string{"/ppp/secret/add"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	// Get created secret
	secretID := reply.Done.Map["ret"]
	if secretID != "" {
		secret, err := s.GetSecret(config.Name)
		if err == nil {
			return secret, nil
		}
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data: model.SimpleSuccessData{
			Created: true,
			Name:    config.Name,
		},
	}, nil
}

// GetSecret mendapatkan detail PPP secret
func (s *Service) GetSecret(name string) (*model.PPPSecretResponse, error) {
	reply, err := s.client.Run("/ppp/secret/print", "?name="+name)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	if len(reply.Re) == 0 {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: "PPP secret not found"},
		}, fmt.Errorf("ppp secret not found")
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    reply.Re[0].Map,
	}, nil
}

// GetAllSecrets mendapatkan semua PPP secrets
func (s *Service) GetAllSecrets() (*model.PPPSecretResponse, error) {
	reply, err := s.client.Run("/ppp/secret/print")
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	secrets := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		secrets[i] = re.Map
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    secrets,
	}, nil
}

// ListSecrets mendapatkan secrets dengan filter
func (s *Service) ListSecrets(profile, service string) (*model.PPPSecretResponse, error) {
	args := []string{}

	if profile != "" {
		args = append(args, "?profile="+profile)
	}
	if service != "" {
		if len(args) > 0 {
			args = append(args, "?service="+service)
		} else {
			args = append(args, "?service="+service)
		}
	}

	sentence := append([]string{"/ppp/secret/print"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	secrets := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		secrets[i] = re.Map
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    secrets,
	}, nil
}

// UpdateSecret mengupdate PPP secret
func (s *Service) UpdateSecret(secretID string, updates model.PPPSecretUpdateRequest) (*model.PPPSecretResponse, error) {
	args := []string{"=.id=" + secretID}

	if updates.Password != "" {
		args = append(args, "=password="+updates.Password)
	}
	if updates.Service != "" {
		args = append(args, "=service="+updates.Service)
	}
	if updates.Profile != "" {
		args = append(args, "=profile="+updates.Profile)
	}
	if updates.LocalAddress != "" {
		args = append(args, "=local-address="+updates.LocalAddress)
	}
	if updates.RemoteAddress != "" {
		args = append(args, "=remote-address="+updates.RemoteAddress)
	}
	if updates.CallerID != "" {
		args = append(args, "=caller-id="+updates.CallerID)
	}
	if updates.Routes != "" {
		args = append(args, "=routes="+updates.Routes)
	}
	if updates.Comment != "" {
		args = append(args, "=comment="+updates.Comment)
	}
	if updates.LimitBytesIn != "" {
		args = append(args, "=limit-bytes-in="+updates.LimitBytesIn)
	}
	if updates.LimitBytesOut != "" {
		args = append(args, "=limit-bytes-out="+updates.LimitBytesOut)
	}
	if updates.Disabled != nil {
		args = append(args, "=disabled="+boolToYesNo(*updates.Disabled))
	}

	sentence := append([]string{"/ppp/secret/set"}, args...)
	_, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DeleteSecret menghapus PPP secret
func (s *Service) DeleteSecret(secretID string) (*model.PPPSecretResponse, error) {
	_, err := s.client.Run("/ppp/secret/remove", "=.id="+secretID)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Deleted: true},
	}, nil
}

// EnableSecret mengaktifkan PPP secret
func (s *Service) EnableSecret(secretID string) (*model.PPPSecretResponse, error) {
	_, err := s.client.Run("/ppp/secret/enable", "=.id="+secretID)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// DisableSecret menonaktifkan PPP secret
func (s *Service) DisableSecret(secretID string) (*model.PPPSecretResponse, error) {
	_, err := s.client.Run("/ppp/secret/disable", "=.id="+secretID)
	if err != nil {
		return &model.PPPSecretResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.PPPSecretResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Updated: true},
	}, nil
}

// boolToYesNo converts bool to "yes" or "no"
func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
