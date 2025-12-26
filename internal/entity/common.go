package entity

import (
	"time"
)

// BaseEntity contains common fields for all entities
type BaseEntity struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
}

// SoftDeleteEntity adds soft delete support
type SoftDeleteEntity struct {
	BaseEntity
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamptz"`
}