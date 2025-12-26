package entity

import (
    "encoding/json"
    "time"
)

// Role represents the roles table
type Role struct {
    ID           int64           `gorm:"primaryKey;column:id"`
    Name         string          `gorm:"column:name;type:varchar(50);unique;not null"`
    DisplayName  string          `gorm:"column:display_name;type:varchar(100);not null"`
    Description  string          `gorm:"column:description;type:text"`
    Permissions  json.RawMessage `gorm:"column:permissions;type:jsonb;default:'[]'"`
    IsSystem     bool            `gorm:"column:is_system;default:false"`
    IsActive     bool            `gorm:"column:is_active;default:true"`
    CreatedAt    time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt    time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

    // Relations
    Users []User `gorm:"foreignKey:RoleID"`
}

func (Role) TableName() string { return "roles" }


