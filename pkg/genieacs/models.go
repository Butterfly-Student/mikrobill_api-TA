// pkg/genieacs/models.go
package genieacs

import (
	"time"
)

// Device merepresentasikan device dari GenieACS
type Device struct {
	ID                    string                 `json:"_id"`
	DeviceID              map[string]interface{} `json:"_deviceId,omitempty"`
	LastInform            *string                `json:"_lastInform,omitempty"`
	Registered            *string                `json:"_registered,omitempty"`
	Tags                  []string               `json:"_tags,omitempty"`
	InternetGatewayDevice map[string]interface{} `json:"InternetGatewayDevice,omitempty"`
	VirtualParameters     map[string]interface{} `json:"VirtualParameters,omitempty"`
	Device                map[string]interface{} `json:"Device,omitempty"`
}

// GetParameterValue helper untuk mendapatkan nilai parameter dari nested map
func (d *Device) GetParameterValue(path string) interface{} {
	// Parse path dan traverse nested map
	// Implementasi sederhana untuk common paths
	if d.VirtualParameters != nil {
		if val, ok := d.VirtualParameters[path]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				if v, ok := m["_value"]; ok {
					return v
				}
			}
			return val
		}
	}
	return nil
}

// GetRXPower mendapatkan nilai RX Power
func (d *Device) GetRXPower() *float64 {
	paths := []string{"RXPower", "redaman"}
	
	for _, path := range paths {
		if val := d.GetParameterValue(path); val != nil {
			if f, ok := val.(float64); ok {
				return &f
			}
		}
	}
	return nil
}

// GetSerialNumber mendapatkan serial number device
func (d *Device) GetSerialNumber() string {
	if d.DeviceID != nil {
		if sn, ok := d.DeviceID["_SerialNumber"].(string); ok {
			return sn
		}
	}
	return "Unknown"
}

// GetPPPoEUsername mendapatkan PPPoE username
func (d *Device) GetPPPoEUsername() string {
	if val := d.GetParameterValue("pppoeUsername"); val != nil {
		if username, ok := val.(string); ok {
			return username
		}
	}
	
	// Coba dari tags
	for _, tag := range d.Tags {
		if len(tag) > 6 && tag[:6] == "pppoe:" {
			return tag[6:]
		}
	}
	
	return "Unknown"
}

// IsOnline mengecek apakah device online
func (d *Device) IsOnline(thresholdMinutes int) bool {
	if d.LastInform == nil {
		return false
	}

	lastInform, err := time.Parse(time.RFC3339, *d.LastInform)
	if err != nil {
		return false
	}

	threshold := time.Duration(thresholdMinutes) * time.Minute
	return time.Since(lastInform) < threshold
}

// CriticalDevice device dengan RX Power kritis
type CriticalDevice struct {
	DeviceID      string  `json:"device_id"`
	SerialNumber  string  `json:"serial_number"`
	RXPower       float64 `json:"rx_power"`
	LastInform    *string `json:"last_inform"`
	PPPoEUsername string  `json:"pppoe_username"`
}

// OfflineDevice device yang offline
type OfflineDevice struct {
	DeviceID      string  `json:"device_id"`
	SerialNumber  string  `json:"serial_number"`
	PPPoEUsername string  `json:"pppoe_username"`
	LastInform    string  `json:"last_inform"`
	OfflineHours  float64 `json:"offline_hours"`
}

