package model

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents mikrotik_profiles table
type Profile struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	TenantID                uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	MikrotikID              uuid.UUID  `json:"mikrotik_id" db:"mikrotik_id"`
	Name                    string     `json:"name" db:"name"`
	ProfileType             string     `json:"profile_type" db:"profile_type"` // pppoe, hotspot, static_ip
	MikrotikObjectID        string     `json:"mikrotik_object_id" db:"mikrotik_object_id"`
	RateLimitUpKbps         *int       `json:"rate_limit_up_kbps,omitempty" db:"rate_limit_up_kbps"`
	RateLimitDownKbps       *int       `json:"rate_limit_down_kbps,omitempty" db:"rate_limit_down_kbps"`
	IdleTimeoutSeconds      *int       `json:"idle_timeout_seconds,omitempty" db:"idle_timeout_seconds"`
	SessionTimeoutSeconds   *int       `json:"session_timeout_seconds,omitempty" db:"session_timeout_seconds"`
	KeepaliveTimeoutSeconds *int       `json:"keepalive_timeout_seconds,omitempty" db:"keepalive_timeout_seconds"`
	OnlyOne                 bool       `json:"only_one" db:"only_one"`
	StatusAuthentication    bool       `json:"status_authentication" db:"status_authentication"`
	DNSServer               *string    `json:"dns_server,omitempty" db:"dns_server"`
	IsActive                bool       `json:"is_active" db:"is_active"`
	SyncWithMikrotik        bool       `json:"sync_with_mikrotik" db:"sync_with_mikrotik"`
	Price                   float64    `json:"price" db:"price"`
	TaxRate                 float64    `json:"tax_rate" db:"tax_rate"`
	LastSync                *time.Time `json:"last_sync,omitempty" db:"last_sync"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
}

// ProfilePPPoE represents mikrotik_profile_pppoe table
type ProfilePPPoE struct {
	ProfileID      uuid.UUID `json:"profile_id" db:"profile_id"`
	LocalAddress   *string   `json:"local_address,omitempty" db:"local_address"`
	RemoteAddress  *string   `json:"remote_address,omitempty" db:"remote_address"`
	AddressPool    *string   `json:"address_pool,omitempty" db:"address_pool"`
	MTU            int       `json:"mtu" db:"mtu"`
	MRU            int       `json:"mru" db:"mru"`
	UseMPLS        bool      `json:"use_mpls" db:"use_mpls"`
	UseCompression bool      `json:"use_compression" db:"use_compression"`
	UseEncryption  bool      `json:"use_encryption" db:"use_encryption"`
}

// ProfileWithPPPoE combines Profile with PPPoE settings
type ProfileWithPPPoE struct {
	Profile
	PPPoE *ProfilePPPoE `json:"pppoe,omitempty"`
}

// ProfileInput is the request DTO for creating a profile
type ProfileInput struct {
	Name                    string   `json:"name" binding:"required"`
	RateLimitUpKbps         *int     `json:"rate_limit_up_kbps"`
	RateLimitDownKbps       *int     `json:"rate_limit_down_kbps"`
	IdleTimeoutSeconds      *int     `json:"idle_timeout_seconds"`
	SessionTimeoutSeconds   *int     `json:"session_timeout_seconds"`
	KeepaliveTimeoutSeconds *int     `json:"keepalive_timeout_seconds"`
	OnlyOne                 *bool    `json:"only_one"`
	StatusAuthentication    *bool    `json:"status_authentication"`
	DNSServer               *string  `json:"dns_server"`
	LocalAddress            *string  `json:"local_address"`
	RemoteAddress           *string  `json:"remote_address"`
	AddressPool             *string  `json:"address_pool"`
	Price                   float64  `json:"price" binding:"required"`
	TaxRate                 *float64 `json:"tax_rate"`
	MTU                     *int     `json:"mtu"`
	MRU                     *int     `json:"mru"`
	UseMPLS                 *bool    `json:"use_mpls"`
	UseCompression          *bool    `json:"use_compression"`
	UseEncryption           *bool    `json:"use_encryption"`
}

// PrepareProfileInput sets default values for optional fields
func PrepareProfileInput(input *ProfileInput) {
	if input.OnlyOne == nil {
		onlyOne := false
		input.OnlyOne = &onlyOne
	}
	if input.TaxRate == nil {
		taxRate := 11.00
		input.TaxRate = &taxRate
	}
	if input.StatusAuthentication == nil {
		statusAuth := true
		input.StatusAuthentication = &statusAuth
	}
	if input.MTU == nil {
		mtu := 1480
		input.MTU = &mtu
	}
	if input.MRU == nil {
		mru := 1480
		input.MRU = &mru
	}
	if input.UseMPLS == nil {
		useMPLS := false
		input.UseMPLS = &useMPLS
	}
	if input.UseCompression == nil {
		useCompression := false
		input.UseCompression = &useCompression
	}
	if input.UseEncryption == nil {
		useEncryption := false
		input.UseEncryption = &useEncryption
	}
}
