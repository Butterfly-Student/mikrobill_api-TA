package model

// UserRole represents the role type in the system
type UserRole string

const (
	RoleSuperAdmin       UserRole = "SUPER_ADMIN"
	RoleTenantOwner      UserRole = "TENANT_OWNER"
	RoleTenantAdmin      UserRole = "TENANT_ADMIN"
	RoleTenantTechnician UserRole = "TENANT_TECHNICIAN"
	RoleTenantViewer     UserRole = "TENANT_VIEWER"
)

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleTenantOwner, RoleTenantAdmin, RoleTenantTechnician, RoleTenantViewer:
		return true
	}
	return false
}

// IsSuperAdmin checks if the role is super admin
func (r UserRole) IsSuperAdmin() bool {
	return r == RoleSuperAdmin
}

// CanAccessAllTenants checks if the role can access all tenants
func (r UserRole) CanAccessAllTenants() bool {
	return r == RoleSuperAdmin
}

// CanManageTenant checks if the role can manage tenant settings
func (r UserRole) CanManageTenant() bool {
	return r == RoleSuperAdmin || r == RoleTenantOwner || r == RoleTenantAdmin
}

// CanManageUsers checks if the role can manage users
func (r UserRole) CanManageUsers() bool {
	return r == RoleSuperAdmin || r == RoleTenantOwner || r == RoleTenantAdmin
}

// CanWrite checks if the role has write permissions
func (r UserRole) CanWrite() bool {
	return r != RoleTenantViewer
}

// IsReadOnly checks if the role is read-only
func (r UserRole) IsReadOnly() bool {
	return r == RoleTenantViewer
}

// HasPermission checks if the role has a specific permission
func (r UserRole) HasPermission(resource string, action string) bool {
	// Super admin has all permissions
	if r.IsSuperAdmin() {
		return true
	}

	// Viewer can only read
	if r.IsReadOnly() {
		return action == "read" || action == "list"
	}

	// Define permission matrix
	permissions := map[UserRole]map[string][]string{
		RoleTenantOwner: {
			"tenant":   {"read", "update", "delete"},
			"user":     {"create", "read", "update", "delete", "list"},
			"role":     {"create", "read", "update", "delete", "list"},
			"customer": {"create", "read", "update", "delete", "list"},
			"mikrotik": {"create", "read", "update", "delete", "list"},
			"profile":  {"create", "read", "update", "delete", "list"},
			"report":   {"read", "list", "export"},
			"billing":  {"create", "read", "update", "delete", "list"},
		},
		RoleTenantAdmin: {
			"user":     {"create", "read", "update", "list"},
			"customer": {"create", "read", "update", "delete", "list"},
			"mikrotik": {"create", "read", "update", "list"},
			"profile":  {"create", "read", "update", "delete", "list"},
			"report":   {"read", "list", "export"},
			"billing":  {"read", "list"},
		},
		RoleTenantTechnician: {
			"customer": {"read", "update", "list"},
			"mikrotik": {"read", "list"},
			"profile":  {"read", "list"},
			"report":   {"read", "list"},
		},
	}

	if rolePerms, ok := permissions[r]; ok {
		if actions, ok := rolePerms[resource]; ok {
			for _, a := range actions {
				if a == action {
					return true
				}
			}
		}
	}

	return false
}

// RoleHierarchy returns the hierarchy level of the role (higher = more privileged)
func (r UserRole) Hierarchy() int {
	switch r {
	case RoleSuperAdmin:
		return 100
	case RoleTenantOwner:
		return 80
	case RoleTenantAdmin:
		return 60
	case RoleTenantTechnician:
		return 40
	case RoleTenantViewer:
		return 20
	default:
		return 0
	}
}

// CanManageRole checks if this role can manage another role
func (r UserRole) CanManageRole(targetRole UserRole) bool {
	// Super admin can manage all roles
	if r.IsSuperAdmin() {
		return true
	}

	// Can only manage roles with lower hierarchy
	return r.Hierarchy() > targetRole.Hierarchy()
}
