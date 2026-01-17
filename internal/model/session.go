// file: models/session.go
package model

import (
	"time"

	"gorm.io/datatypes"
)

// UserSession - Enhanced session management
type UserSession struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           string         `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID         *string        `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	TokenHash        string         `gorm:"type:text;unique;not null" json:"-"`
	RefreshTokenHash *string        `gorm:"type:text;unique" json:"-"`
	IPAddress        *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent        *string        `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceInfo       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"device_info"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	ExpiresAt        time.Time      `json:"expires_at"`
	RefreshedAt      *time.Time     `json:"refreshed_at,omitempty"`
	LastActivityAt   time.Time      `gorm:"default:now()" json:"last_activity_at"`
	CreatedAt        time.Time      `json:"created_at"`
	RevokedAt        *time.Time     `json:"revoked_at,omitempty"`
	RevokedBy        *string        `gorm:"type:uuid" json:"revoked_by,omitempty"`
	RevokeReason     *string        `gorm:"type:text" json:"revoke_reason,omitempty"`

	// Relations
	User           User               `gorm:"foreignKey:UserID" json:"user"`
	Tenant         *Tenant            `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	RevokedByUser  *User              `gorm:"foreignKey:RevokedBy" json:"revoked_by_user,omitempty"`
	Activities     []SessionActivity  `gorm:"foreignKey:SessionID" json:"activities,omitempty"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}

// SessionActivity - Session activity tracking
type SessionActivity struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SessionID string         `gorm:"type:uuid;not null;index" json:"session_id"`
	Action    string         `gorm:"type:varchar(100);not null" json:"action"`
	IPAddress *string        `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent *string        `gorm:"type:text" json:"user_agent,omitempty"`
	Metadata  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`

	// Relation
	Session UserSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

func (SessionActivity) TableName() string {
	return "session_activities"
}


type SessionResponse struct {
	ID             string         `json:"id"`
	UserID         string         `json:"user_id"`
	TenantID       *string        `json:"tenant_id,omitempty"`
	IPAddress      *string        `json:"ip_address,omitempty"`
	UserAgent      *string        `json:"user_agent,omitempty"`
	DeviceInfo     datatypes.JSON `json:"device_info"`
	IsActive       bool           `json:"is_active"`
	ExpiresAt      time.Time      `json:"expires_at"`
	RefreshedAt    *time.Time     `json:"refreshed_at,omitempty"`
	LastActivityAt time.Time      `json:"last_activity_at"`
	CreatedAt      time.Time      `json:"created_at"`
}