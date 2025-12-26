package entity

import "time"

type NetworkSegment struct {
	ID               string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name             string        `gorm:"column:name;type:varchar(100);not null"`
	StartODPID       string        `gorm:"column:start_odp_id;type:uuid;not null"`
	EndODPID         *string       `gorm:"column:end_odp_id;type:uuid"`
	SegmentType      SegmentType   `gorm:"column:segment_type;type:segment_type;default:'Backbone'"`
	CableLength      float64       `gorm:"column:cable_length;type:decimal(10,2)"`
	Status           SegmentStatus `gorm:"column:status;type:segment_status;default:'active'"`
	InstallationDate *time.Time    `gorm:"column:installation_date;type:date"`
	Notes            string        `gorm:"column:notes;type:text"`
	CreatedAt        time.Time     `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time     `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	StartODP *ODP `gorm:"foreignKey:StartODPID"`
	EndODP   *ODP `gorm:"foreignKey:EndODPID"`
	MaintenanceLogs []CableMaintenanceLog `gorm:"foreignKey:NetworkSegmentID"`
}

func (NetworkSegment) TableName() string { return "network_segments" }