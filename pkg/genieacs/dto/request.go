// pkg/genieacs/dto/request.go
package genieacs_dto

// SetParameterValuesRequest request untuk set parameter values
type SetParameterValuesRequest struct {
	DeviceID          string          `json:"device_id"`
	ParameterValues   [][]interface{} `json:"parameter_values"`
	ConnectionRequest bool            `json:"connection_request"`
}

// RefreshObjectRequest request untuk refresh object
type RefreshObjectRequest struct {
	DeviceID          string `json:"device_id"`
	ObjectName        string `json:"object_name"`
	ConnectionRequest bool   `json:"connection_request"`
}

// ChangeWiFiRequest request untuk mengubah WiFi SSID dan password
type ChangeWiFiRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	SSID     string `json:"ssid,omitempty"`
	Password string `json:"password,omitempty"`
	Band     string `json:"band,omitempty"` // "2.4" atau "5"
}

// AddWANRequest request untuk menambahkan konfigurasi WAN
type AddWANRequest struct {
	DeviceID       string `json:"device_id" binding:"required"`
	WANType        string `json:"wan_type" binding:"required,oneof=ppp ip"`
	ConnectionMode string `json:"connection_mode" binding:"required,oneof=bridge route"`
}

