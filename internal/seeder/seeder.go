package seeder

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

// Internal models for seeding to ensure isolation
type Tenant struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string
	Subdomain   string `gorm:"uniqueIndex"`
	CompanyName string
	Status      string
	IsActive    bool
}

type GlobalSetting struct {
	ID           string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SettingKey   string `gorm:"uniqueIndex"`
	SettingValue string
	SettingType  string
	Category     string
	Description  string
	IsPublic     bool
}

type Role struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	DisplayName string
	Description string
	IsSystem    bool
	TenantID    *string `gorm:"type:uuid"`
	Permissions string  `gorm:"type:jsonb"`
	IsActive    bool
}

type User struct {
	ID                string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID          *string `gorm:"type:uuid"`
	Username          string  `gorm:"uniqueIndex"`
	Email             string  `gorm:"uniqueIndex"`
	EncryptedPassword string
	Fullname          string
	UserRole          string `gorm:"type:user_role"`
	Status            string `gorm:"type:user_status"`
	IsSuperadmin      bool
}

type TenantUser struct {
	TenantID  string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;primaryKey"`
	RoleID    string `gorm:"type:uuid"`
	IsPrimary bool
	IsActive  bool
}

type Mikrotik struct {
	ID                   string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID             string `gorm:"type:uuid"`
	Name                 string
	Host                 string `gorm:"type:inet"`
	Port                 int
	APIUsername          string
	APIEncryptedPassword string
	Location             string
	Description          string
	IsActive             bool
	Status               string `gorm:"type:mikrotik_status"`
}

// TableName overrides the default table name
func (Mikrotik) TableName() string {
	return "mikrotik"
}

func (s *Seeder) Seed() error {
	log.Println("Starting data seeding with GORM...")

	if err := s.fixEnum(); err != nil {
		log.Printf("Warning: failed to fix enum: %v", err)
	}

	// Seed dalam urutan yang benar
	if err := s.seedTenants(); err != nil {
		return err
	}

	if err := s.seedSettings(); err != nil {
		return err
	}

	if err := s.seedRoles(); err != nil {
		return err
	}

	if err := s.seedSuperAdmin(); err != nil {
		return err
	}

	if err := s.seedTenantsWithUsersAndMikrotiks(); err != nil {
		return err
	}

	log.Println("Seeding completed successfully.")
	return nil
}

func (s *Seeder) fixEnum() error {
	var hasNewEnum bool
	err := s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_enum e 
			JOIN pg_type t ON e.enumtypid = t.oid 
			WHERE t.typname = 'user_role' 
			AND e.enumlabel = 'superadmin'
		)`).Scan(&hasNewEnum).Error
	if err != nil {
		return err
	}

	if hasNewEnum {
		log.Println("Enum user_role already has new values, skipping fix.")
		return nil
	}

	log.Println("Fixing user_role enum...")
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := []string{
		`ALTER TYPE user_role RENAME TO user_role_old`,
		`CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'technician', 'viewer')`,
		`ALTER TABLE users ALTER COLUMN user_role DROP DEFAULT`,
		`ALTER TABLE users ALTER COLUMN user_role TYPE user_role USING (
			CASE user_role::text
				WHEN 'SUPER_ADMIN' THEN 'superadmin'::user_role
				WHEN 'TENANT_OWNER' THEN 'admin'::user_role
				WHEN 'TENANT_ADMIN' THEN 'admin'::user_role
				WHEN 'TENANT_TECHNICIAN' THEN 'technician'::user_role
				WHEN 'TENANT_VIEWER' THEN 'viewer'::user_role
				ELSE 'viewer'::user_role
			END
		)`,
		`ALTER TABLE users ALTER COLUMN user_role SET DEFAULT 'viewer'`,
		`DROP TYPE user_role_old`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Enum user_role fixed successfully.")
	return nil
}

func (s *Seeder) seedTenants() error {
	tenants := []Tenant{
		{Name: "PT Teknologi Nusantara", Subdomain: "teknusantara", CompanyName: "PT Teknologi Nusantara", Status: "active", IsActive: true},
		{Name: "CV Maju Bersama", Subdomain: "majubersama", CompanyName: "CV Maju Bersama Sejahtera", Status: "active", IsActive: true},
		{Name: "PT Digital Indonesia", Subdomain: "digiindo", CompanyName: "PT Digital Indonesia Jaya", Status: "active", IsActive: true},
	}

	for _, tenant := range tenants {
		var existing Tenant
		if err := s.db.Where("subdomain = ?", tenant.Subdomain).First(&existing).Error; err != nil {
			if err := s.db.Create(&tenant).Error; err != nil {
				log.Printf("Error creating tenant %s: %v", tenant.Subdomain, err)
			} else {
				log.Printf("Created tenant: %s", tenant.Subdomain)
			}
		} else {
			log.Printf("Tenant %s already exists, skipping...", tenant.Subdomain)
		}
	}
	return nil
}

func (s *Seeder) seedSettings() error {
	if !s.db.Migrator().HasTable("global_settings") {
		log.Println("Table global_settings doesn't exist, skipping...")
		return nil
	}

	settings := []GlobalSetting{
		{SettingKey: "default_max_mikrotiks", SettingValue: "3", SettingType: "number", Category: "limits", Description: "Default maksimal mikrotik per tenant", IsPublic: false},
		{SettingKey: "default_max_network_users", SettingValue: "50", SettingType: "number", Category: "limits", Description: "Default maksimal user jaringan per tenant", IsPublic: false},
		{SettingKey: "default_max_staff_users", SettingValue: "5", SettingType: "number", Category: "limits", Description: "Default maksimal staff per tenant", IsPublic: false},
		{SettingKey: "system_name", SettingValue: "MikroBill System", SettingType: "string", Category: "system", Description: "Nama sistem", IsPublic: true},
		{SettingKey: "system_version", SettingValue: "1.0.0", SettingType: "string", Category: "system", Description: "Versi sistem saat ini", IsPublic: true},
	}

	for _, setting := range settings {
		var existing GlobalSetting
		if err := s.db.Where("setting_key = ?", setting.SettingKey).First(&existing).Error; err != nil {
			if err := s.db.Create(&setting).Error; err != nil {
				log.Printf("Error creating setting %s: %v", setting.SettingKey, err)
			} else {
				log.Printf("Created setting: %s", setting.SettingKey)
			}
		}
	}
	return nil
}

func (s *Seeder) seedRoles() error {
	if !s.db.Migrator().HasTable("roles") {
		log.Println("Table roles doesn't exist, skipping...")
		return nil
	}

	s.db.Exec("DELETE FROM roles WHERE name IN ('sales', 'cs', 'finance') AND tenant_id IS NULL")

	roles := []Role{
		{Name: "superadmin", DisplayName: "Super Administrator", Description: "Full system access across all tenants", IsSystem: true, IsActive: true, Permissions: `["*"]`},
		{Name: "admin", DisplayName: "Administrator", Description: "Administrative access within tenant", IsSystem: true, IsActive: true, Permissions: `["customer:*", "profile:*", "mikrotik:*", "monitor:*", "user:*", "billing:*"]`},
		{Name: "technician", DisplayName: "Technician", Description: "Technical operations within tenant", IsSystem: true, IsActive: true, Permissions: `["mikrotik:*", "monitor:*", "customer:read", "customer:list", "profile:read", "profile:list"]`},
		{Name: "viewer", DisplayName: "Viewer", Description: "Read-only access within tenant", IsSystem: true, IsActive: true, Permissions: `["*:read", "*:list"]`},
	}

	for _, role := range roles {
		role.TenantID = nil
		var existing Role
		if err := s.db.Where("name = ? AND tenant_id IS NULL", role.Name).First(&existing).Error; err != nil {
			if err := s.db.Create(&role).Error; err != nil {
				log.Printf("Error creating role %s: %v", role.Name, err)
			} else {
				log.Printf("Created role: %s", role.Name)
			}
		} else {
			existing.DisplayName = role.DisplayName
			existing.Description = role.Description
			existing.Permissions = role.Permissions
			existing.IsActive = role.IsActive
			s.db.Save(&existing)
		}
	}
	return nil
}

func (s *Seeder) seedSuperAdmin() error {
	if !s.db.Migrator().HasTable("users") {
		log.Println("Table users doesn't exist, skipping...")
		return nil
	}

	hashedPassword := "$2a$10$G3KwVzSrJO1SuWNSYYu5KuISTXyeK5MIlCa5Z4c4OdU3esDN/i1bK" // password123

	superadmin := User{
		Username:          "superadmin",
		Email:             "superadmin@mikrobill.com",
		Fullname:          "Super Administrator",
		UserRole:          "superadmin",
		Status:            "active",
		IsSuperadmin:      true,
		EncryptedPassword: hashedPassword,
	}

	var existing User
	if err := s.db.Where("username = ?", superadmin.Username).First(&existing).Error; err != nil {
		if err := s.db.Create(&superadmin).Error; err != nil {
			log.Printf("Error creating superadmin: %v", err)
		} else {
			log.Printf("Created superadmin user")
		}
	} else {
		log.Printf("Superadmin already exists, skipping...")
	}
	return nil
}

func (s *Seeder) seedTenantsWithUsersAndMikrotiks() error {
	hashedPassword := "$2a$10$G3KwVzSrJO1SuWNSYYu5KuISTXyeK5MIlCa5Z4c4OdU3esDN/i1bK" // password123
	// hashedPassword := "$2a$10$X2C8yPiFjowK0fxZ.Co2IekyBzq1t8D7rfYqYhHvJbqVl8nPlm7DK" // This hash might be incorrect

	// Data untuk setiap tenant
	tenantData := []struct {
		subdomain string
		users     []struct{ username, email, fullname, role string }
		mikrotiks []struct{ name, host, location, description string }
	}{
		{
			subdomain: "teknusantara",
			users: []struct{ username, email, fullname, role string }{
				{"teknusantara_admin", "admin@teknusantara.com", "Admin Teknologi Nusantara", "admin"},
				{"teknusantara_tech", "tech@teknusantara.com", "Teknisi Teknologi Nusantara", "technician"},
				{"teknusantara_viewer", "viewer@teknusantara.com", "Viewer Teknologi Nusantara", "viewer"},
			},
			mikrotiks: []struct{ name, host, location, description string }{
				{"MikroTik Jakarta Pusat", "192.168.1.1", "Jakarta Pusat", "Router utama kantor pusat"},
				{"MikroTik Jakarta Selatan", "192.168.2.1", "Jakarta Selatan", "Router cabang selatan"},
				{"MikroTik Tangerang", "192.168.3.1", "Tangerang", "Router cabang Tangerang"},
			},
		},
		{
			subdomain: "majubersama",
			users: []struct{ username, email, fullname, role string }{
				{"majubersama_admin", "admin@majubersama.com", "Admin Maju Bersama", "admin"},
				{"majubersama_tech", "tech@majubersama.com", "Teknisi Maju Bersama", "technician"},
				{"majubersama_viewer", "viewer@majubersama.com", "Viewer Maju Bersama", "viewer"},
			},
			mikrotiks: []struct{ name, host, location, description string }{
				{"MikroTik Bandung", "10.10.1.1", "Bandung", "Router pusat Bandung"},
				{"MikroTik Cimahi", "10.10.2.1", "Cimahi", "Router area Cimahi"},
				{"MikroTik Sumedang", "10.10.3.1", "Sumedang", "Router area Sumedang"},
			},
		},
		{
			subdomain: "digiindo",
			users: []struct{ username, email, fullname, role string }{
				{"digiindo_admin", "admin@digiindo.com", "Admin Digital Indonesia", "admin"},
				{"digiindo_tech", "tech@digiindo.com", "Teknisi Digital Indonesia", "technician"},
				{"digiindo_viewer", "viewer@digiindo.com", "Viewer Digital Indonesia", "viewer"},
			},
			mikrotiks: []struct{ name, host, location, description string }{
				{"MikroTik Surabaya", "172.16.1.1", "Surabaya", "Router kantor Surabaya"},
				{"MikroTik Malang", "172.16.2.1", "Malang", "Router kantor Malang"},
				{"MikroTik Gresik", "172.16.3.1", "Gresik", "Router area Gresik"},
			},
		},
	}

	for _, td := range tenantData {
		// Ambil tenant
		var tenant Tenant
		if err := s.db.Where("subdomain = ?", td.subdomain).First(&tenant).Error; err != nil {
			log.Printf("Tenant %s not found, skipping...", td.subdomain)
			continue
		}

		log.Printf("Processing tenant: %s", td.subdomain)

		// Seed users untuk tenant ini
		for _, userData := range td.users {
			user := User{
				Username:          userData.username,
				Email:             userData.email,
				Fullname:          userData.fullname,
				UserRole:          userData.role,
				Status:            "active",
				IsSuperadmin:      false,
				EncryptedPassword: hashedPassword,
				TenantID:          &tenant.ID,
			}

			var existingUser User
			if err := s.db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
				if err := s.db.Create(&user).Error; err != nil {
					log.Printf("Error creating user %s: %v", user.Username, err)
					continue
				}
				log.Printf("Created user: %s", user.Username)
				existingUser = user
			} else {
				log.Printf("User %s already exists, updating password...", user.Username)
				// Update password and ensure tenant link
				existingUser.EncryptedPassword = hashedPassword
				existingUser.TenantID = &tenant.ID
				s.db.Save(&existingUser)
			}

			// Link user ke tenant
			if err := s.linkUserToTenant(tenant.ID, existingUser.ID, existingUser.UserRole); err != nil {
				log.Printf("Error linking user %s to tenant: %v", user.Username, err)
			}
		}

		// Seed mikrotiks untuk tenant ini
		for _, mtData := range td.mikrotiks {
			mikrotik := Mikrotik{
				TenantID:             tenant.ID,
				Name:                 mtData.name,
				Host:                 mtData.host,
				Port:                 8728,
				APIUsername:          "admin",
				APIEncryptedPassword: hashedPassword,
				Location:             mtData.location,
				Description:          mtData.description,
				IsActive:             true,
				Status:               "offline",
			}

			var existingMT Mikrotik
			if err := s.db.Where("tenant_id = ? AND host = ?", tenant.ID, mikrotik.Host).First(&existingMT).Error; err != nil {
				if err := s.db.Create(&mikrotik).Error; err != nil {
					log.Printf("Error creating mikrotik %s: %v", mikrotik.Name, err)
				} else {
					log.Printf("Created mikrotik: %s (%s)", mikrotik.Name, mikrotik.Host)
				}
			} else {
				log.Printf("Mikrotik %s already exists", mtData.name)
			}
		}
	}

	return nil
}

func (s *Seeder) linkUserToTenant(tenantID, userID, userRole string) error {
	if !s.db.Migrator().HasTable("tenant_users") {
		return nil
	}

	// Ambil role yang sesuai
	var role Role
	if err := s.db.Where("name = ? AND tenant_id IS NULL", userRole).First(&role).Error; err != nil {
		return fmt.Errorf("role not found for %s: %v", userRole, err)
	}

	// Cek apakah sudah ada link
	var tenantUser TenantUser
	if err := s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&tenantUser).Error; err != nil {
		// Create new link
		tenantUser = TenantUser{
			TenantID:  tenantID,
			UserID:    userID,
			RoleID:    role.ID,
			IsPrimary: true,
			IsActive:  true,
		}
		if err := s.db.Create(&tenantUser).Error; err != nil {
			return err
		}
		log.Printf("Linked user to tenant with role %s", userRole)
	}
	return nil
}
