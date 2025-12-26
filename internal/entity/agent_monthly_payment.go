package entity

import "time"

// AgentMonthlyPayment represents the agent_monthly_payments table
type AgentMonthlyPayment struct {
	ID               string              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID          string              `gorm:"column:agent_id;type:uuid;not null"`
	CustomerID       string              `gorm:"column:customer_id;type:uuid;not null"`
	InvoiceID        string              `gorm:"column:invoice_id;type:uuid;not null"`
	PaymentAmount    float64             `gorm:"column:payment_amount;type:decimal(15,2);not null"`
	CommissionAmount float64             `gorm:"column:commission_amount;type:decimal(15,2);default:0.00"`
	PaymentMethod    string              `gorm:"column:payment_method;type:varchar(50);default:'cash'"`
	Notes            string              `gorm:"column:notes;type:text"`
	Status           TransactionStatus   `gorm:"column:status;type:transaction_status;default:'completed'"`
	PaidAt           time.Time           `gorm:"column:paid_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Agent    *Agent    `gorm:"foreignKey:AgentID"`
	Customer *Customer `gorm:"foreignKey:CustomerID"`
	Invoice  *Invoice  `gorm:"foreignKey:InvoiceID"`
}

func (AgentMonthlyPayment) TableName() string { return "agent_monthly_payments" }