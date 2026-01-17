// file: models/profile.go
package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)


type ProfileType string

const (

	ProfileTypePPPoE     ProfileType = "pppoe"
	ProfileTypeHotspot   ProfileType = "hotspot"
	ProfileTypeStaticIP  ProfileType = "static_ip"

)


// MikrotikProfile - Network profile base
type MikrotikProfile struct {
	ID         string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID   string     `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MikrotikID string     `gorm:"type:uuid;not null;index" json:"mikrotik_id"`
	Name       string     `gorm:"type:varchar(100);not null" json:"name"`
	Type       ProfileType `gorm:"type:profile_type;not null" json:"type"`
	IsDefault  bool       `gorm:"default:false" json:"is_default"`
	
	// Common parameters
	RateLimit      *string `gorm:"type:varchar(100)" json:"rate_limit,omitempty"`
	SessionTimeout *string `gorm:"type:interval" json:"session_timeout,omitempty"`
	IdleTimeout    *string `gorm:"type:interval" json:"idle_timeout,omitempty"`
	
	// Billing parameters
	Price    float64 `gorm:"type:numeric(15,2);default:0.00" json:"price"`
	TaxRate  float64 `gorm:"type:numeric(5,2);default:0.00" json:"tax_rate"`
	
	// Metadata & Audit
	Metadata  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant      Tenant                `gorm:"foreignKey:TenantID" json:"tenant"`
	Mikrotik    Mikrotik              `gorm:"foreignKey:MikrotikID" json:"mikrotik"`
	PPPoEDetails *MikrotikProfilePPPoE `gorm:"foreignKey:ProfileID" json:"pppoe_details,omitempty"`
	HotspotDetails *MikrotikProfileHotspot `gorm:"foreignKey:ProfileID" json:"hotspot_details,omitempty"`
	StaticIPDetails *MikrotikProfileStaticIP `gorm:"foreignKey:ProfileID" json:"static_ip_details,omitempty"`
	Services     []CustomerService    `gorm:"foreignKey:ProfileID" json:"services,omitempty"`
}

func (MikrotikProfile) TableName() string {
	return "mikrotik_profiles"
}

// MikrotikProfilePPPoE - PPPoE specific details
type MikrotikProfilePPPoE struct {
	ProfileID        string  `gorm:"type:uuid;primaryKey" json:"profile_id"`
	LocalAddress     *string `gorm:"type:inet" json:"local_address,omitempty"`
	RemoteAddress    *string `gorm:"type:inet" json:"remote_address,omitempty"`
	AddressPool      *string `gorm:"type:varchar(50)" json:"address_pool,omitempty"`
	MTU              int     `gorm:"default:1480" json:"mtu"`
	MRU              int     `gorm:"default:1480" json:"mru"`
	UseMPLS          bool    `gorm:"default:false" json:"use_mpls"`
	UseCompression   bool    `gorm:"default:false" json:"use_compression"`
	UseEncryption    bool    `gorm:"default:false" json:"use_encryption"`

	// Relation
	Profile MikrotikProfile `gorm:"foreignKey:ProfileID" json:"profile"`
}

func (MikrotikProfilePPPoE) TableName() string {
	return "mikrotik_profile_pppoe"
}

// MikrotikProfileHotspot - Hotspot specific details
type MikrotikProfileHotspot struct {
	ProfileID              string  `gorm:"type:uuid;primaryKey" json:"profile_id"`
	SharedUsers            int     `gorm:"default:1" json:"shared_users"`
	AddressPool            *string `gorm:"type:varchar(50)" json:"address_pool,omitempty"`
	MacAuth                bool    `gorm:"default:false" json:"mac_auth"`
	MacAuthMode            string  `gorm:"type:varchar(20);default:'none'" json:"mac_auth_mode"`
	LoginTimeoutSeconds    *int    `json:"login_timeout_seconds,omitempty"`
	CookieTimeoutSeconds   *int    `json:"cookie_timeout_seconds,omitempty"`

	// Relation
	Profile MikrotikProfile `gorm:"foreignKey:ProfileID" json:"profile"`
}

func (MikrotikProfileHotspot) TableName() string {
	return "mikrotik_profile_hotspot"
}

// MikrotikProfileStaticIP - Static IP specific details
type MikrotikProfileStaticIP struct {
	ProfileID    string  `gorm:"type:uuid;primaryKey" json:"profile_id"`
	IPAddress    *string `gorm:"type:inet" json:"ip_address,omitempty"`
	Gateway      *string `gorm:"type:inet" json:"gateway,omitempty"`
	VLANID       *int    `json:"vlan_id,omitempty"`
	RoutingMark  *string `gorm:"type:varchar(50)" json:"routing_mark,omitempty"`

	// Relation
	Profile MikrotikProfile `gorm:"foreignKey:ProfileID" json:"profile"`
}

func (MikrotikProfileStaticIP) TableName() string {
	return "mikrotik_profile_static_ip"
}

// REQUEST/RESPONSE MODELS
type CreateProfileRequest struct {
	Name          string      `json:"name" binding:"required"`
	Type          ProfileType `json:"type" binding:"required,oneof=pppoe hotspot static_ip"`
	RateLimit     *string     `json:"rate_limit,omitempty"`
	SessionTimeout *int       `json:"session_timeout,omitempty"`
	IdleTimeout   *int        `json:"idle_timeout,omitempty"`
	Price         float64     `json:"price" binding:"required"`
	TaxRate       *float64    `json:"tax_rate,omitempty"`
	// PPPoE specific
	LocalAddress   *string `json:"local_address,omitempty"`
	RemoteAddress  *string `json:"remote_address,omitempty"`
	AddressPool    *string `json:"address_pool,omitempty"`
	MTU            *int    `json:"mtu,omitempty"`
	MRU            *int    `json:"mru,omitempty"`
	// Hotspot specific
	SharedUsers  *int    `json:"shared_users,omitempty"`
	MacAuth      *bool   `json:"mac_auth,omitempty"`
	MacAuthMode  *string `json:"mac_auth_mode,omitempty"`
	// Static IP specific
	IPAddress   *string `json:"ip_address,omitempty"`
	Gateway     *string `json:"gateway,omitempty"`
	VLANID      *int    `json:"vlan_id,omitempty"`
	RoutingMark *string `json:"routing_mark,omitempty"`
}

type ProfileResponse struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	MikrotikID    string     `json:"mikrotik_id"`
	Name          string     `json:"name"`
	Type          ProfileType `json:"type"`
	IsDefault     bool       `json:"is_default"`
	RateLimit     *string    `json:"rate_limit,omitempty"`
	SessionTimeout *string   `json:"session_timeout,omitempty"`
	IdleTimeout   *string    `json:"idle_timeout,omitempty"`
	Price         float64    `json:"price"`
	TaxRate       float64    `json:"tax_rate"`
	Metadata      datatypes.JSON `json:"metadata"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// Type-specific details
	PPPoEDetails   *MikrotikProfilePPPoE     `json:"pppoe_details,omitempty"`
	HotspotDetails *MikrotikProfileHotspot   `json:"hotspot_details,omitempty"`
	StaticIPDetails *MikrotikProfileStaticIP `json:"static_ip_details,omitempty"`
}

type ProfileListResponse struct {
	Profiles []ProfileResponse `json:"profiles"`
	Total    int64             `json:"total"`
	PPPoE    int64             `json:"pppoe"`
	Hotspot  int64             `json:"hotspot"`
	StaticIP int64             `json:"static_ip"`
}

func (p *MikrotikProfile) ToResponse() *ProfileResponse {
	resp := &ProfileResponse{
		ID:            p.ID,
		TenantID:      p.TenantID,
		MikrotikID:    p.MikrotikID,
		Name:          p.Name,
		Type:          p.Type,
		IsDefault:     p.IsDefault,
		RateLimit:     p.RateLimit,
		SessionTimeout: p.SessionTimeout,
		IdleTimeout:   p.IdleTimeout,
		Price:         p.Price,
		TaxRate:       p.TaxRate,
		Metadata:      p.Metadata,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}

	// Add type-specific details based on profile type
	switch p.Type {
	case ProfileTypePPPoE:
		if p.PPPoEDetails != nil {
			resp.PPPoEDetails = p.PPPoEDetails
		}
	case ProfileTypeHotspot:
		if p.HotspotDetails != nil {
			resp.HotspotDetails = p.HotspotDetails
		}
	case ProfileTypeStaticIP:
		if p.StaticIPDetails != nil {
			resp.StaticIPDetails = p.StaticIPDetails
		}
	}

	return resp
}