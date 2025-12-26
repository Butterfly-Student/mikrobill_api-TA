package entity

import "time"




// AgentNotification represents the agent_notifications table
type AgentNotification struct {
	ID               string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AgentID          string           `gorm:"column:agent_id;type:uuid;not null"`
	NotificationType NotificationType `gorm:"column:notification_type;type:notification_type;not null"`
	Title            string           `gorm:"column:title;type:varchar(255);not null"`
	Message          string           `gorm:"column:message;type:text;not null"`
	IsRead           bool             `gorm:"column:is_read;default:false"`
	CreatedAt        time.Time        `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Agent *Agent `gorm:"foreignKey:AgentID"`
}

func (AgentNotification) TableName() string { return "agent_notifications" }