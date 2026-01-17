// file: models/logging.go
package model

import (
	"time"

	"gorm.io/datatypes"
)

// ActivityLog - Tenant activity logging
type ActivityLog struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID     *string        `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	UserID       *string        `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action       string         `gorm:"type:varchar(100);not null" json:"action"`
	ResourceType string         `gorm:"type:varchar(100);not null" json:"resource_type"`
	ResourceID   *string        `gorm:"type:uuid" json:"resource_id,omitempty"`
	Description  *string        `gorm:"type:text" json:"description,omitempty"`
	OldValues    datatypes.JSON `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues    datatypes.JSON `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress    *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent    *string        `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

// SuperadminLog - Superadmin activity tracking
type SuperadminLog struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SuperadminID   *string        `gorm:"type:uuid;index" json:"superadmin_id,omitempty"`
	TargetTenantID *string        `gorm:"type:uuid;index" json:"target_tenant_id,omitempty"`
	Action         string         `gorm:"type:varchar(100);not null" json:"action"`
	ResourceType   *string        `gorm:"type:varchar(100)" json:"resource_type,omitempty"`
	ResourceID     *string        `gorm:"type:uuid" json:"resource_id,omitempty"`
	Description    *string        `gorm:"type:text" json:"description,omitempty"`
	OldValues      datatypes.JSON `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues      datatypes.JSON `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress      *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent      *string        `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`

	// Relations
	Superadmin   *User   `gorm:"foreignKey:SuperadminID" json:"superadmin,omitempty"`
	TargetTenant *Tenant `gorm:"foreignKey:TargetTenantID" json:"target_tenant,omitempty"`
}

func (SuperadminLog) TableName() string {
	return "superadmin_logs"
}

// REQUEST/RESPONSE MODELS
type ActivityLogResponse struct {
	ID           string         `json:"id"`
	TenantID     *string        `json:"tenant_id,omitempty"`
	UserID       *string        `json:"user_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   *string        `json:"resource_id,omitempty"`
	Description  *string        `json:"description,omitempty"`
	IPAddress    *string        `json:"ip_address,omitempty"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	// Relations
	User   *UserResponse   `json:"user,omitempty"`
	Tenant *TenantResponse `json:"tenant,omitempty"`
}

type ActivityLogFilter struct {
	TenantID     *string    `json:"tenant_id,omitempty"`
	UserID       *string    `json:"user_id,omitempty"`
	Action       *string    `json:"action,omitempty"`
	ResourceType *string    `json:"resource_type,omitempty"`
	ResourceID   *string    `json:"resource_id,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Limit        int        `json:"limit" default:"50"`
	Offset       int        `json:"offset" default:"0"`
}

type ActivityLogListResponse struct {
	Logs   []ActivityLogResponse `json:"logs"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func (al *ActivityLog) ToResponse() *ActivityLogResponse {
	return &ActivityLogResponse{
		ID:           al.ID,
		TenantID:     al.TenantID,
		UserID:       al.UserID,
		Action:       al.Action,
		ResourceType: al.ResourceType,
		ResourceID:   al.ResourceID,
		Description:  al.Description,
		IPAddress:    al.IPAddress,
		UserAgent:    al.UserAgent,
		CreatedAt:    al.CreatedAt,
	}
}