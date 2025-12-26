package entity

import "time"

type CollectorStatus string


const (
	CollectorStatusActive    CollectorStatus = "active"
	CollectorStatusInactive  CollectorStatus = "inactive"
	CollectorStatusSuspended CollectorStatus = "suspended"
)


// Collector represents the collectors table
type Collector struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name           string         `gorm:"column:name;type:varchar(255);not null"`
	Phone          string         `gorm:"column:phone;type:varchar(20);not null"`
	Email          string         `gorm:"column:email;type:varchar(255)"`
	Address        string         `gorm:"column:address;type:text"`
	Status         CollectorStatus `gorm:"column:status;type:collector_status;default:'active'"`
	CommissionRate float64        `gorm:"column:commission_rate;type:decimal(5,2);default:5.00"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Payments []CollectorPayment `gorm:"foreignKey:CollectorID"`
}

func (Collector) TableName() string { return "collectors" }