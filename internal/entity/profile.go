package entity

import (
	"time"
)

// MikrotikProfile represents a base profile (PPPoE, Hotspot, etc.)
type MikrotikProfile struct {
	ID                   string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MikrotikID           string     `json:"mikrotik_id" gorm:"column:mikrotik_id;not null;type:uuid"`
	Name                 string     `json:"name" gorm:"type:varchar(100);not null"`
	ProfileType          string     `json:"profile_type" gorm:"type:profile_type;not null"` // pppoe, hotspot, static_ip
	RateLimitUp          *string    `json:"rate_limit_up,omitempty" gorm:"column:rate_limit_up;type:varchar(50)"`
	RateLimitDown        *string    `json:"rate_limit_down,omitempty" gorm:"column:rate_limit_down;type:varchar(50)"`
	IdleTimeout          *string    `json:"idle_timeout,omitempty" gorm:"column:idle_timeout;type:varchar(20)"`
	SessionTimeout       *string    `json:"session_timeout,omitempty" gorm:"column:session_timeout;type:varchar(20)"`
	KeepaliveTimeout     *string    `json:"keepalive_timeout,omitempty" gorm:"column:keepalive_timeout;type:varchar(20)"`
	OnlyOne              bool       `json:"only_one" gorm:"column:only_one;default:false"`
	StatusAuthentication bool       `json:"status_authentication" gorm:"column:status_authentication;default:true"`
	DNSServer            *string    `json:"dns_server,omitempty" gorm:"column:dns_server;type:varchar(100)"`
	Price                *float64   `json:"price,omitempty" gorm:"type:decimal(15,2)"`
	IsActive             bool       `json:"is_active" gorm:"default:true"`
	SyncWithMikrotik     bool       `json:"sync_with_mikrotik" gorm:"column:sync_with_mikrotik;default:true"`
	LastSync             *time.Time `json:"last_sync,omitempty" gorm:"column:last_sync"`
	CreatedAt            time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"not null;default:now()"`

	// Relations
	Mikrotik     *Mikrotik             `json:"mikrotik,omitempty" gorm:"foreignKey:MikrotikID"`
	PPPoEDetails *MikrotikProfilePPPoE `json:"pppoe_details,omitempty" gorm:"foreignKey:ProfileID"`
}

// MikrotikProfilePPPoE represents PPPoE-specific profile settings
type MikrotikProfilePPPoE struct {
	ProfileID      string  `json:"profile_id" gorm:"primaryKey;type:uuid;column:profile_id"`
	LocalAddress   string  `json:"local_address" gorm:"column:local_address;type:varchar(50);not null"`
	RemoteAddress  *string `json:"remote_address,omitempty" gorm:"column:remote_address;type:varchar(50)"`
	AddressPool    string  `json:"address_pool" gorm:"column:address_pool;type:varchar(100);not null"`
	MTU            string  `json:"mtu" gorm:"type:varchar(10);default:'1480'"`
	MRU            string  `json:"mru" gorm:"type:varchar(10);default:'1480'"`
	ServiceName    *string `json:"service_name,omitempty" gorm:"column:service_name;type:varchar(50)"`
	MaxMTU         *string `json:"max_mtu,omitempty" gorm:"column:max_mtu;type:varchar(10)"`
	MaxMRU         *string `json:"max_mru,omitempty" gorm:"column:max_mru;type:varchar(10)"`
	UseMPLS        bool    `json:"use_mpls" gorm:"column:use_mpls;default:false"`
	UseCompression bool    `json:"use_compression" gorm:"column:use_compression;default:false"`
	UseEncryption  bool    `json:"use_encryption" gorm:"column:use_encryption;default:false"`

	// Relation
	Profile *MikrotikProfile `json:"-" gorm:"foreignKey:ProfileID"`
}

// ProfileWithPPPoE combines profile with PPPoE details for API responses
type ProfileWithPPPoE struct {
	MikrotikProfile
	PPPoEDetails *MikrotikProfilePPPoE `json:"pppoe_details,omitempty"`
}

// ProfileRepository defines database operations for profiles
type ProfileRepository interface {
	// CRUD operations
	CreateProfile(profile *MikrotikProfile, pppoeDetails *MikrotikProfilePPPoE) error
	UpdateProfile(profile *MikrotikProfile, pppoeDetails *MikrotikProfilePPPoE) error
	DeleteProfile(id string) error
	GetProfileByID(id string) (*ProfileWithPPPoE, error)

	// Query operations
	ListProfiles(mikrotikID *string, page, limit int) ([]*ProfileWithPPPoE, int64, error)
	GetProfilesByMikrotikID(mikrotikID string) ([]*ProfileWithPPPoE, error)
	GetProfileByName(mikrotikID string, name string) (*ProfileWithPPPoE, error)

	// Sync operations
	UpdateSyncStatus(id string, lastSync time.Time) error
}

func (MikrotikProfile) TableName() string {
	return "mikrotik_profiles"
}

func (MikrotikProfilePPPoE) TableName() string {
	return "mikrotik_profile_pppoe"
}
