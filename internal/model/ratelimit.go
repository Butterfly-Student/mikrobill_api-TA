// file: models/ratelimit.go
package model

import (
	"time"
)

// RateLimitRule - Rate limiting rules
type RateLimitRule struct {
	ID                 string   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name               string   `gorm:"type:varchar(100);unique;not null" json:"name"`
	Description        *string  `gorm:"type:text" json:"description,omitempty"`
	EndpointPattern    *string  `gorm:"type:varchar(200)" json:"endpoint_pattern,omitempty"`
	Method             *string  `gorm:"type:varchar(10)" json:"method,omitempty"`
	UserRole           *UserRole `gorm:"type:user_role" json:"user_role,omitempty"`
	RequestsPerMinute  int      `gorm:"not null;default:60" json:"requests_per_minute"`
	RequestsPerHour    int      `gorm:"not null;default:1000" json:"requests_per_hour"`
	RequestsPerDay     int      `gorm:"not null;default:10000" json:"requests_per_day"`
	IsActive           bool     `gorm:"default:true" json:"is_active"`
	Priority           int      `gorm:"default:0" json:"priority"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (RateLimitRule) TableName() string {
	return "rate_limit_rules"
}

// RateLimitViolation - Rate limit violation logs
type RateLimitViolation struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        *string   `gorm:"type:uuid;index" json:"user_id,omitempty"`
	IPAddress     string    `gorm:"type:inet;not null" json:"ip_address"`
	Endpoint      string    `gorm:"type:varchar(200);not null" json:"endpoint"`
	Method        *string   `gorm:"type:varchar(10)" json:"method,omitempty"`
	RuleName      *string   `gorm:"type:varchar(100)" json:"rule_name,omitempty"`
	ExceededLimit string    `gorm:"type:varchar(50)" json:"exceeded_limit"`
	RequestCount  int       `json:"request_count"`
	UserAgent     *string   `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// Relation
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (RateLimitViolation) TableName() string {
	return "rate_limit_violations"
}

// ============================================================================
// RATE LIMIT REQUEST/RESPONSE MODELS
// ============================================================================

type CreateRateLimitRuleRequest struct {
	Name               string   `json:"name" binding:"required"`
	Description        *string  `json:"description,omitempty"`
	EndpointPattern    *string  `json:"endpoint_pattern,omitempty"`
	Method             *string  `json:"method,omitempty" binding:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	UserRole           *UserRole `json:"user_role,omitempty"`
	RequestsPerMinute  int      `json:"requests_per_minute" binding:"required,min=1"`
	RequestsPerHour    int      `json:"requests_per_hour" binding:"required,min=1"`
	RequestsPerDay     int      `json:"requests_per_day" binding:"required,min=1"`
	IsActive           *bool    `json:"is_active,omitempty"`
	Priority           *int     `json:"priority,omitempty"`
}

type RateLimitRuleResponse struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       *string  `json:"description,omitempty"`
	EndpointPattern   *string  `json:"endpoint_pattern,omitempty"`
	Method            *string  `json:"method,omitempty"`
	UserRole          *UserRole `json:"user_role,omitempty"`
	RequestsPerMinute int      `json:"requests_per_minute"`
	RequestsPerHour   int      `json:"requests_per_hour"`
	RequestsPerDay    int      `json:"requests_per_day"`
	IsActive          bool     `json:"is_active"`
	Priority          int      `json:"priority"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RateLimitStats struct {
	TotalViolations24h int64 `json:"total_violations_24h"`
	UniqueIPs24h       int64 `json:"unique_ips_24h"`
	BlockedUsers24h    int64 `json:"blocked_users_24h"`
	MostViolatedRule   string `json:"most_violated_rule"`
}