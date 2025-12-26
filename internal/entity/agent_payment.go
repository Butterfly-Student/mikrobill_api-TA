package entity

import "time"


// AgentPayment represents the agent_payments table
type AgentPayment struct {
	ID            string            `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID       string            `gorm:"column:agent_id;type:uuid;not null"`
	CustomerID    string            `gorm:"column:customer_id;type:uuid;not null"`
	InvoiceID     string            `gorm:"column:invoice_id;type:uuid;not null"`
	Amount        float64           `gorm:"column:amount;type:decimal(15,2);not null"`
	PaymentMethod string            `gorm:"column:payment_method;type:varchar(50);default:'cash'"`
	Notes         string            `gorm:"column:notes;type:text"`
	Status        TransactionStatus `gorm:"column:status;type:transaction_status;default:'completed'"`
	PaidAt        time.Time         `gorm:"column:paid_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Agent    *Agent    `gorm:"foreignKey:AgentID"`
	Customer *Customer `gorm:"foreignKey:CustomerID"`
	Invoice  *Invoice  `gorm:"foreignKey:InvoiceID"`
}

func (AgentPayment) TableName() string { return "agent_payments" }