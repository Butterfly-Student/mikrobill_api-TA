package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Tenant struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	Subdomain   *string        `gorm:"type:varchar(100);unique" json:"subdomain,omitempty"`
	CompanyName *string        `gorm:"type:varchar(200)" json:"company_name,omitempty"`
	Phone       *string        `gorm:"type:varchar(50)" json:"phone,omitempty"`
	Address     *string        `gorm:"type:text" json:"address,omitempty"`
	Timezone    string         `gorm:"type:varchar(50);default:'Asia/Jakarta'" json:"timezone"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"`

	// Limit Management
	MaxMikrotiks    int `gorm:"default:3" json:"max_mikrotiks"`
	MaxNetworkUsers int `gorm:"default:50" json:"max_network_users"`
	MaxStaffUsers   int `gorm:"default:5" json:"max_staff_users"`

	// Features & Metadata
	Features datatypes.JSON `gorm:"type:jsonb;default:'{\"api_access\":true,\"reports\":true,\"backup\":true}'" json:"features"`
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`

	// Audit & Lifecycle
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	SuspendedAt *time.Time     `json:"suspended_at,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Users         []User            `gorm:"foreignKey:TenantID" json:"-"`
	Mikrotiks     []Mikrotik        `gorm:"foreignKey:TenantID" json:"-"`
	Roles         []Role            `gorm:"foreignKey:TenantID" json:"-"`
	Customers     []Customer        `gorm:"foreignKey:TenantID" json:"-"`
	Profiles      []MikrotikProfile `gorm:"foreignKey:TenantID" json:"-"`
	ActivityLogs  []ActivityLog     `gorm:"foreignKey:TenantID" json:"-"`
	TenantUsers   []TenantUser      `gorm:"foreignKey:TenantID" json:"-"`
}

func (Tenant) TableName() string {
	return "tenants"
}
// TenantUser - User association with tenant
type TenantUser struct {
	TenantID  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    *uuid.UUID `gorm:"type:uuid" json:"role_id,omitempty"`
	IsPrimary bool       `gorm:"default:false" json:"is_primary"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Role   *Role  `gorm:"foreignKey:RoleID" json:"-"`
}

func (TenantUser) TableName() string {
	return "tenant_users"
}

// ============================================================================
// TENANT REQUEST/RESPONSE MODELS
// ============================================================================

type CreateTenantRequest struct {
	Name            string         `json:"name" binding:"required,min=3,max=200"`
	Subdomain       *string        `json:"subdomain,omitempty" binding:"omitempty,min=3,max=100,alphanum"`
	CompanyName     *string        `json:"company_name,omitempty" binding:"omitempty,max=200"`
	Phone           *string        `json:"phone,omitempty" binding:"omitempty,max=50"`
	Address         *string        `json:"address,omitempty"`
	Timezone        *string        `json:"timezone,omitempty" binding:"omitempty,timezone"`
	MaxMikrotiks    *int           `json:"max_mikrotiks,omitempty" binding:"omitempty,min=1,max=100"`
	MaxNetworkUsers *int           `json:"max_network_users,omitempty" binding:"omitempty,min=1,max=100000"`
	MaxStaffUsers   *int           `json:"max_staff_users,omitempty" binding:"omitempty,min=1,max=1000"`
	Features        datatypes.JSON `json:"features,omitempty"`
	Metadata        datatypes.JSON `json:"metadata,omitempty"`
}

type UpdateTenantRequest struct {
	Name            *string        `json:"name,omitempty" binding:"omitempty,min=3,max=200"`
	Subdomain       *string        `json:"subdomain,omitempty" binding:"omitempty,min=3,max=100,alphanum"`
	CompanyName     *string        `json:"company_name,omitempty" binding:"omitempty,max=200"`
	Phone           *string        `json:"phone,omitempty" binding:"omitempty,max=50"`
	Address         *string        `json:"address,omitempty"`
	Timezone        *string        `json:"timezone,omitempty" binding:"omitempty,timezone"`
	IsActive        *bool          `json:"is_active,omitempty"`
	Status          *string        `json:"status,omitempty" binding:"omitempty,oneof=active suspended inactive"`
	MaxMikrotiks    *int           `json:"max_mikrotiks,omitempty" binding:"omitempty,min=1,max=100"`
	MaxNetworkUsers *int           `json:"max_network_users,omitempty" binding:"omitempty,min=1,max=100000"`
	MaxStaffUsers   *int           `json:"max_staff_users,omitempty" binding:"omitempty,min=1,max=1000"`
	Features        datatypes.JSON `json:"features,omitempty"`
	Metadata        datatypes.JSON `json:"metadata,omitempty"`
}

type TenantResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Subdomain       *string        `json:"subdomain,omitempty"`
	CompanyName     *string        `json:"company_name,omitempty"`
	Phone           *string        `json:"phone,omitempty"`
	Address         *string        `json:"address,omitempty"`
	Timezone        string         `json:"timezone"`
	IsActive        bool           `json:"is_active"`
	Status          string         `json:"status"`
	MaxMikrotiks    int            `json:"max_mikrotiks"`
	MaxNetworkUsers int            `json:"max_network_users"`
	MaxStaffUsers   int            `json:"max_staff_users"`
	Features        datatypes.JSON `json:"features"`
	Metadata        datatypes.JSON `json:"metadata"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	SuspendedAt     *time.Time     `json:"suspended_at,omitempty"`
}

type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// TenantFilter represents filter options for listing tenants
type TenantFilter struct {
	Status   *string `json:"status,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	Search   *string `json:"search,omitempty"` // Search in name, company_name
	Limit    int     `json:"limit" binding:"omitempty,min=1,max=100"`
	Offset   int     `json:"offset" binding:"omitempty,min=0"`
}





// TenantStatsResponse represents usage statistics for a tenant
type TenantStatsResponse struct {
	TenantID                 uuid.UUID `json:"tenant_id"`
	MikrotiksCount           int       `json:"mikrotiks_count"`
	NetworkUsersCount        int       `json:"network_users_count"`
	StaffUsersCount          int       `json:"staff_users_count"`
	ActiveSessionsCount      int       `json:"active_sessions_count"`
	MaxMikrotiks             int       `json:"max_mikrotiks"`
	MaxNetworkUsers          int       `json:"max_network_users"`
	MaxStaffUsers            int       `json:"max_staff_users"`
	MikrotiksUsagePercent    float64   `json:"mikrotiks_usage_percent"`
	NetworkUsersUsagePercent float64   `json:"network_users_usage_percent"`
	StaffUsersUsagePercent   float64   `json:"staff_users_usage_percent"`
}

func (t *Tenant) ToResponse() *TenantResponse {
	return &TenantResponse{
		ID:              t.ID,
		Name:            t.Name,
		Subdomain:       t.Subdomain,
		CompanyName:     t.CompanyName,
		Phone:           t.Phone,
		Address:         t.Address,
		Timezone:        t.Timezone,
		IsActive:        t.IsActive,
		Status:          t.Status,
		MaxMikrotiks:    t.MaxMikrotiks,
		MaxNetworkUsers: t.MaxNetworkUsers,
		MaxStaffUsers:   t.MaxStaffUsers,
		Features:        t.Features,
		Metadata:        t.Metadata,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
		SuspendedAt:     t.SuspendedAt,
	}
}

func (t *Tenant) IsUnderLimit(resourceType string, currentCount int) bool {
	switch resourceType {
	case "mikrotik":
		return currentCount < t.MaxMikrotiks
	case "network_user":
		return currentCount < t.MaxNetworkUsers
	case "staff_user":
		return currentCount < t.MaxStaffUsers
	default:
		return false
	}
}

func (t *Tenant) HasFeature(feature string) bool {
	if t.Features == nil {
		return false
	}
	
	var features map[string]interface{}
	if err := t.Features.Scan(&features); err != nil {
		return false
	}
	
	if enabled, ok := features[feature].(bool); ok {
		return enabled
	}
	return false
}

