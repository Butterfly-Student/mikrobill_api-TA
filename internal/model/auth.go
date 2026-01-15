package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserRole Scan and Value methods for database compatibility
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		*r = UserRole(v)
		return nil
	case []byte:
		*r = UserRole(v)
		return nil
	default:
		return errors.New("incompatible type for UserRole")
	}
}

func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

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
	TenantID    *uuid.UUID  `json:"tenant_id" db:"tenant_id"`
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
	TenantID            *uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Username            string     `json:"username" db:"username"`
	Email               string     `json:"email" db:"email"`
	EncryptedPassword   string     `json:"-" db:"encrypted_password"`
	Fullname            string     `json:"fullname" db:"fullname"`
	Phone               *string    `json:"phone" db:"phone"`
	Avatar              *string    `json:"avatar" db:"avatar"`
	RoleID              *uuid.UUID `json:"role_id" db:"role_id"`
	UserRole            UserRole   `json:"user_role" db:"user_role"`
	Status              UserStatus `json:"status" db:"status"`
	IsSuperadmin        bool       `json:"is_superadmin" db:"is_superadmin"`
	LastLoginAt         *time.Time `json:"last_login_at" db:"last_login_at"`
	LastIP              *string    `json:"last_ip" db:"last_ip"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	PasswordChangedAt   *time.Time `json:"password_changed_at" db:"password_changed_at"`
	ForcePasswordChange bool       `json:"force_password_change" db:"force_password_change"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-" db:"two_factor_secret"`
	CreatedBy           *uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy           *uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at" db:"deleted_at"`
}

type UserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
}

type UserFilter struct {
	IDs       []uuid.UUID `json:"ids"`
	Emails    []string    `json:"emails"`
	Usernames []string    `json:"usernames"`
}

func (f UserFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Emails) == 0 && len(f.Usernames) == 0
}
