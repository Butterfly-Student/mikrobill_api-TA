package entity

import "time"

// ODPConnection represents the odp_connections table
type ODPConnection struct {
	ID               string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FromODPID        string         `gorm:"column:from_odp_id;type:uuid;not null"`
	ToODPID          string         `gorm:"column:to_odp_id;type:uuid;not null"`
	ConnectionType   ConnectionType `gorm:"column:connection_type;type:connection_type;default:'fiber'"`
	CableLength      float64        `gorm:"column:cable_length;type:decimal(8,2)"`
	CableCapacity    CableCapacity  `gorm:"column:cable_capacity;type:cable_capacity;default:'1G'"`
	Status           ODPStatus      `gorm:"column:status;type:odp_status;default:'active'"`
	InstallationDate *time.Time     `gorm:"column:installation_date;type:date"`
	Notes            string         `gorm:"column:notes;type:text"`
	CreatedAt        time.Time      `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	FromODP *ODP `gorm:"foreignKey:FromODPID"`
	ToODP   *ODP `gorm:"foreignKey:ToODPID"`
}

func (ODPConnection) TableName() string { return "odp_connections" }