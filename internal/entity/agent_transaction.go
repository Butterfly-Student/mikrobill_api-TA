package entity

import "time"

// AgentTransaction represents the agent_transactions table
type AgentTransaction struct {
	ID              string              `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID         string              `gorm:"column:agent_id;type:uuid;not null"`
	TransactionType AgentTransactionType `gorm:"column:transaction_type;type:agent_transaction_type;not null"`
	Amount          float64             `gorm:"column:amount;type:decimal(15,2);not null"`
	Description     string              `gorm:"column:description;type:text"`
	ReferenceID     string              `gorm:"column:reference_id;type:varchar(100)"`
	Status          TransactionStatus   `gorm:"column:status;type:transaction_status;default:'completed'"`
	CreatedAt       time.Time           `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Agent *Agent `gorm:"foreignKey:AgentID"`
}

func (AgentTransaction) TableName() string { return "agent_transactions" }