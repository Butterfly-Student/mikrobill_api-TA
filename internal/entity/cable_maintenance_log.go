package entity

import "time"

// CableMaintenanceLog represents the cable_maintenance_logs table
type CableMaintenanceLog struct {
	ID                string          `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CableRouteID      *string         `gorm:"column:cable_route_id;type:uuid"`
	NetworkSegmentID  *string         `gorm:"column:network_segment_id;type:uuid"`
	MaintenanceType   MaintenanceType `gorm:"column:maintenance_type;type:maintenance_type;not null"`
	Description       string          `gorm:"column:description;type:text;not null"`
	PerformedBy       *string         `gorm:"column:performed_by;type:uuid"`
	MaintenanceDate   time.Time       `gorm:"column:maintenance_date;type:date;not null"`
	DurationHours     float64         `gorm:"column:duration_hours;type:decimal(4,2)"`
	Cost              float64         `gorm:"column:cost;type:decimal(12,2)"`
	Notes             string          `gorm:"column:notes;type:text"`
	CreatedAt         time.Time       `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	CableRoute     *CableRoute     `gorm:"foreignKey:CableRouteID"`
	NetworkSegment *NetworkSegment `gorm:"foreignKey:NetworkSegmentID"`
	PerformedByUser *User          `gorm:"foreignKey:PerformedBy"`
}

func (CableMaintenanceLog) TableName() string { return "cable_maintenance_logs" }