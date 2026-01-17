// file: models/global_settings.go
package model

import (
	"time"
)

// GlobalSetting - System-wide settings
type GlobalSetting struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SettingKey   string    `gorm:"type:varchar(100);unique;not null" json:"setting_key"`
	SettingValue *string   `gorm:"type:text" json:"setting_value,omitempty"`
	SettingType  string    `gorm:"type:varchar(50);default:'string'" json:"setting_type"`
	Category     string    `gorm:"type:varchar(50);default:'general'" json:"category"`
	Description  *string   `gorm:"type:text" json:"description,omitempty"`
	IsPublic     bool      `gorm:"default:false" json:"is_public"`
	CreatedBy    *string   `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy    *string   `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	CreatedByUser *User `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	UpdatedByUser *User `gorm:"foreignKey:UpdatedBy" json:"updated_by_user,omitempty"`
}

func (GlobalSetting) TableName() string {
	return "global_settings"
}

// ============================================================================
// GLOBAL SETTINGS REQUEST/RESPONSE MODELS
// ============================================================================

type CreateGlobalSettingRequest struct {
	SettingKey   string  `json:"setting_key" binding:"required"`
	SettingValue string  `json:"setting_value" binding:"required"`
	SettingType  string  `json:"setting_type" binding:"required,oneof=string number boolean json"`
	Category     string  `json:"category" binding:"required"`
	Description  *string `json:"description,omitempty"`
	IsPublic     *bool   `json:"is_public,omitempty"`
}

type UpdateGlobalSettingRequest struct {
	SettingValue *string `json:"setting_value,omitempty"`
	SettingType  *string `json:"setting_type,omitempty" binding:"omitempty,oneof=string number boolean json"`
	Category     *string `json:"category,omitempty"`
	Description  *string `json:"description,omitempty"`
	IsPublic     *bool   `json:"is_public,omitempty"`
}

type GlobalSettingResponse struct {
	ID           string    `json:"id"`
	SettingKey   string    `json:"setting_key"`
	SettingValue *string   `json:"setting_value,omitempty"`
	SettingType  string    `json:"setting_type"`
	Category     string    `json:"category"`
	Description  *string   `json:"description,omitempty"`
	IsPublic     bool      `json:"is_public"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}