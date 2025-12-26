package entity

import "time"

// AgentBalanceRequest represents the agent_balance_requests table
type AgentBalanceRequest struct {
	ID           string        `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID      string        `gorm:"column:agent_id;type:uuid;not null"`
	Amount       float64       `gorm:"column:amount;type:decimal(15,2);not null"`
	Status       RequestStatus `gorm:"column:status;type:request_status;default:'pending'"`
	AdminNotes   string        `gorm:"column:admin_notes;type:text"`
	RequestedAt  time.Time     `gorm:"column:requested_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	ProcessedAt  *time.Time    `gorm:"column:processed_at;type:timestamptz"`
	ProcessedBy  *string       `gorm:"column:processed_by;type:uuid"`

	// Relations
	Agent       *Agent `gorm:"foreignKey:AgentID"`
	ProcessedByUser *User `gorm:"foreignKey:ProcessedBy"`
}

func (AgentBalanceRequest) TableName() string { return "agent_balance_requests" }