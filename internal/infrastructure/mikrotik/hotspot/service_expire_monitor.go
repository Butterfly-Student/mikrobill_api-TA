package hotspot

import (
	"fmt"
)

// ========== EXPIRE MONITOR ==========

// ExpireMonitorResponse adalah response untuk setup expire monitor
type ExpireMonitorResponse struct {
	Message string      `json:"message"`
	Action  string      `json:"action,omitempty"` // "created", "updated", "already_exists"
	Data    interface{} `json:"data,omitempty"`
}

// ExpireMonitorData adalah data expire monitor
type ExpireMonitorData struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Disabled bool   `json:"disabled"`
}

// SetupExpireMonitor setup expire monitor scheduler
func (s *Service) SetupExpireMonitor(script string) (*ExpireMonitorResponse, error) {
	if script == "" {
		return nil, fmt.Errorf("expire monitor script is required")
	}

	// Check if expire monitor exists
	reply, err := s.client.Run("/system/scheduler/print", "?name=Expire-Monitor")
	if err != nil {
		return nil, err
	}

	// Case 1: Monitor doesn't exist - create new
	if len(reply.Re) == 0 {
		addReply, err := s.client.Run("/system/scheduler/add",
			"=name=Expire-Monitor",
			"=start-time=00:00:00",
			"=interval=00:01:00",
			"=on-event="+script,
			"=disabled=no",
			"=comment=Expire Monitor System - Auto-generated",
		)
		if err != nil {
			return nil, err
		}

		return &ExpireMonitorResponse{
			Message: "Expire monitor created successfully",
			Action:  "created",
			Data: ExpireMonitorData{
				ID:   addReply.Done.Map["ret"],
				Name: "Expire-Monitor",
			},
		}, nil
	}

	// Monitor exists - get details
	monitor := reply.Re[0].Map
	monitorID := monitor[".id"]
	disabled := yesNoToBool(monitor["disabled"])

	// Case 2: Monitor exists but disabled - enable and update
	if disabled {
		_, err := s.client.Run("/system/scheduler/set",
			"=.id="+monitorID,
			"=start-time=00:00:00",
			"=interval=00:01:00",
			"=on-event="+script,
			"=disabled=no",
		)
		if err != nil {
			return nil, err
		}

		return &ExpireMonitorResponse{
			Message: "Expire monitor enabled and updated successfully",
			Action:  "updated",
			Data: ExpireMonitorData{
				ID:   monitorID,
				Name: "Expire-Monitor",
			},
		}, nil
	}

	// Case 3: Monitor exists and enabled - update script only
	_, err = s.client.Run("/system/scheduler/set",
		"=.id="+monitorID,
		"=on-event="+script,
	)
	if err != nil {
		return nil, err
	}

	return &ExpireMonitorResponse{
		Message: "Expire monitor already exists and is active",
		Action:  "already_exists",
		Data: ExpireMonitorData{
			ID:   monitorID,
			Name: "Expire-Monitor",
		},
	}, nil
}

// ExpireMonitorStatus adalah status expire monitor
type ExpireMonitorStatus struct {
	Exists  bool        `json:"exists"`
	Enabled bool        `json:"enabled"`
	Data    interface{} `json:"data,omitempty"`
}

// GetExpireMonitorStatus mendapatkan status expire monitor
func (s *Service) GetExpireMonitorStatus() (*ExpireMonitorStatus, error) {
	reply, err := s.client.Run("/system/scheduler/print", "?name=Expire-Monitor")
	if err != nil {
		return nil, err
	}

	if len(reply.Re) == 0 {
		return &ExpireMonitorStatus{
			Exists:  false,
			Enabled: false,
		}, nil
	}

	monitor := reply.Re[0].Map
	return &ExpireMonitorStatus{
		Exists:  true,
		Enabled: !yesNoToBool(monitor["disabled"]),
		Data: map[string]interface{}{
			"id":        monitor[".id"],
			"name":      monitor["name"],
			"startTime": monitor["start-time"],
			"interval":  monitor["interval"],
			"nextRun":   monitor["next-run"],
			"runCount":  monitor["run-count"],
			"disabled":  yesNoToBool(monitor["disabled"]),
		},
	}, nil
}

// DisableExpireMonitor menonaktifkan expire monitor
func (s *Service) DisableExpireMonitor() error {
	reply, err := s.client.Run("/system/scheduler/print", "?name=Expire-Monitor")
	if err != nil {
		return err
	}

	if len(reply.Re) == 0 {
		return fmt.Errorf("expire monitor not found")
	}

	monitorID := reply.Re[0].Map[".id"]
	_, err = s.client.Run("/system/scheduler/set",
		"=.id="+monitorID,
		"=disabled=yes",
	)
	return err
}

// EnableExpireMonitor mengaktifkan expire monitor
func (s *Service) EnableExpireMonitor() error {
	reply, err := s.client.Run("/system/scheduler/print", "?name=Expire-Monitor")
	if err != nil {
		return err
	}

	if len(reply.Re) == 0 {
		return fmt.Errorf("expire monitor not found")
	}

	monitorID := reply.Re[0].Map[".id"]
	_, err = s.client.Run("/system/scheduler/set",
		"=.id="+monitorID,
		"=disabled=no",
	)
	return err
}

// RemoveExpireMonitor menghapus expire monitor
func (s *Service) RemoveExpireMonitor() error {
	reply, err := s.client.Run("/system/scheduler/print", "?name=Expire-Monitor")
	if err != nil {
		return err
	}

	if len(reply.Re) == 0 {
		return fmt.Errorf("expire monitor not found")
	}

	monitorID := reply.Re[0].Map[".id"]
	_, err = s.client.Run("/system/scheduler/remove", "=.id="+monitorID)
	return err
}