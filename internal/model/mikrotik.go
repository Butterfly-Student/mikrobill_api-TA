// file: models/mikrotik.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type MikrotikStatus string

const (
	MikrotikStatusOnline  MikrotikStatus = "online"
	MikrotikStatusOffline MikrotikStatus = "offline"
	MikrotikStatusError   MikrotikStatus = "error"
)

// Mikrotik - MikroTik device model
type Mikrotik struct {
	ID                   string        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID             string        `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name                 string        `gorm:"type:text;not null" json:"name"`
	Host                 string        `gorm:"type:inet;not null" json:"host"`
	Port                 int           `gorm:"not null;default:8728" json:"port"`
	APIUsername          string        `gorm:"type:text;not null" json:"api_username"`
	APIEncryptedPassword *string       `gorm:"type:text" json:"-"`
	Keepalive            bool          `gorm:"default:true" json:"keepalive"`
	Timeout              int           `gorm:"default:300000" json:"timeout"`
	Location             *string       `gorm:"type:varchar(100)" json:"location,omitempty"`
	Description          *string       `gorm:"type:text" json:"description,omitempty"`
	IsActive             bool          `gorm:"not null;default:false" json:"is_active"`
	Status               MikrotikStatus `gorm:"type:mikrotik_status;default:'offline'" json:"status"`
	LastSync             *time.Time    `json:"last_sync,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant    Tenant            `gorm:"foreignKey:TenantID" json:"tenant"`
	Profiles  []MikrotikProfile `gorm:"foreignKey:MikrotikID" json:"profiles,omitempty"`
	Customers []Customer        `gorm:"foreignKey:MikrotikID" json:"customers,omitempty"`
}

func (Mikrotik) TableName() string {
	return "mikrotik"
}

// REQUEST/RESPONSE MODELS
type CreateMikrotikRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Host                 string  `json:"host" binding:"required,ip"`
	Port                 int     `json:"port" binding:"omitempty,min=1,max=65535"`
	APIUsername          string  `json:"api_username" binding:"required"`
	APIPassword          string  `json:"api_password" binding:"required"`
	Keepalive            *bool   `json:"keepalive,omitempty"`
	Timeout              *int    `json:"timeout,omitempty"`
	Location             *string `json:"location,omitempty"`
	Description          *string `json:"description,omitempty"`
}

type UpdateMikrotikRequest struct {
	Name                 *string `json:"name,omitempty"`
	Host                 *string `json:"host,omitempty" binding:"omitempty,ip"`
	Port                 *int    `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	APIUsername          *string `json:"api_username,omitempty"`
	APIPassword          *string `json:"api_password,omitempty"`
	Keepalive            *bool   `json:"keepalive,omitempty"`
	Timeout              *int    `json:"timeout,omitempty"`
	Location             *string `json:"location,omitempty"`
	Description          *string `json:"description,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
}

type MikrotikResponse struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	Name        string        `json:"name"`
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	APIUsername string        `json:"api_username"`
	Keepalive   bool          `json:"keepalive"`
	Timeout     int           `json:"timeout"`
	Location    *string       `json:"location,omitempty"`
	Description *string       `json:"description,omitempty"`
	IsActive    bool          `json:"is_active"`
	Status      MikrotikStatus `json:"status"`
	LastSync    *time.Time    `json:"last_sync,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	TotalProfiles  int `json:"total_profiles,omitempty"`
	TotalCustomers int `json:"total_customers,omitempty"`
}

type MikrotikListResponse struct {
	Mikrotiks []MikrotikResponse `json:"mikrotiks"`
	Total     int64              `json:"total"`
	Online    int64              `json:"online"`
	Offline   int64              `json:"offline"`
	Error     int64              `json:"error"`
}


func (m *Mikrotik) ToResponse() *MikrotikResponse {
	return &MikrotikResponse{
		ID:          m.ID,
		TenantID:    m.TenantID,
		Name:        m.Name,
		Host:        m.Host,
		Port:        m.Port,
		APIUsername: m.APIUsername,
		Keepalive:   m.Keepalive,
		Timeout:     m.Timeout,
		Location:    m.Location,
		Description: m.Description,
		IsActive:    m.IsActive,
		Status:      m.Status,
		LastSync:    m.LastSync,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (s *MikrotikStatus) Scan(value interface{}) error {
	*s = MikrotikStatus(value.(string))
	return nil
}

func (s MikrotikStatus) Value() (interface{}, error) {
	return string(s), nil
}