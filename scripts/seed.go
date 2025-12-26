// scripts/seed.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mikrobill/config"
	"mikrobill/internal/entity"
	database "mikrobill/internal/infrastructure/db/postgres"
	"mikrobill/internal/port/service"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

// Permission structure untuk JSONB
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Path     string `json:"path"`
}

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	fmt.Println("Starting RBAC seeding...")

	// === 1. Define all permissions ===
	allPermissions := []Permission{
		// Users Management
		{Resource: "users", Action: "read", Path: "/api/v1/users"},
		{Resource: "users", Action: "create", Path: "/api/v1/users"},
		{Resource: "users", Action: "read", Path: "/api/v1/users/:id"},
		{Resource: "users", Action: "update", Path: "/api/v1/users/:id"},
		{Resource: "users", Action: "delete", Path: "/api/v1/users/:id"},

		// Roles Management
		{Resource: "roles", Action: "read", Path: "/api/v1/roles"},
		{Resource: "roles", Action: "create", Path: "/api/v1/roles"},
		{Resource: "roles", Action: "read", Path: "/api/v1/roles/:id"},
		{Resource: "roles", Action: "update", Path: "/api/v1/roles/:id"},
		{Resource: "roles", Action: "delete", Path: "/api/v1/roles/:id"},

		// Mikrotik Management
		{Resource: "mikrotiks", Action: "read", Path: "/api/v1/mikrotiks"},
		{Resource: "mikrotiks", Action: "create", Path: "/api/v1/mikrotiks"},
		{Resource: "mikrotiks", Action: "read", Path: "/api/v1/mikrotiks/:id"},
		{Resource: "mikrotiks", Action: "update", Path: "/api/v1/mikrotiks/:id"},
		{Resource: "mikrotiks", Action: "delete", Path: "/api/v1/mikrotiks/:id"},
		{Resource: "mikrotiks", Action: "test", Path: "/api/v1/mikrotiks/:id/test-connection"},

		// Customer Management
		{Resource: "customers", Action: "read", Path: "/api/v1/customers"},
		{Resource: "customers", Action: "create", Path: "/api/v1/customers"},
		{Resource: "customers", Action: "read", Path: "/api/v1/customers/:id"},
		{Resource: "customers", Action: "update", Path: "/api/v1/customers/:id"},
		{Resource: "customers", Action: "delete", Path: "/api/v1/customers/:id"},

		// Package Management
		{Resource: "packages", Action: "read", Path: "/api/v1/packages"},
		{Resource: "packages", Action: "create", Path: "/api/v1/packages"},
		{Resource: "packages", Action: "read", Path: "/api/v1/packages/:id"},
		{Resource: "packages", Action: "update", Path: "/api/v1/packages/:id"},
		{Resource: "packages", Action: "delete", Path: "/api/v1/packages/:id"},

		// PPPoE Accounts
		{Resource: "pppoe_accounts", Action: "read", Path: "/api/v1/pppoe_accounts"},
		{Resource: "pppoe_accounts", Action: "create", Path: "/api/v1/pppoe_accounts"},
		{Resource: "pppoe_accounts", Action: "read", Path: "/api/v1/pppoe_accounts/:id"},
		{Resource: "pppoe_accounts", Action: "update", Path: "/api/v1/pppoe_accounts/:id"},
		{Resource: "pppoe_accounts", Action: "delete", Path: "/api/v1/pppoe_accounts/:id"},

		// Hotspot Accounts
		{Resource: "hotspot_accounts", Action: "read", Path: "/api/v1/hotspot_accounts"},
		{Resource: "hotspot_accounts", Action: "create", Path: "/api/v1/hotspot_accounts"},
		{Resource: "hotspot_accounts", Action: "read", Path: "/api/v1/hotspot_accounts/:id"},
		{Resource: "hotspot_accounts", Action: "update", Path: "/api/v1/hotspot_accounts/:id"},
		{Resource: "hotspot_accounts", Action: "delete", Path: "/api/v1/hotspot_accounts/:id"},

		// Invoice Management
		{Resource: "invoices", Action: "read", Path: "/api/v1/invoices"},
		{Resource: "invoices", Action: "create", Path: "/api/v1/invoices"},
		{Resource: "invoices", Action: "read", Path: "/api/v1/invoices/:id"},
		{Resource: "invoices", Action: "update", Path: "/api/v1/invoices/:id"},

		// Payment Management
		{Resource: "payments", Action: "read", Path: "/api/v1/payments"},
		{Resource: "payments", Action: "create", Path: "/api/v1/payments"},
		{Resource: "payments", Action: "verify", Path: "/api/v1/payments/:id/verify"},

		// Reports
		{Resource: "reports", Action: "read", Path: "/api/v1/reports"},
		{Resource: "reports", Action: "export", Path: "/api/v1/reports/export"},
	}

	// === 2. Filter permissions by resource ===
	filterByResources := func(perms []Permission, resources []string) []Permission {
		allowed := make(map[string]bool)
		for _, r := range resources {
			allowed[r] = true
		}
		var result []Permission
		for _, p := range perms {
			if allowed[p.Resource] {
				result = append(result, p)
			}
		}
		return result
	}

	filterByActions := func(perms []Permission, actions []string) []Permission {
		allowed := make(map[string]bool)
		for _, a := range actions {
			allowed[a] = true
		}
		var result []Permission
		for _, p := range perms {
			if allowed[p.Action] {
				result = append(result, p)
			}
		}
		return result
	}

	// === 3. Define role permissions ===
	rolePermissions := map[string][]Permission{
		// Superadmin: Full access
		"superadmin": allPermissions,

		// Admin: Full access to all resources
		"admin": allPermissions,

		// Technician: Manage mikrotik, packages, accounts
		"technician": filterByActions(
			filterByResources(allPermissions, []string{
				"mikrotiks", "packages", "pppoe_accounts", "hotspot_accounts", "customers",
			}),
			[]string{"read", "create", "update", "test"},
		),

		// Sales: Manage customers and packages
		"sales": filterByActions(
			filterByResources(allPermissions, []string{
				"customers", "packages", "invoices",
			}),
			[]string{"read", "create", "update"},
		),

		// CS: Read customers, accounts, help with issues
		"cs": filterByActions(
			filterByResources(allPermissions, []string{
				"customers", "pppoe_accounts", "hotspot_accounts", "packages",
			}),
			[]string{"read"},
		),

		// Finance: Manage invoices and payments
		"finance": filterByActions(
			filterByResources(allPermissions, []string{
				"customers", "invoices", "payments", "reports",
			}),
			[]string{"read", "create", "verify", "export"},
		),

		// Viewer: Read-only access
		"viewer": filterByActions(allPermissions, []string{"read"}),
	}

	// === 4. Update roles with permissions ===
	for roleName, perms := range rolePermissions {
		var role entity.Role
		result := db.WithContext(ctx).Where("name = ?", roleName).First(&role)
		if result.Error != nil {
			log.Printf("Role not found: %s, skipping...", roleName)
			continue
		}

		// Convert permissions to JSON
		permsJSON, err := json.Marshal(perms)
		if err != nil {
			log.Printf("Failed to marshal permissions for role %s: %v", roleName, err)
			continue
		}

		role.Permissions = permsJSON
		if err := db.WithContext(ctx).Save(&role).Error; err != nil {
			log.Printf("Failed to update role %s: %v", roleName, err)
		} else {
			fmt.Printf("✓ Updated role: %s with %d permissions\n", role.DisplayName, len(perms))
		}
	}

	// === 5. Create Sample Users ===
	userData := []struct {
		username, fullname, email, pass, role string
	}{
		{"superadmin", "Super Administrator", "super@mikrobill.com", "super123", "superadmin"},
		{"admin", "Administrator", "admin@mikrobill.com", "admin123", "admin"},
		{"technician", "Technician User", "tech@mikrobill.com", "tech123", "technician"},
		{"sales", "Sales User", "sales@mikrobill.com", "sales123", "sales"},
		{"cs", "Customer Service", "cs@mikrobill.com", "cs123", "cs"},
		{"finance", "Finance User", "finance@mikrobill.com", "finance123", "finance"},
		{"viewer", "Viewer User", "viewer@mikrobill.com", "viewer123", "viewer"},
	}

	for _, u := range userData {
		// Get role
		var role entity.Role
		if err := db.Where("name = ?", u.role).First(&role).Error; err != nil {
			log.Printf("Role not found: %s", u.role)
			continue
		}

		// Hash password
		hash, err := service.NewPasswordService().Hash(u.pass)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.email, err)
			continue
		}

		// Check if user exists
		var existingUser entity.User
		err = db.WithContext(ctx).Where("email = ?", u.email).First(&existingUser).Error
		
		if err == nil {
			// User exists, update it
			existingUser.Username = u.username
			existingUser.EncryptedPassword = hash
			existingUser.Fullname = u.fullname
			existingUser.RoleID = &role.ID
			existingUser.UserRole = entity.UserRole(u.role)
			existingUser.Status = "active"
			
			if err := db.WithContext(ctx).Omit("last_ip", "api_token").Save(&existingUser).Error; err != nil {
				log.Printf("Failed to update user %s: %v", u.email, err)
			} else {
				fmt.Printf("✓ User updated: %s (%s) → %s\n", u.fullname, u.email, u.role)
			}
		} else {
			// User doesn't exist, create new
			user := entity.User{
				Username:          u.username,
				Email:             u.email,
				EncryptedPassword: hash,
				Fullname:          u.fullname,
				RoleID:            &role.ID,
				UserRole:          entity.UserRole(u.role),
				Status:            "active",
			}

			if err := db.WithContext(ctx).Omit("last_ip", "api_token").Create(&user).Error; err != nil {
				log.Printf("Failed to create user %s: %v", u.email, err)
			} else {
				fmt.Printf("✓ User created: %s (%s) → %s\n", u.fullname, u.email, u.role)
			}
		}
	}

	// === 6. Sync to Casbin ===
	fmt.Println("🔁 Syncing to Casbin...")

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("Failed to create Casbin GORM adapter: %v", err)
	}

	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	enforcer.ClearPolicy()

	// Load all roles and their permissions
	var roles []entity.Role
	db.WithContext(ctx).Find(&roles)

	for _, role := range roles {
		var perms []Permission
		if err := json.Unmarshal(role.Permissions, &perms); err != nil {
			log.Printf("Failed to unmarshal permissions for role %s: %v", role.Name, err)
			continue
		}

		// Add Casbin policies
		for _, perm := range perms {
			// Convert action to HTTP method
			method := actionToMethod(perm.Action)
			if method != "" {
				enforcer.AddPolicy(role.Name, perm.Path, method)
			}
		}
	}

	if err := enforcer.SavePolicy(); err != nil {
		log.Printf("Failed to save Casbin policies: %v", err)
	} else {
		fmt.Println("✅ Casbin policies synced!")
	}

	// === 7. Seed Mikrotik Devices ===
	fmt.Println("✓ Seeding Mikrotik devices...")

	encryptedAePassword, err := service.NewEncryptionService().Encrypt("001", cfg.Crypto.EncryptionKey)
	if err != nil {
		log.Printf("Error encrypting password: %v", err)
	}

	encryptedPassword, err := service.NewEncryptionService().Encrypt("r00t", cfg.Crypto.EncryptionKey)
	if err != nil {
		log.Printf("Error encrypting password: %v", err)
	}

	mikrotiks := []entity.Mikrotik{
		{
			Name:                 "Aeng Panas",
			Host:                 "103.139.193.128",
			Port:                 1012,
			ApiUsername:          "fandi1",
			ApiEncryptedPassword: encryptedAePassword,
			Keepalive:            false,
			Timeout:              0,
			Location:             "",
			Description:          "",
			IsActive:             true,
		},
		{
			Name:                 "Test",
			Host:                 "103.139.193.128",
			Port:                 1251,
			ApiUsername:          "admin",
			ApiEncryptedPassword: encryptedPassword,
			Keepalive:            false,
			Timeout:              0,
			Location:             "",
			Description:          "",
			IsActive:             true,
		},
	}

	for i := range mikrotiks {
		result := db.WithContext(ctx).
			Where("host = ? AND port = ?", mikrotiks[i].Host, mikrotiks[i].Port).
			FirstOrCreate(&mikrotiks[i])

		if result.Error != nil {
			log.Printf("Warning: Failed to create Mikrotik %s: %v", mikrotiks[i].Name, result.Error)
		} else {
			fmt.Printf("✓ Mikrotik created: %s (%s:%d)\n", mikrotiks[i].Name, mikrotiks[i].Host, mikrotiks[i].Port)
		}
	}

	fmt.Println("\n=== Seeding completed! ===")
	fmt.Println("\nDefault users:")
	fmt.Println("super@mikrobill.com     → super123 (superadmin)")
	fmt.Println("admin@mikrobill.com     → admin123 (admin)")
	fmt.Println("tech@mikrobill.com      → tech123 (technician)")
	fmt.Println("sales@mikrobill.com     → sales123 (sales)")
	fmt.Println("cs@mikrobill.com        → cs123 (cs)")
	fmt.Println("finance@mikrobill.com   → finance123 (finance)")
	fmt.Println("viewer@mikrobill.com    → viewer123 (viewer)")
}

// actionToMethod converts action name to HTTP method
func actionToMethod(action string) string {
	switch action {
	case "read":
		return "GET"
	case "create":
		return "POST"
	case "update":
		return "PUT"
	case "delete":
		return "DELETE"
	case "test", "verify", "export", "sync", "manage":
		return "POST"
	default:
		return ""
	}
}