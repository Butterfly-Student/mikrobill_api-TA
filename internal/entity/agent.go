package entity

import "time"

// Agent represents the agents table
type Agent struct {
	ID             string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username       string     `gorm:"column:username;type:varchar(100);not null"`
	Name           string     `gorm:"column:name;type:varchar(255);not null"`
	Phone          string     `gorm:"column:phone;type:varchar(20);not null"`
	Email          string     `gorm:"column:email;type:varchar(255)"`
	Password       string     `gorm:"column:password;type:varchar(255);not null"`
	Address        string     `gorm:"column:address;type:text"`
	Status         AgentStatus `gorm:"column:status;type:agent_status;default:'active'"`
	CommissionRate float64    `gorm:"column:commission_rate;type:decimal(5,2);default:5.00"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Balance           *AgentBalance            `gorm:"foreignKey:AgentID"`
	Transactions      []AgentTransaction       `gorm:"foreignKey:AgentID"`
	VoucherSales      []AgentVoucherSale       `gorm:"foreignKey:AgentID"`
	BalanceRequests   []AgentBalanceRequest    `gorm:"foreignKey:AgentID"`
	MonthlyPayments   []AgentMonthlyPayment    `gorm:"foreignKey:AgentID"`
	Payments          []AgentPayment           `gorm:"foreignKey:AgentID"`
	Notifications     []AgentNotification      `gorm:"foreignKey:AgentID"`
	HotspotVouchers   []HotspotVoucher         `gorm:"foreignKey:AgentID"`
}

func (Agent) TableName() string { return "agents" }