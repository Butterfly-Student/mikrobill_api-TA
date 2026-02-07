package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	IPAllocationStatusActive   = "active"
	IPAllocationStatusInactive = "inactive"
	IPAllocationStatusReserved = "reserved"
)

type IPAMPool struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" db:"id"`
	TenantID    string    `json:"tenant_id" gorm:"index:tenant_id;not null" db:"tenant_id"`
	Name        string    `json:"name" gorm:"not null" db:"name"`
	Description string    `json:"description" gorm:"description"`
	PoolType    string    `json:"pool_type" gorm:"not null" db:"pool_type"` // "public", "private", "pppoe", "hotspot"
	StartIP     string    `json:"start_ip" gorm:"start_ip" db:"start_ip"`
	EndIP       string    `json:"end_ip" gorm:"end_ip" db:"end_ip"`
	Network     string    `json:"network" gorm:"network" db:"network"` // CIDR notation: "192.168.1.0/24"
	Netmask     string    `json:"netmask" gorm:"netmask" db:"netmask"`
	Gateway     string    `json:"gateway" gorm:"gateway" db:"gateway"`
	TotalIPs    int       `json:"total_ips" gorm:"total_ips"`
	UsedIPs     int       `json:"used_ips" gorm:"used_ips"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

type IPAllocation struct {
	ID             uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" db:"id"`
	TenantID       string     `json:"tenant_id" gorm:"index:tenant_id,customer_id;not null" db:"tenant_id"`
	CustomerID     string     `json:"customer_id" gorm:"index:customer_id;not null" db:"customer_id"`
	PoolID         uuid.UUID  `json:"pool_id" gorm:"index:pool_id;not null" db:"pool_id"`
	IPAddress      string     `json:"ip_address" gorm:"not null;uniqueIndex:tenant_id,pool_id,ip_address" db:"ip_address"`
	MacAddress     *string    `json:"mac_address" db:"mac_address"`
	Hostname       *string    `json:"hostname" db:"hostname"`
	Status         string     `json:"status" gorm:"status"` // "active", "inactive", "reserved"
	LeaseStartTime *time.Time `json:"lease_start_time" gorm:"lease_start_time"`
	LeaseEndTime   *time.Time `json:"lease_end_time" gorm:"lease_end_time"`
	AllocatedAt    time.Time  `json:"allocated_at" gorm:"autoCreateTime" db:"allocated_at"`
	ReleasedAt     *time.Time `json:"released_at" gorm:"released_at"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt      time.Time  `json:"autoUpdateTime" gorm:"autoUpdateTime" db:"updated_at"`
}

type IPAllocationRequest struct {
	PoolID        uuid.UUID `json:"pool_id"`
	CustomerID    string    `json:"customer_id"`
	IPAddress     string    `json:"ip_address,omitempty"`
	PreferredIP   string    `json:"preferred_ip"`
	MacAddress    *string   `json:"mac_address,omitempty"`
	Hostname      *string   `json:"hostname,omitempty"`
	LeaseDuration int       `json:"lease_duration,omitempty"` // days
}

type IPAMStats struct {
	TotalPools       int     `json:"total_pools"`
	TotalIPs         int     `json:"total_ips"`
	UsedIPs          int     `json:"used_ips"`
	AvailableIPs     int     `json:"available_ips"`
	ReservedIPs      int     `json:"reserved_ips"`
	AllocationRate   float64 `json:"allocation_rate"` // allocations per day
	AverageLeaseDays float64 `json:"average_lease_days"`
}

type IPAMConfig struct {
	DefaultLeaseDays     int  `json:"default_lease_days"`
	EnableAutoCleanup    bool `json:"enable_auto_cleanup"`
	CleanupIntervalHours int  `json:"cleanup_interval_hours"`
	MaxInactiveDays      int  `json:"max_inactive_days"`
	EnableMonitoring     bool `json:"enable_monitoring"`
}

type IPAMPoolUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Network     string `json:"network"`
	Netmask     string `json:"netmask"`
	Gateway     string `json:"gateway"`
}

type IPAllocationFilter struct {
	PoolID   *uuid.UUID `json:"pool_id"`
	Status   string     `json:"status"`
	TenantID *string    `json:"tenant_id"`
}

type IPUsage struct {
	CustomerID    string    `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	IPAddress     string    `json:"ip_address"`
	BytesUpload   uint64    `json:"bytes_upload" db:"bytes_upload"`
	BytesDownload uint64    `json:"bytes_download" db:"bytes_download"`
	LastActivity  time.Time `json:"last_activity" db:"last_activity"`
}
