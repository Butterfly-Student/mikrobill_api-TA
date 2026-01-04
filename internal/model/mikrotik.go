package model

type CreateMikrotikRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port" binding:"required"`
	APIUsername string `json:"api_username" binding:"required"`
	APIPassword string `json:"api_password" binding:"required"`
	Keepalive   bool   `json:"keepalive"`
	Timeout     int    `json:"timeout"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpdateMikrotikRequest struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	APIUsername string `json:"api_username"`
	APIPassword string `json:"api_password"`
	Keepalive   *bool  `json:"keepalive"`
	Timeout     int    `json:"timeout"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// Mikrotik Active Management
type SetActiveMikrotikRequest struct {
	ID string `json:"id" binding:"required"`
}

type MikrotikResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	APIUsername string `json:"api_username"`
	Keepalive   bool   `json:"keepalive"`
	Timeout     int    `json:"timeout"`
	Location    string `json:"location"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	Status      string `json:"status"`
	Version     string `json:"version"`
	Uptime      string `json:"uptime"`
	LastSync    string `json:"last_sync"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type MikrotikStatusResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	IsActive bool   `json:"is_active"`
	Status   string `json:"status"`
}

type PaginationRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Search   string `form:"search"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir" binding:"oneof=asc desc"`
}

type PaginationResponse struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}
