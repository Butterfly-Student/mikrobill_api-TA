package hotspot

import "mikrobill/internal/infrastructure/mikrotik/model"

// ========== SCHEDULER MANAGEMENT ==========

// SchedulerConfig adalah konfigurasi scheduler
type SchedulerConfig struct {
	Name      string `json:"name" binding:"required"`
	StartTime string `json:"startTime" binding:"required"`
	Interval  string `json:"interval" binding:"required"`
	OnEvent   string `json:"onEvent" binding:"required"`
	Disabled  bool   `json:"disabled,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// GetSchedulers mendapatkan scheduler
func (s *Service) GetSchedulers(name string) ([]map[string]string, error) {
	args := []string{}
	if name != "" {
		args = append(args, "?name="+name)
	}

	sentence := append([]string{"/system/scheduler/print"}, args...)
	reply, err := s.client.RunArgs(sentence)
	if err != nil {
		return nil, err
	}

	schedulers := make([]map[string]string, len(reply.Re))
	for i, re := range reply.Re {
		schedulers[i] = re.Map
	}

	return schedulers, nil
}

// AddScheduler menambahkan scheduler
func (s *Service) AddScheduler(config SchedulerConfig) (*model.UserResponse, error) {
	args := []string{
		"=name=" + config.Name,
		"=start-time=" + config.StartTime,
		"=interval=" + config.Interval,
		"=on-event=" + config.OnEvent,
		"=disabled=" + boolToYesNo(config.Disabled),
	}

	if config.Comment != "" {
		args = append(args, "=comment="+config.Comment)
	}

	sentence := append([]string{"/system/scheduler/add"}, args...)
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

// UpdateScheduler mengupdate scheduler
func (s *Service) UpdateScheduler(schedulerID string, updates SchedulerConfig) (*model.UserResponse, error) {
	args := []string{"=.id=" + schedulerID}

	if updates.StartTime != "" {
		args = append(args, "=start-time="+updates.StartTime)
	}
	if updates.Interval != "" {
		args = append(args, "=interval="+updates.Interval)
	}
	if updates.OnEvent != "" {
		args = append(args, "=on-event="+updates.OnEvent)
	}
	if updates.Comment != "" {
		args = append(args, "=comment="+updates.Comment)
	}
	args = append(args, "=disabled="+boolToYesNo(updates.Disabled))

	sentence := append([]string{"/system/scheduler/set"}, args...)
	_, err := s.client.RunArgs(sentence)
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

// RemoveScheduler menghapus scheduler
func (s *Service) RemoveScheduler(schedulerID string) (*model.UserResponse, error) {
	_, err := s.client.Run("/system/scheduler/remove", "=.id="+schedulerID)
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
