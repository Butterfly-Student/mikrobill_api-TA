package seeder

import (
	"fmt"
	"log"
	"time"

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
	Slug        *string `gorm:"uniqueIndex"`
	Subdomain   string  `gorm:"uniqueIndex"`
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

type MikrotikProfile struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID   string `gorm:"type:uuid"`
	MikrotikID string `gorm:"type:uuid"`
	Name       string
	Type       string `gorm:"type:service_type"`
	RateLimit  *string
	Price      float64 `gorm:"type:decimal(15,2)"`
	TaxRate    float64 `gorm:"type:decimal(5,2)"`
}

func (MikrotikProfile) TableName() string {
	return "mikrotik_profiles"
}

type Customer struct {
	ID                       string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID                 string `gorm:"type:uuid"`
	MikrotikID               string `gorm:"type:uuid"`
	Name                     string
	Phone                    string
	Email                    *string
	Address                  *string
	ServiceType              string `gorm:"type:service_type"`
	ServiceUsername          string
	ServicePasswordEncrypted *string
	ServicePasswordVisible   bool
	Status                   string  `gorm:"type:customer_status"`
	PortalEmail              *string `gorm:"uniqueIndex"`
	PortalPasswordHash       *string
}

type CustomerService struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenantID   string `gorm:"type:uuid"`
	CustomerID string `gorm:"type:uuid"`
	ProfileID  string `gorm:"type:uuid"`
	Price      float64
	TaxRate    float64
	StartDate  time.Time  `gorm:"type:date"`
	EndDate    *time.Time `gorm:"type:date"`
	Status     string     `gorm:"type:service_status"`
}

func (CustomerService) TableName() string {
	return "customer_services"
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

	if err := s.seedProfilesAndCustomers(); err != nil {
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
	slugTeknusantara := "teknusantara"
	slugMajubersama := "ispbandung" // Changed to ispbandung for testing
	slugDigiindo := "digiindo"

	tenants := []Tenant{
		{Name: "PT Teknologi Nusantara", Slug: &slugTeknusantara, Subdomain: "teknusantara", CompanyName: "PT Teknologi Nusantara", Status: "active", IsActive: true},
		{Name: "CV Maju Bersama", Slug: &slugMajubersama, Subdomain: "majubersama", CompanyName: "CV Maju Bersama Sejahtera", Status: "active", IsActive: true},
		{Name: "PT Digital Indonesia", Slug: &slugDigiindo, Subdomain: "digiindo", CompanyName: "PT Digital Indonesia Jaya", Status: "active", IsActive: true},
	}

	for _, tenant := range tenants {
		var existing Tenant
		if err := s.db.Where("subdomain = ?", tenant.Subdomain).First(&existing).Error; err != nil {
			if err := s.db.Create(&tenant).Error; err != nil {
				log.Printf("Error creating tenant %s: %v", tenant.Subdomain, err)
			} else {
				log.Printf("Created tenant: %s with slug: %s", tenant.Subdomain, *tenant.Slug)
			}
		} else {
			// Update existing tenant with slug if not set
			if existing.Slug == nil {
				existing.Slug = tenant.Slug
				s.db.Save(&existing)
				log.Printf("Updated tenant %s with slug: %s", tenant.Subdomain, *tenant.Slug)
			} else {
				log.Printf("Tenant %s already exists with slug: %s, skipping...", tenant.Subdomain, *existing.Slug)
			}
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
				{"admin@teknusantara.com", "admin@teknusantara.com", "Admin Teknisi Nusantara", "admin"},
			},
			mikrotiks: []struct{ name, host, location, description string }{
				{"MikroTik Teknisi", "192.168.88.1", "Kantor Pusat", "Router Utama"},
			},
		},
		{
			subdomain: "aengpanas",
			users: []struct{ username, email, fullname, role string }{
				{"aengpanas", "aengpanas@mikops.com", "Aeng Panas", "admin"},
				{"aengpanas_tech", "tech@aengpanas.com", "Teknisi Teknologi Nusantara", "technician"},
				{"aengpanas_viewer", "viewer@aengpanas.com", "Viewer Teknologi Nusantara", "viewer"},
			},
			mikrotiks: []struct{ name, host, location, description string }{
				{"MikroTik Utama", "103.139.193.128", "Jakarta Pusat", "Router utama"},
				{"MikroTik Cabang 1", "192.168.2.1", "Jakarta Selatan", "Router cabang selatan"},
				{"MikroTik Cabang 2", "192.168.3.1", "Tangerang", "Router cabang Tangerang"},
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

func (s *Seeder) seedProfilesAndCustomers() error {
	log.Println("Seeding billing profiles and sample customers...")

	// Define tenant profiles
	tenantProfiles := []struct {
		subdomain string
		profiles  []struct {
			name      string
			price     float64
			taxRate   float64
			type_     string
			rateLimit string
		}
		customers []struct {
			name        string
			phone       string
			email       string
			address     string
			portalEmail string
			portalPass  string
			profileIdx  int
			status      string
		}
	}{
		{
			subdomain: "teknusantara",
			profiles: []struct {
				name      string
				price     float64
				taxRate   float64
				type_     string
				rateLimit string
			}{
				{"10Mbps PPPoE", 150000, 11.00, "pppoe", "10M/10M"},
				{"20Mbps PPPoE", 250000, 11.00, "pppoe", "20M/20M"},
				{"5Mbps Hotspot", 100000, 11.00, "hotspot", "5M/5M"},
				{"Corporate Static IP", 500000, 11.00, "static_ip", ""},
			},
			customers: []struct {
				name        string
				phone       string
				email       string
				address     string
				portalEmail string
				portalPass  string
				profileIdx  int
				status      string
			}{
				{"John Doe", "081234567890", "john.doe@example.com", "Jl. Merdeka No. 123, Jakarta", "john.portal@example.com", "password123", 0, "active"},
				{"Jane Smith", "082345678901", "jane.smith@example.com", "Jl. Sudirman No. 456, Jakarta", "jane.portal@example.com", "password123", 1, "active"},
				{"Bob Wilson", "083456789012", "bob.wilson@example.com", "Jl. Thamrin No. 789, Jakarta", "bob.portal@example.com", "password123", 0, "prospect"},
			},
		},
		{
			subdomain: "majubersama",
			profiles: []struct {
				name      string
				price     float64
				taxRate   float64
				type_     string
				rateLimit string
			}{
				{"Paket Home 10Mbps", 125000, 11.00, "pppoe", "10M/10M"},
				{"Paket Home 20Mbps", 225000, 11.00, "pppoe", "20M/20M"},
				{"Paket Hotspot 5Mbps", 85000, 11.00, "hotspot", "5M/5M"},
			},
			customers: []struct {
				name        string
				phone       string
				email       string
				address     string
				portalEmail string
				portalPass  string
				profileIdx  int
				status      string
			}{
				{"Ahmad Fauzi", "081111222333", "ahmad.fauzi@example.com", "Jl. Ahmad Yani No. 100, Bandung", "ahmad.portal@example.com", "password123", 0, "active"},
				{"Siti Nurhaliza", "082222333444", "siti.nur@example.com", "Jl. Braga No. 50, Bandung", "siti.portal@example.com", "password123", 1, "prospect"},
			},
		},
		{
			subdomain: "digiindo",
			profiles: []struct {
				name      string
				price     float64
				taxRate   float64
				type_     string
				rateLimit string
			}{
				{"Business 50Mbps", 750000, 11.00, "pppoe", "50M/50M"},
				{"Enterprise 100Mbps", 1500000, 11.00, "pppoe", "100M/100M"},
				{"Dedicated Static IP", 2000000, 11.00, "static_ip", ""},
			},
			customers: []struct {
				name        string
				phone       string
				email       string
				address     string
				portalEmail string
				portalPass  string
				profileIdx  int
				status      string
			}{
				{"PT Sukses Jaya", "081999888777", "info@suksesjaya.com", "Jl. HR Muhammad No. 500, Surabaya", "suksesjaya.portal@example.com", "password123", 0, "active"},
			},
		},
	}

	hashedPassword := "$2a$10$G3KwVzSrJO1SuWNSYYu5KuISTXyeK5MIlCa5Z4c4OdU3esDN/i1bK" // password123

	for _, td := range tenantProfiles {
		// Get tenant
		var tenant Tenant
		if err := s.db.Where("subdomain = ?", td.subdomain).First(&tenant).Error; err != nil {
			log.Printf("Tenant %s not found, skipping profiles/customers...", td.subdomain)
			continue
		}

		// Get first active mikrotik for this tenant
		var mikrotik Mikrotik
		if err := s.db.Where("tenant_id = ? AND is_active = true", tenant.ID).First(&mikrotik).Error; err != nil {
			log.Printf("No active mikrotik found for tenant %s, skipping profiles...", td.subdomain)
			continue
		}

		log.Printf("Creating profiles for tenant: %s", td.subdomain)

		// Create profiles
		var createdProfiles []MikrotikProfile
		for _, profData := range td.profiles {
			var rateLimit *string
			if profData.rateLimit != "" {
				rateLimit = &profData.rateLimit
			}

			profile := MikrotikProfile{
				TenantID:   tenant.ID,
				MikrotikID: mikrotik.ID,
				Name:       profData.name,
				Type:       profData.type_,
				RateLimit:  rateLimit,
				Price:      profData.price,
				TaxRate:    profData.taxRate,
			}

			var existing MikrotikProfile
			if err := s.db.Where("tenant_id = ? AND name = ?", tenant.ID, profile.Name).First(&existing).Error; err != nil {
				if err := s.db.Create(&profile).Error; err != nil {
					log.Printf("Error creating profile %s: %v", profile.Name, err)
					continue
				}
				log.Printf("Created profile: %s (%.0f)", profile.Name, profile.Price)
				createdProfiles = append(createdProfiles, profile)
			} else {
				log.Printf("Profile %s already exists", profData.name)
				createdProfiles = append(createdProfiles, existing)
			}
		}

		if len(createdProfiles) == 0 {
			log.Printf("No profiles created for tenant %s, skipping customers...", td.subdomain)
			continue
		}

		log.Printf("Creating sample customers for tenant: %s", td.subdomain)

		// TODO: Customer seeding disabled temporarily due to complex schema
		// Customers can be created via public registration API for testing
		log.Printf("Customer seeding skipped - use public registration API instead")

		// Create customers
		for _, custData := range td.customers {
			if custData.profileIdx >= len(createdProfiles) {
				log.Printf("Invalid profile index %d for customer %s", custData.profileIdx, custData.name)
				continue
			}

			profile := createdProfiles[custData.profileIdx]

			// Generate service credentials for all customers
			// ServiceUsername is required by the schema (not nullable)
			username := fmt.Sprintf("%s%03d", td.subdomain[:4], custData.profileIdx+1)
			password := "password123"

			customer := Customer{
				TenantID:                 tenant.ID,
				MikrotikID:               mikrotik.ID,
				Name:                     custData.name,
				Phone:                    custData.phone,
				Email:                    &custData.email,
				Address:                  &custData.address,
				ServiceType:              profile.Type,
				ServiceUsername:          username,
				ServicePasswordEncrypted: &password,
				ServicePasswordVisible:   profile.Type == "hotspot",
				Status:                   custData.status,
				PortalEmail:              &custData.portalEmail,
				PortalPasswordHash:       &hashedPassword,
			}

			var existing Customer
			var createdCustomer Customer
			if err := s.db.Where("portal_email = ?", customer.PortalEmail).First(&existing).Error; err != nil {
				if err := s.db.Create(&customer).Error; err != nil {
					log.Printf("Error creating customer %s: %v", customer.Name, err)
					continue
				}
				log.Printf("Created customer: %s [%s] - Profile: %s", customer.Name, customer.Status, profile.Name)
				createdCustomer = customer
			} else {
				log.Printf("Customer %s already exists", custData.name)
				createdCustomer = existing
			}

			// Create CustomerService to link customer with their billing profile
			var existingService CustomerService
			if err := s.db.Where("customer_id = ? AND profile_id = ?", createdCustomer.ID, profile.ID).First(&existingService).Error; err != nil {
				// Determine service status based on customer status
				serviceStatus := "active"
				if custData.status != "active" {
					serviceStatus = "suspended"
				}

				customerService := CustomerService{
					TenantID:   tenant.ID,
					CustomerID: createdCustomer.ID,
					ProfileID:  profile.ID,
					Price:      profile.Price,
					TaxRate:    profile.TaxRate,
					StartDate:  time.Now(),
					Status:     serviceStatus,
				}

				if err := s.db.Create(&customerService).Error; err != nil {
					log.Printf("Error creating customer service for %s: %v", customer.Name, err)
				} else {
					log.Printf("Created customer service: %s -> %s (Rp%.0f)", customer.Name, profile.Name, profile.Price)
				}
			} else {
				log.Printf("Customer service already exists for %s", custData.name)
			}
		}
	}

	log.Println("Finished seeding profiles and customers.")
	return nil
}
