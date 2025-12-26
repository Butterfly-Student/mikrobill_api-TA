// internal/dto/response/role_response.go
package model

// RoleResponse provides detailed role information
type RoleResponse struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	IsSystem    bool         `json:"is_system"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

// RoleSummary provides basic role information
type RoleSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// Permission structure for JSONB permissions field
// This matches the structure stored in roles.permissions column
type Permission struct {
	Resource string   `json:"resource"` // e.g., "users", "mikrotiks"
	Actions  []string `json:"actions"`  // e.g., ["read", "create", "update"]
	Path     string   `json:"path"`     // e.g., "/api/v1/users"
}

// PermissionSummary provides a flattened view of permissions
// Useful for UI display
type PermissionSummary struct {
	Resource    string `json:"resource"`
	Path        string `json:"path"`
	CanRead     bool   `json:"can_read"`
	CanCreate   bool   `json:"can_create"`
	CanUpdate   bool   `json:"can_update"`
	CanDelete   bool   `json:"can_delete"`
	CanSync     bool   `json:"can_sync"`
	CanTest     bool   `json:"can_test"`
	CanExport   bool   `json:"can_export"`
	CanManage   bool   `json:"can_manage"`
}

// RoleWithPermissionSummary provides role with flattened permissions
type RoleWithPermissionSummary struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name"`
	Description string              `json:"description"`
	Permissions []PermissionSummary `json:"permissions"`
	IsSystem    bool                `json:"is_system"`
	IsActive    bool                `json:"is_active"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

// Helper function to convert Permission to PermissionSummary
func (p Permission) ToSummary() PermissionSummary {
	summary := PermissionSummary{
		Resource: p.Resource,
		Path:     p.Path,
	}

	for _, action := range p.Actions {
		switch action {
		case "read":
			summary.CanRead = true
		case "create":
			summary.CanCreate = true
		case "update":
			summary.CanUpdate = true
		case "delete":
			summary.CanDelete = true
		case "sync":
			summary.CanSync = true
		case "test":
			summary.CanTest = true
		case "export":
			summary.CanExport = true
		case "manage":
			summary.CanManage = true
		}
	}

	return summary
}

// Helper function to convert RoleResponse to RoleWithPermissionSummary
func (r RoleResponse) ToPermissionSummary() RoleWithPermissionSummary {
	summaries := make([]PermissionSummary, len(r.Permissions))
	for i, perm := range r.Permissions {
		summaries[i] = perm.ToSummary()
	}

	return RoleWithPermissionSummary{
		ID:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		Permissions: summaries,
		IsSystem:    r.IsSystem,
		IsActive:    r.IsActive,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}


// CreateRoleRequest for creating new role
type CreateRoleRequest struct {
	Name        string       `json:"name" binding:"required,min=3,max=50"`
	DisplayName string       `json:"display_name" binding:"required"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// UpdateRoleRequest for updating role
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}
