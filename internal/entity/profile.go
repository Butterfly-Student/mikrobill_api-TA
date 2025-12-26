package entity

import "time"



type ProfileType string

const (
	ProfileTypePPPoE   ProfileType = "pppoe"
	ProfileTypeHotspot ProfileType = "hotspot"
)

type MikrotikProfile struct {
	ID                   string      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID           string      `gorm:"column:mikrotik_id;type:uuid;not null"`
	Name                 string      `gorm:"column:name;type:varchar(100);not null"`
	ProfileType          ProfileType `gorm:"column:profile_type;type:profile_type;not null"`
	RateLimitUp          string      `gorm:"column:rate_limit_up;type:varchar(50)"`
	RateLimitDown        string      `gorm:"column:rate_limit_down;type:varchar(50)"`
	IdleTimeout          string      `gorm:"column:idle_timeout;type:varchar(20)"`
	SessionTimeout       string      `gorm:"column:session_timeout;type:varchar(20)"`
	KeepaliveTimeout     string      `gorm:"column:keepalive_timeout;type:varchar(20)"`
	OnlyOne              bool        `gorm:"column:only_one;default:false"`
	StatusAuthentication bool        `gorm:"column:status_authentication;default:true"`
	DNSServer            string      `gorm:"column:dns_server;type:varchar(100)"`
	IsActive             bool        `gorm:"column:is_active;default:true"`
	SyncWithMikrotik     bool        `gorm:"column:sync_with_mikrotik;default:true"`
	LastSync             *time.Time  `gorm:"column:last_sync;type:timestamptz"`
	CreatedAt            time.Time   `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt            time.Time   `gorm:"column:updated_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik *Mikrotik `gorm:"foreignKey:MikrotikID"`
}

func (MikrotikProfile) TableName() string { return "mikrotik_profiles" }