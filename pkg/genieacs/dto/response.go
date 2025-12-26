// pkg/genieacs/dto/response.go
package genieacs_dto

// TaskResponse response dari task GenieACS
type TaskResponse struct {
	ID        string                 `json:"_id,omitempty"`
	DeviceID  string                 `json:"device,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Fault     map[string]interface{} `json:"fault,omitempty"`
}

// DeviceInfoResponse response informasi device
type DeviceInfoResponse struct {
	ID            string   `json:"id"`
	SerialNumber  string   `json:"serial_number"`
	Manufacturer  string   `json:"manufacturer"`
	Model         string   `json:"model"`
	Firmware      string   `json:"firmware"`
	Hardware      string   `json:"hardware"`
	LastInform    *string  `json:"last_inform"`
	Online        bool     `json:"online"`
	MAC           string   `json:"mac"`
	PPPoEUsername string   `json:"pppoe_username,omitempty"`
	RXPower       *float64 `json:"rx_power,omitempty"`
	Temperature   *float64 `json:"temperature,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// WiFiInfoResponse response informasi WiFi
type WiFiInfoResponse struct {
	SSID24GHz string `json:"ssid_2_4ghz"`
	SSID5GHz  string `json:"ssid_5ghz,omitempty"`
	Band      string `json:"band"`
}

// MonitorResponse response untuk monitoring
type MonitorResponse struct {
	TotalDevices     int              `json:"total_devices"`
	CriticalDevices  []CriticalDevice `json:"critical_devices,omitempty"`
	OfflineDevices   []OfflineDevice  `json:"offline_devices,omitempty"`
	Threshold        interface{}      `json:"threshold"`
	MonitoredAt      string           `json:"monitored_at"`
}

// CriticalDevice device dengan kondisi kritis
type CriticalDevice struct {
	DeviceID      string  `json:"device_id"`
	SerialNumber  string  `json:"serial_number"`
	RXPower       float64 `json:"rx_power"`
	PPPoEUsername string  `json:"pppoe_username"`
	LastInform    *string `json:"last_inform"`
}

// OfflineDevice device yang offline
type OfflineDevice struct {
	DeviceID      string  `json:"device_id"`
	SerialNumber  string  `json:"serial_number"`
	PPPoEUsername string  `json:"pppoe_username"`
	LastInform    string  `json:"last_inform"`
	OfflineHours  float64 `json:"offline_hours"`
}

// ErrorResponse response untuk error
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}