package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// JSONB is a custom type for JSONB columns
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, j)
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Subdomain   *string   `json:"subdomain,omitempty" db:"subdomain"`
	CompanyName *string   `json:"company_name,omitempty" db:"company_name"`
	Phone       *string   `json:"phone,omitempty" db:"phone"`
	Address     *string   `json:"address,omitempty" db:"address"`
	Timezone    string    `json:"timezone" db:"timezone"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Status      string    `json:"status" db:"status"`

	// Limits
	MaxMikrotiks    int `json:"max_mikrotiks" db:"max_mikrotiks"`
	MaxNetworkUsers int `json:"max_network_users" db:"max_network_users"`
	MaxStaffUsers   int `json:"max_staff_users" db:"max_staff_users"`

	// Features & Metadata
	Features JSONB `json:"features" db:"features"`
	Metadata JSONB `json:"metadata,omitempty" db:"metadata"`

	// Audit
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty" db:"suspended_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TenantInput represents input for creating/updating tenant
type TenantInput struct {
	Name        string  `json:"name" binding:"required,min=3,max=200"`
	Subdomain   *string `json:"subdomain,omitempty" binding:"omitempty,min=3,max=100,alphanum"`
	CompanyName *string `json:"company_name,omitempty" binding:"omitempty,max=200"`
	Phone       *string `json:"phone,omitempty" binding:"omitempty,max=50"`
	Address     *string `json:"address,omitempty"`
	Timezone    *string `json:"timezone,omitempty" binding:"omitempty,timezone"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=active suspended inactive"`

	// Limits (only super admin can set)
	MaxMikrotiks    *int `json:"max_mikrotiks,omitempty" binding:"omitempty,min=1,max=100"`
	MaxNetworkUsers *int `json:"max_network_users,omitempty" binding:"omitempty,min=1,max=100000"`
	MaxStaffUsers   *int `json:"max_staff_users,omitempty" binding:"omitempty,min=1,max=1000"`

	// Features
	Features JSONB `json:"features,omitempty"`
	Metadata JSONB `json:"metadata,omitempty"`
}

// TenantStats represents usage statistics for a tenant
type TenantStats struct {
	TenantID            uuid.UUID `json:"tenant_id"`
	MikrotiksCount      int       `json:"mikrotiks_count"`
	NetworkUsersCount   int       `json:"network_users_count"`
	StaffUsersCount     int       `json:"staff_users_count"`
	ActiveSessionsCount int       `json:"active_sessions_count"`

	// Limits
	MaxMikrotiks    int `json:"max_mikrotiks"`
	MaxNetworkUsers int `json:"max_network_users"`
	MaxStaffUsers   int `json:"max_staff_users"`

	// Percentages
	MikrotiksUsagePercent    float64 `json:"mikrotiks_usage_percent"`
	NetworkUsersUsagePercent float64 `json:"network_users_usage_percent"`
	StaffUsersUsagePercent   float64 `json:"staff_users_usage_percent"`
}

// TenantFilter represents filter options for listing tenants
type TenantFilter struct {
	Status   *string `json:"status,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	Search   *string `json:"search,omitempty"` // Search in name, company_name

	// Pagination
	Limit  int `json:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `json:"offset" binding:"omitempty,min=0"`
}

// IsUnderLimit checks if tenant is under the specified limit
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

// HasFeature checks if tenant has a specific feature enabled
func (t *Tenant) HasFeature(feature string) bool {
	if t.Features == nil {
		return false
	}

	if enabled, ok := t.Features[feature].(bool); ok {
		return enabled
	}

	return false
}
