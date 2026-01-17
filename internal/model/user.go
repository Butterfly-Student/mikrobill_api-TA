// file: models/user.go
package model

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// ENUMS
// ============================================================================

type UserRole string
type UserStatus string

const (
	UserRoleSuperAdmin UserRole = "superadmin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleTechnician UserRole = "technician"
	UserRoleViewer     UserRole = "viewer"

	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

// ============================================================================
// USER MODELS (Database)
// ============================================================================

// User - System user model
type User struct {
	ID                string  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID          *string `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	Username          string  `gorm:"type:varchar(50);uniqueIndex:idx_users_username;not null" json:"username"`
	Email             string  `gorm:"type:varchar(255);uniqueIndex:idx_users_email;not null" json:"email"`
	EncryptedPassword string  `gorm:"type:text;not null" json:"-"`
	Fullname          string  `gorm:"type:varchar(255);not null" json:"fullname"`
	Phone             *string `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Avatar            *string `gorm:"type:text" json:"avatar,omitempty"`

	RoleID   *string    `gorm:"type:uuid;index" json:"role_id,omitempty"`
	UserRole UserRole   `gorm:"type:user_role;default:'viewer'" json:"user_role"`
	Status   UserStatus `gorm:"type:user_status;default:'active'" json:"status"`

	// Security
	IsSuperadmin        bool       `gorm:"default:false" json:"is_superadmin"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastIP              *string    `gorm:"type:inet" json:"last_ip,omitempty"`
	FailedLoginAttempts int        `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	PasswordChangedAt   *time.Time `json:"password_changed_at,omitempty"`
	ForcePasswordChange bool       `gorm:"default:false" json:"force_password_change"`
	TwoFactorEnabled    bool       `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret     *string    `gorm:"type:text" json:"-"`

	// Audit
	CreatedBy *string        `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy *string        `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant         *Tenant         `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Role           *Role           `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Sessions       []UserSession   `gorm:"foreignKey:UserID" json:"-"`
	TenantUsers    []TenantUser    `gorm:"foreignKey:UserID" json:"-"`
	ActivityLogs   []ActivityLog   `gorm:"foreignKey:UserID" json:"-"`
	SuperadminLogs []SuperadminLog `gorm:"foreignKey:SuperadminID" json:"-"`
	CreatedByUser  *User           `gorm:"foreignKey:CreatedBy" json:"created_by_user,omitempty"`
	UpdatedByUser  *User           `gorm:"foreignKey:UpdatedBy" json:"updated_by_user,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// ============================================================================
// USER REQUEST/RESPONSE MODELS
// ============================================================================

// CreateUserRequest - Request for creating a user
type CreateUserRequest struct {
	Username string      `json:"username" binding:"required,min=3,max=50"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=8"`
	Fullname string      `json:"fullname" binding:"required"`
	Phone    *string     `json:"phone,omitempty"`
	UserRole *UserRole   `json:"user_role,omitempty"`
	Status   *UserStatus `json:"status,omitempty"`
	RoleID   *string     `json:"role_id,omitempty"`
}

// UpdateUserRequest - Request for updating a user
type UpdateUserRequest struct {
	Fullname *string     `json:"fullname,omitempty"`
	Phone    *string     `json:"phone,omitempty"`
	Avatar   *string     `json:"avatar,omitempty"`
	Status   *UserStatus `json:"status,omitempty"`
	UserRole *UserRole   `json:"user_role,omitempty"`
	RoleID   *string     `json:"role_id,omitempty"`
}

// UserResponse - Response for user data
type UserResponse struct {
	ID                  string     `json:"id"`
	TenantID            *string    `json:"tenant_id,omitempty"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	Fullname            string     `json:"fullname"`
	Phone               *string    `json:"phone,omitempty"`
	Avatar              *string    `json:"avatar,omitempty"`
	RoleID              *string    `json:"role_id,omitempty"`
	UserRole            UserRole   `json:"user_role"`
	Status              UserStatus `json:"status"`
	IsSuperadmin        bool       `json:"is_superadmin"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastIP              *string    `json:"last_ip,omitempty"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	ForcePasswordChange bool       `json:"force_password_change"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Relations (optional)
	Tenant *TenantResponse `json:"tenant,omitempty"`
	Role   *RoleResponse   `json:"role,omitempty"`
}

// UserListResponse - Response for list of users
type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// UserStats - User statistics
type UserStats struct {
	TotalUsers     int64 `json:"total_users"`
	ActiveUsers    int64 `json:"active_users"`
	InactiveUsers  int64 `json:"inactive_users"`
	SuspendedUsers int64 `json:"suspended_users"`
	BannedUsers    int64 `json:"banned_users"`
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:                  u.ID,
		TenantID:            u.TenantID,
		Username:            u.Username,
		Email:               u.Email,
		Fullname:            u.Fullname,
		Phone:               u.Phone,
		Avatar:              u.Avatar,
		RoleID:              u.RoleID,
		UserRole:            u.UserRole,
		Status:              u.Status,
		IsSuperadmin:        u.IsSuperadmin,
		LastLoginAt:         u.LastLoginAt,
		LastIP:              u.LastIP,
		TwoFactorEnabled:    u.TwoFactorEnabled,
		ForcePasswordChange: u.ForcePasswordChange,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}
}

// IsLocked checks if user is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanLogin checks if user can attempt login
func (u *User) CanLogin() bool {
	if u.Status == UserStatusBanned || u.Status == UserStatusSuspended {
		return false
	}
	return !u.IsLocked()
}
