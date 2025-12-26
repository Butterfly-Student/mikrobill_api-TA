package entity

import "time"

type StaticIPAssignment struct {
	ID                  string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID          string     `gorm:"column:mikrotik_id;type:uuid;not null"`
	IPAddress           string     `gorm:"column:ip_address;type:inet;not null"`
	Gateway             string     `gorm:"column:gateway;type:inet"`
	Netmask             int        `gorm:"column:netmask"`
	AllowedMACAddresses []string   `gorm:"column:allowed_mac_addresses;type:text[]"`
	FirewallChain       string     `gorm:"column:firewall_chain;type:varchar(50)"`
	RateLimit           string     `gorm:"column:rate_limit;type:varchar(50)"`
	IsActive            bool       `gorm:"column:is_active;default:true"`
	SyncWithMikrotik    bool       `gorm:"column:sync_with_mikrotik;default:true"`
	LastSync            *time.Time `gorm:"column:last_sync;type:timestamptz"`
	CreatedAt           time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik *Mikrotik `gorm:"foreignKey:MikrotikID"`
}

func (StaticIPAssignment) TableName() string { return "static_ip_assignments" }