package entity

import "time"


type CableStatus string


const (
	CableStatusConnected    CableStatus = "connected"
	CableStatusDisconnected CableStatus = "disconnected"
	CableStatusMaintenance  CableStatus = "maintenance"
	CableStatusDamaged      CableStatus = "damaged"
)

// CableRoute represents the cable_routes table
type CableRoute struct {
	ID               string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CustomerID       string        `gorm:"column:customer_id;type:uuid;not null"`
	ODPID            string        `gorm:"column:odp_id;type:uuid;not null"`
	CableLength      float64       `gorm:"column:cable_length;type:decimal(8,2)"`
	CableType        string        `gorm:"column:cable_type;type:varchar(50);default:'Fiber Optic'"`
	InstallationDate *time.Time    `gorm:"column:installation_date;type:date"`
	Status           CableStatus   `gorm:"column:status;type:cable_status;default:'connected'"`
	PortNumber       int           `gorm:"column:port_number"`
	Notes            string        `gorm:"column:notes;type:text"`
	CreatedAt        time.Time     `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time     `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Customer       *Customer      `gorm:"foreignKey:CustomerID"`
	ODP            *ODP           `gorm:"foreignKey:ODPID"`
	MaintenanceLogs []CableMaintenanceLog `gorm:"foreignKey:CableRouteID"`
}

func (CableRoute) TableName() string { return "cable_routes" }