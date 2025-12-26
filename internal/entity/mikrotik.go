package entity

import "time"


type MikrotikStatus string

const (

    MikrotikStatusOnline      MikrotikStatus = "online"
	MikrotikStatusOffline     MikrotikStatus = "offline"
	MikrotikStatusError       MikrotikStatus = "error"
)

type Mikrotik struct {
	ID                    string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name                  string        `gorm:"column:name;type:text;not null"`
	Host                  string        `gorm:"column:host;type:inet;not null"`
	Port                  int           `gorm:"column:port;not null;default:8728"`
	APIUsername           string        `gorm:"column:api_username;type:text;not null"`
	APIEncryptedPassword  string        `gorm:"column:api_encrypted_password;type:text"`
	Keepalive             bool          `gorm:"column:keepalive;default:true"`
	Timeout               int           `gorm:"column:timeout;default:300000"`
	Location              string        `gorm:"column:location;type:varchar(100)"`
	Description           string        `gorm:"column:description;type:text"`
	IsActive              bool          `gorm:"column:is_active;not null;default:true"`
	Status                MikrotikStatus `gorm:"column:status;type:mikrotik_status;not null;default:'offline'"`
	Version               string        `gorm:"column:version;type:varchar(50)"`
	Uptime                string        `gorm:"column:uptime;type:varchar(50)"`
	CPUUsage              int           `gorm:"column:cpu_usage"`
	MemoryUsage           int           `gorm:"column:memory_usage"`
	LastSync              *time.Time    `gorm:"column:last_sync;type:timestamptz"`
	SyncInterval          int           `gorm:"column:sync_interval;default:300"`
	CreatedAt             time.Time     `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time     `gorm:"column:updated_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

	// Relations (these are defined in respective files)
}

func (Mikrotik) TableName() string { return "mikrotik" }