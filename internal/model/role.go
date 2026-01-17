package model

import (
	"time"

	"gorm.io/datatypes"
)

// Role - Role model with permissions
type Role struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID    *string        `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	Name        string         `gorm:"type:varchar(50);not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null" json:"display_name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Permissions datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"permissions"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Users  []User  `gorm:"foreignKey:RoleID" json:"-"`
	TenantUsers []TenantUser `gorm:"foreignKey:RoleID" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

// REQUEST/RESPONSE MODELS
type CreateRoleRequest struct {
	Name        string         `json:"name" binding:"required,min=3,max=50"`
	DisplayName string         `json:"display_name" binding:"required,min=3,max=100"`
	Description *string        `json:"description,omitempty"`
	Permissions datatypes.JSON `json:"permissions" binding:"required"`
	IsActive    *bool          `json:"is_active,omitempty"`
}

type UpdateRoleRequest struct {
	Name        *string        `json:"name,omitempty" binding:"omitempty,min=3,max=50"`
	DisplayName *string        `json:"display_name,omitempty" binding:"omitempty,min=3,max=100"`
	Description *string        `json:"description,omitempty"`
	Permissions datatypes.JSON `json:"permissions,omitempty"`
	IsActive    *bool          `json:"is_active,omitempty"`
}

type RoleResponse struct {
	ID          string         `json:"id"`
	TenantID    *string        `json:"tenant_id,omitempty"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Description *string        `json:"description,omitempty"`
	Permissions datatypes.JSON `json:"permissions"`
	IsSystem    bool           `json:"is_system"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (r *Role) ToResponse() *RoleResponse {
	return &RoleResponse{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Permissions: r.Permissions,
		IsSystem:    r.IsSystem,
		IsActive:    r.IsActive,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}