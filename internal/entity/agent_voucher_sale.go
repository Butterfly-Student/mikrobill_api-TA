package entity

import "time"

// AgentVoucherSale represents the agent_voucher_sales table
type AgentVoucherSale struct {
	ID              string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID         string        `gorm:"column:agent_id;type:uuid;not null"`
	MikrotikID      string        `gorm:"column:mikrotik_id;type:uuid;not null"`
	VoucherCode     string        `gorm:"column:voucher_code;type:varchar(100);not null"`
	PackageID       string        `gorm:"column:package_id;type:varchar(50);not null"`
	PackageName     string        `gorm:"column:package_name;type:varchar(100);not null"`
	CustomerPhone   string        `gorm:"column:customer_phone;type:varchar(20)"`
	CustomerName    string        `gorm:"column:customer_name;type:varchar(255)"`
	Price           float64       `gorm:"column:price;type:decimal(10,2);not null"`
	AgentPrice      float64       `gorm:"column:agent_price;type:decimal(10,2);default:0.00"`
	Commission      float64       `gorm:"column:commission;type:decimal(10,2);default:0.00"`
	CommissionAmount float64      `gorm:"column:commission_amount;type:decimal(10,2);default:0.00"`
	Status          VoucherStatus `gorm:"column:status;type:voucher_status;default:'active'"`
	SoldAt          time.Time     `gorm:"column:sold_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UsedAt          *time.Time    `gorm:"column:used_at;type:timestamptz"`
	Notes           string        `gorm:"column:notes;type:text"`

	// Relations
	Agent    *Agent    `gorm:"foreignKey:AgentID"`
	Mikrotik *Mikrotik `gorm:"foreignKey:MikrotikID"`
}

func (AgentVoucherSale) TableName() string { return "agent_voucher_sales" }