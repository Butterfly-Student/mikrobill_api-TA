package entity

import "time"

// HotspotVoucher represents the hotspot_vouchers table
type HotspotVoucher struct {
	ID               string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID       string        `gorm:"column:mikrotik_id;type:uuid;not null"`
	Code             string        `gorm:"column:code;type:varchar(50);not null"`
	Username         string        `gorm:"column:username;type:varchar(100);not null"`
	Password         string        `gorm:"column:password;type:varchar(100);not null"`
	PackageID        *string       `gorm:"column:package_id;type:uuid"`
	ProfileID        *string       `gorm:"column:profile_id;type:uuid"`
	Price            float64       `gorm:"column:price;type:decimal(10,2);not null"`
	AgentPrice       float64       `gorm:"column:agent_price;type:decimal(10,2)"`
	Commission       float64       `gorm:"column:commission;type:decimal(10,2)"`
	ValidityDays     int           `gorm:"column:validity_days"`
	ValidityHours    int           `gorm:"column:validity_hours"`
	ValidFrom        *time.Time    `gorm:"column:valid_from;type:timestamptz"`
	ValidUntil       *time.Time    `gorm:"column:valid_until;type:timestamptz"`
	Status           VoucherStatus `gorm:"column:status;type:voucher_status;default:'active'"`
	ActivatedAt      *time.Time    `gorm:"column:activated_at;type:timestamptz"`
	ActivatedBy      string        `gorm:"column:activated_by;type:varchar(100)"`
	ExpiredAt        *time.Time    `gorm:"column:expired_at;type:timestamptz"`
	CreatedBy        *string       `gorm:"column:created_by;type:uuid"`
	AgentID          *string       `gorm:"column:agent_id;type:uuid"`
	SyncWithMikrotik bool          `gorm:"column:sync_with_mikrotik;default:true"`
	LastSync         *time.Time    `gorm:"column:last_sync;type:timestamptz"`
	MikrotikUserID   string        `gorm:"column:mikrotik_user_id;type:varchar(100)"`
	BatchID          *string       `gorm:"column:batch_id;type:uuid"`
	BatchName        string        `gorm:"column:batch_name;type:varchar(100)"`
	Notes            string        `gorm:"column:notes;type:text"`
	CreatedAt        time.Time     `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time     `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik     *Mikrotik        `gorm:"foreignKey:MikrotikID"`
	Package      *Package         `gorm:"foreignKey:PackageID"`
	Profile      *MikrotikProfile `gorm:"foreignKey:ProfileID"`
	CreatedByUser *User           `gorm:"foreignKey:CreatedBy"`
	Agent        *Agent           `gorm:"foreignKey:AgentID"`
	UsageLogs    []VoucherUsageLog `gorm:"foreignKey:VoucherID"`
}

func (HotspotVoucher) TableName() string { return "hotspot_vouchers" }
