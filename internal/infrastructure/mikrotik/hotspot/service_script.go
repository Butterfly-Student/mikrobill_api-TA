package hotspot

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik/model"
)

// ========== SYSTEM SCRIPTS ==========

// SystemScriptConfig adalah konfigurasi system script
type SystemScriptConfig struct {
	Name     string `json:"name" binding:"required"`
	Source   string `json:"source" binding:"required"`
	Comment  string `json:"comment,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

// GetSystemScripts mendapatkan system scripts
func (s *Service) GetSystemScripts(source, name string) ([]map[string]string, error) {
	args := []string{}
	if source != "" {
		args = append(args, "?source="+source)
	}
	if name != "" {
		args = append(args, "?name="+name)
	}

	sentence := append([]string{"/system/script/print"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return nil, err
	}

	scripts := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		scripts[i] = re.Map
	}

	return scripts, nil
}

// AddSystemScript menambahkan system script
func (s *Service) AddSystemScript(config SystemScriptConfig) (*model.UserResponse, error) {
	args := []string{
		"=name=" + config.Name,
		"=source=" + config.Source,
		"=disabled=" + boolToYesNo(config.Disabled),
	}

	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}

	sentence := append([]string{"/system/script/add"}, args...)
	_, err := s.client.RunArgs(sentence)
	if err != nil {
		return &model.UserResponse{
			Message: "error",
			Data:    model.ErrorData{Error: err.Error()},
		}, err
	}

	return &model.UserResponse{
		Message: "success",
		Data:    model.SimpleSuccessData{Created: true, Name: config.Name},
	}, nil
}

// ========== GENERIC REMOVE ==========

// RemoveByID menghapus item berdasarkan ID dengan target yang spesifik
func (s *Service) RemoveByID(where, id string) error {
	var menuPath string

	switch where {
	case "user_":
		menuPath = "/ip/hotspot/user/remove"
	case "profile_":
		menuPath = "/ip/hotspot/user/profile/remove"
	case "active_":
		menuPath = "/ip/hotspot/active/remove"
	case "host_":
		menuPath = "/ip/hotspot/host/remove"
	default:
		return fmt.Errorf("invalid remove target: %s", where)
	}

	_, err := s.client.Run(menuPath, "=.id="+id)
	return err
}
