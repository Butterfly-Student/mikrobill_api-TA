package model

import "time"


// LoginRequest for user authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	IP       string `json:"-"` // Set by handler from client IP
}

// CreateUserRequest for user registration
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Name     string   `json:"name" binding:"required"`
	Phone    string   `json:"phone"`
	Status   string   `json:"status"` // active, inactive, locked
	UserRole string   `json:"user_role"` // admin, manager, technician, viewer
	RoleIDs  []int64  `json:"role_ids"` // Custom role to assign (only first one used)
}

// ChangePasswordRequest for changing user password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}





// LoginResponse returned after successful authentication
type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      UserSummary `json:"user"`
}
