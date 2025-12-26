package entity

import "time"

// Package represents the packages table


type ServiceType string


const (
	ServiceTypePPPoE    ServiceType = "pppoe"
	ServiceTypeHotspot  ServiceType = "hotspot"
	ServiceTypeStaticIP ServiceType = "static_ip"
)

type Package struct {
	MikrotikID      string      `gorm:"column:mikrotik_id;type:uuid;not null"`
	Name            string      `gorm:"column:name;type:varchar(100);not null"`
	DisplayName     string      `gorm:"column:display_name;type:varchar(150)"`
	Description     string      `gorm:"column:description;type:text"`
	Speed           string      `gorm:"column:speed;type:varchar(50);not null"`
	BandwidthUp     string      `gorm:"column:bandwidth_up;type:varchar(50)"`
	BandwidthDown   string      `gorm:"column:bandwidth_down;type:varchar(50)"`
	Price           float64     `gorm:"column:price;type:decimal(10,2);not null"`
	TaxRate         float64     `gorm:"column:tax_rate;type:decimal(5,2);default:11.00"`
	ServiceType     ServiceType `gorm:"column:service_type;type:service_type;not null"`
	ProfileID       *string     `gorm:"column:profile_id;type:uuid"`
	StaticIPPool    string      `gorm:"column:static_ip_pool;type:varchar(50)"`
	FUPEnabled      bool        `gorm:"column:fup_enabled;default:false"`
	FUPQuota        int         `gorm:"column:fup_quota"`
	FUPSpeed        string      `gorm:"column:fup_speed;type:varchar(50)"`
	ValidityDays    int         `gorm:"column:validity_days"`
	ValidityHours   int         `gorm:"column:validity_hours"`
	ImageFilename   string      `gorm:"column:image_filename;type:varchar(255)"`
	IsFeatured      bool        `gorm:"column:is_featured;default:false"`
	IsPopular       bool        `gorm:"column:is_popular;default:false"`
	SortOrder       int         `gorm:"column:sort_order;default:0"`
	Category        string      `gorm:"column:category;type:varchar(50)"`
	Tags            []string    `gorm:"column:tags;type:text[]"`
	IsActive        bool        `gorm:"column:is_active;default:true"`
	CreatedAt       time.Time   `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time   `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik *Mikrotik       `gorm:"foreignKey:MikrotikID"`
	Profile  *MikrotikProfile `gorm:"foreignKey:ProfileID"`
	Customers []Customer     `gorm:"foreignKey:PackageID"`
	Invoices  []Invoice      `gorm:"foreignKey:PackageID"`
	Vouchers  []HotspotVoucher `gorm:"foreignKey:PackageID"`
}

func (Package) TableName() string { return "packages" }