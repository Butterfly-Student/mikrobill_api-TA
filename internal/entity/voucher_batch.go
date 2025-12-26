package entity

import "time"

// VoucherBatch represents the voucher_batches table
type VoucherBatch struct {
	ID               string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID       string     `gorm:"column:mikrotik_id;type:uuid;not null"`
	BatchName        string     `gorm:"column:batch_name;type:varchar(100);not null"`
	PackageID        *string    `gorm:"column:package_id;type:uuid"`
	ProfileID        *string    `gorm:"column:profile_id;type:uuid"`
	TotalVouchers    int        `gorm:"column:total_vouchers;not null"`
	PricePerVoucher  float64    `gorm:"column:price_per_voucher;type:decimal(10,2);not null"`
	CodePrefix       string     `gorm:"column:code_prefix;type:varchar(20)"`
	CodeSuffix       string     `gorm:"column:code_suffix;type:varchar(20)"`
	CodeLength       int        `gorm:"column:code_length;default:8"`
	CreatedBy        *string    `gorm:"column:created_by;type:uuid"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik     *Mikrotik        `gorm:"foreignKey:MikrotikID"`
	Package      *Package         `gorm:"foreignKey:PackageID"`
	Profile      *MikrotikProfile `gorm:"foreignKey:ProfileID"`
	CreatedByUser *User           `gorm:"foreignKey:CreatedBy"`
}

func (VoucherBatch) TableName() string { return "voucher_batches" }