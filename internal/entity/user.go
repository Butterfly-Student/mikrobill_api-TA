package entity

import "time"

// User Role
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
    UserStatusActive     UserStatus = "active"
    UserStatusInactive   UserStatus = "inactive"
    UserStatusSuspended  UserStatus = "suspended"
    UserStatusLocked     UserStatus = "locked"
)

// User represents the users table
type User struct {
    ID                    int64           `gorm:"primaryKey;column:id"`
    Username              string          `gorm:"column:username;type:varchar(50);unique;not null"`
    Email                 string          `gorm:"column:email;type:varchar(255);unique;not null"`
    EncryptedPassword     string          `gorm:"column:encrypted_password;type:text;not null"`
    Fullname              string          `gorm:"column:fullname;type:varchar(255);not null"`
    Phone                 string          `gorm:"column:phone;type:varchar(20)"`
    Avatar                string          `gorm:"column:avatar;type:text"`
    RoleID                *int64          `gorm:"column:role_id"`
    UserRole              UserRole        `gorm:"column:user_role;type:user_role;not null;default:'viewer'"`
    Status                UserStatus      `gorm:"column:status;type:user_status;not null;default:'active'"`
    LastLogin             *time.Time      `gorm:"column:last_login;type:timestamptz"`
    LastIP                string          `gorm:"column:last_ip;type:inet"`
    FailedLoginAttempts   int             `gorm:"column:failed_login_attempts;default:0"`
    LockedUntil           *time.Time      `gorm:"column:locked_until;type:timestamptz"`
    PasswordChangedAt     *time.Time      `gorm:"column:password_changed_at;type:timestamptz"`
    ForcePasswordChange   bool            `gorm:"column:force_password_change;default:false"`
    TwoFactorEnabled      bool            `gorm:"column:two_factor_enabled;default:false"`
    TwoFactorSecret       string          `gorm:"column:two_factor_secret;type:text"`
    APIToken              string          `gorm:"column:api_token;type:text;unique"`
    APITokenExpiresAt     *time.Time      `gorm:"column:api_token_expires_at;type:timestamptz"`
    CreatedBy             *int64          `gorm:"column:created_by"`
    UpdatedBy             *int64          `gorm:"column:updated_by"`
    CreatedAt             time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt             time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`

    // Relations
    Role                     *Role                 `gorm:"foreignKey:RoleID"`
    CreatedByUser            *User                 `gorm:"foreignKey:CreatedBy"`
    UpdatedByUser            *User                 `gorm:"foreignKey:UpdatedBy"`
}

func (User) TableName() string { return "users" }