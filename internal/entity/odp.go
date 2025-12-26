package entity

import "time"


type ODPStatus string

const (
	ODPStatusActive      ODPStatus = "active"
	ODPStatusMaintenance ODPStatus = "maintenance"
	ODPStatusInactive    ODPStatus = "inactive"
)

// ODP represents the odps table
type ODP struct {
	ID               string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name             string     `gorm:"column:name;type:varchar(100);unique;not null"`
	Code             string     `gorm:"column:code;type:varchar(50);unique;not null"`
	ParentODPID      *string    `gorm:"column:parent_odp_id;type:uuid"`
	Latitude         float64    `gorm:"column:latitude;type:decimal(10,8);not null"`
	Longitude        float64    `gorm:"column:longitude;type:decimal(11,8);not null"`
	Address          string     `gorm:"column:address;type:text"`
	Capacity         int        `gorm:"column:capacity;default:64"`
	UsedPorts        int        `gorm:"column:used_ports;default:0"`
	IsPole           bool       `gorm:"column:is_pole;default:false"`
	Status           ODPStatus  `gorm:"column:status;type:odp_status;default:'active'"`
	InstallationDate *time.Time `gorm:"column:installation_date;type:date"`
	Notes            string     `gorm:"column:notes;type:text"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	ParentODP       *ODP             `gorm:"foreignKey:ParentODPID"`
	ChildODPs       []ODP            `gorm:"foreignKey:ParentODPID"`
	ConnectionsFrom []ODPConnection  `gorm:"foreignKey:FromODPID"`
	ConnectionsTo   []ODPConnection  `gorm:"foreignKey:ToODPID"`
	Customers       []Customer       `gorm:"foreignKey:ODPID"`
	CableRoutes     []CableRoute     `gorm:"foreignKey:ODPID"`
	SegmentsStart   []NetworkSegment `gorm:"foreignKey:StartODPID"`
	SegmentsEnd     []NetworkSegment `gorm:"foreignKey:EndODPID"`
	ONUDevices      []ONUDevice      `gorm:"foreignKey:ODPID"`
}

func (ODP) TableName() string { return "odps" }