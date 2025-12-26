package entity

import "time"

// AgentBalance represents the agent_balances table
type AgentBalance struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID     string     `gorm:"column:agent_id;type:uuid;not null"`
	Balance     float64    `gorm:"column:balance;type:decimal(15,2);default:0.00"`
	LastUpdated time.Time  `gorm:"column:last_updated;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Agent *Agent `gorm:"foreignKey:AgentID"`
}

func (AgentBalance) TableName() string { return "agent_balances" }