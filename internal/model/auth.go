package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "superadmin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleTechnician UserRole = "technician"
	UserRoleSales      UserRole = "sales"
	UserRoleCS         UserRole = "cs"
	UserRoleFinance    UserRole = "finance"
	UserRoleViewer     UserRole = "viewer"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

type Permissions []string

func (p Permissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Permissions) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &p)
}

type Role struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	DisplayName string      `json:"display_name" db:"display_name"`
	Description string      `json:"description" db:"description"`
	Permissions Permissions `json:"permissions" db:"permissions"`
	IsSystem    bool        `json:"is_system" db:"is_system"`
	IsActive    bool        `json:"is_active" db:"is_active"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Username            string     `json:"username" db:"username"`
	Email               string     `json:"email" db:"email"`
	EncryptedPassword   string     `json:"-" db:"encrypted_password"`
	Fullname            string     `json:"fullname" db:"fullname"`
	Phone               *string    `json:"phone" db:"phone"`
	Avatar              *string    `json:"avatar" db:"avatar"`
	RoleID              *uuid.UUID `json:"role_id" db:"role_id"`
	UserRole            UserRole   `json:"user_role" db:"user_role"`
	Status              UserStatus `json:"status" db:"status"`
	LastLogin           *time.Time `json:"last_login" db:"last_login"`
	LastIP              *string    `json:"last_ip" db:"last_ip"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	PasswordChangedAt   *time.Time `json:"password_changed_at" db:"password_changed_at"`
	ForcePasswordChange bool       `json:"force_password_change" db:"force_password_change"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-" db:"two_factor_secret"`
	APIToken            *string    `json:"api_token" db:"api_token"`
	APITokenExpiresAt   *time.Time `json:"api_token_expires_at" db:"api_token_expires_at"`
	CreatedBy           *uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy           *uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

type UserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
	RoleID   string `json:"role_id"`
}

type UserFilter struct {
	IDs       []uuid.UUID `json:"ids"`
	Emails    []string    `json:"emails"`
	Usernames []string    `json:"usernames"`
}

func (f UserFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Emails) == 0 && len(f.Usernames) == 0
}
