package model

// ProvisioningMessage represents a message to provision a new customer on MikroTik
type ProvisioningMessage struct {
	CustomerID      string `json:"customer_id"`
	TenantID        string `json:"tenant_id"`
	ProfileID       string `json:"profile_id"`
	ServiceUsername string `json:"service_username"`
	ServicePassword string `json:"service_password"` // Decrypted plain text for MikroTik API
	ServiceType     string `json:"service_type"`     // "pppoe" or "hotspot"
}

// DisconnectMessage represents a message to disconnect a customer from MikroTik
type DisconnectMessage struct {
	CustomerID      string `json:"customer_id"`
	TenantID        string `json:"tenant_id"`
	ServiceUsername string `json:"service_username"`
	ServiceType     string `json:"service_type"` // "pppoe" or "hotspot"
}
