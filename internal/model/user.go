package model


// UserResponse represents user data in response
type UserResponse struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Status    string         `json:"status"`
	Roles     []RoleSummary  `json:"roles"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}


// UserSummary berisi informasi ringkas pengguna yang dikembalikan dalam respons autentikasi atau profil.
//
// @Description Digunakan sebagai bagian dari respons login dan profil pengguna.
type UserSummary struct {
	// ID unik pengguna dalam sistem
	ID int64 `json:"id" example:"123"`
	// Nama lengkap pengguna
	Name string `json:"name" example:"Admin User"`
	// Email pengguna yang terdaftar
	Email string `json:"email" example:"admin@example.com"`
	// Status akun pengguna (misal: "active", "inactive")
	Status string `json:"status" example:"active"`
	// Daftar peran (roles) yang dimiliki pengguna
	Roles []string `json:"roles" example:"[\"admin\",\"user\"]"`
}

// UpdateUserRequest represents user update payload
type UpdateUserRequest struct {
	Name    string  `json:"name"`
	Email   string  `json:"email" binding:"omitempty,email"`
	Status  string  `json:"status"`
	RoleIDs []int64 `json:"role_ids"`
}


// AssignRolesRequest represents role assignment payload
type AssignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids" binding:"required" example:"1,2,3"`
}