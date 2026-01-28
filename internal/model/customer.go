// file: models/customer.go
package model

import (
	"time"

	"gorm.io/gorm"
)

type CustomerStatus string
type ServiceType string
type ServiceStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
	CustomerStatusPending   CustomerStatus = "pending"
	CustomerStatusProspect  CustomerStatus = "prospect"

	ServiceTypePPPoE    ServiceType = "pppoe"
	ServiceTypeHotspot  ServiceType = "hotspot"
	ServiceTypeStaticIP ServiceType = "static_ip"

	ServiceStatusActive     ServiceStatus = "active"
	ServiceStatusSuspended  ServiceStatus = "suspended"
	ServiceStatusTerminated ServiceStatus = "terminated"
)

// Customer - Customer model
type Customer struct {
	ID         string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID   string `gorm:"type:uuid;not null;index" json:"tenant_id"`
	MikrotikID string `gorm:"type:uuid;not null;index" json:"mikrotik_id"`

	// Portal Login Credentials (for customer portal access)
	PortalEmail        *string `gorm:"type:varchar(255)" json:"portal_email,omitempty"`
	PortalPasswordHash *string `gorm:"type:text" json:"-"`

	// Service Credentials (for PPPoE/Hotspot on MikroTik)
	ServiceUsername          string  `gorm:"column:service_username;type:varchar(100);not null" json:"service_username"`
	ServicePasswordEncrypted *string `gorm:"type:text" json:"-"`
	ServicePasswordVisible   bool    `gorm:"default:false" json:"-"`

	// Basic info
	Name             string      `gorm:"type:varchar(255);not null" json:"name"`
	Phone            string      `gorm:"type:varchar(20);not null" json:"phone"`
	Email            *string     `gorm:"type:varchar(255)" json:"email,omitempty"`
	Address          *string     `gorm:"type:text" json:"address,omitempty"`
	MikrotikObjectID *string     `gorm:"type:varchar(50)" json:"mikrotik_object_id,omitempty"`
	ServiceType      ServiceType `gorm:"type:service_type;not null" json:"service_type"`

	// Network info
	AssignedIP *string    `gorm:"type:inet" json:"assigned_ip,omitempty"`
	MacAddress *string    `gorm:"type:macaddr" json:"mac_address,omitempty"`
	Interface  *string    `gorm:"type:varchar(50)" json:"interface,omitempty"`
	LastOnline *time.Time `json:"last_online,omitempty"`
	LastIP     *string    `gorm:"type:inet" json:"last_ip,omitempty"`

	// Status & Provisioning
	Status             CustomerStatus `gorm:"type:customer_status;default:'inactive'" json:"status"`
	ProvisioningStatus string         `gorm:"type:varchar(20);default:'pending'" json:"provisioning_status"`
	ProvisioningError  *string        `gorm:"type:text" json:"provisioning_error,omitempty"`
	ProvisionedAt      *time.Time     `json:"provisioned_at,omitempty"`

	// Billing
	AutoSuspension bool      `gorm:"default:true" json:"auto_suspension"`
	BillingDay     int       `gorm:"default:1" json:"billing_day"`
	JoinDate       time.Time `gorm:"type:date;default:CURRENT_DATE" json:"join_date"`
	CustomerNotes  *string   `gorm:"type:text" json:"customer_notes,omitempty"`

	// Audit
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant   Tenant            `gorm:"foreignKey:TenantID" json:"tenant"`
	Mikrotik Mikrotik          `gorm:"foreignKey:MikrotikID" json:"mikrotik"`
	Services []CustomerService `gorm:"foreignKey:CustomerID" json:"services,omitempty"`
}

func (Customer) TableName() string {
	return "customers"
}

// CustomerService - Service subscription
type CustomerService struct {
	ID         string        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID   string        `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CustomerID string        `gorm:"type:uuid;not null;index" json:"customer_id"`
	ProfileID  string        `gorm:"type:uuid;not null;index" json:"profile_id"`
	Price      float64       `gorm:"type:numeric(15,2);not null" json:"price"`
	TaxRate    float64       `gorm:"type:numeric(5,2);default:0.00" json:"tax_rate"`
	StartDate  time.Time     `gorm:"type:date;not null" json:"start_date"`
	EndDate    *time.Time    `gorm:"type:date" json:"end_date,omitempty"`
	Status     ServiceStatus `gorm:"type:service_status;default:'active'" json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`

	// Relations
	Tenant   Tenant          `gorm:"foreignKey:TenantID" json:"tenant"`
	Customer Customer        `gorm:"foreignKey:CustomerID" json:"customer"`
	Profile  MikrotikProfile `gorm:"foreignKey:ProfileID" json:"profile"`
}

func (CustomerService) TableName() string {
	return "customer_services"
}

// REQUEST/RESPONSE MODELS
type CreateCustomerRequest struct {
	Username       string      `json:"username" binding:"required"`
	Name           string      `json:"name" binding:"required"`
	Phone          string      `json:"phone" binding:"required"`
	Email          *string     `json:"email,omitempty"`
	Address        *string     `json:"address,omitempty"`
	Password       string      `json:"password" binding:"required,min=8"`
	ProfileID      string      `json:"profile_id" binding:"required,uuid4"`
	ServiceType    ServiceType `json:"service_type" binding:"required,oneof=pppoe hotspot static_ip"`
	BillingDay     *int        `json:"billing_day,omitempty"`
	AutoSuspension *bool       `json:"auto_suspension,omitempty"`
	StartDate      *time.Time  `json:"start_date,omitempty"`
	CustomerNotes  *string     `json:"customer_notes,omitempty"`
}

type CustomerResponse struct {
	ID               string         `json:"id"`
	TenantID         string         `json:"tenant_id"`
	MikrotikID       string         `json:"mikrotik_id"`
	Username         string         `json:"username"`
	Name             string         `json:"name"`
	Phone            string         `json:"phone"`
	Email            *string        `json:"email,omitempty"`
	Address          *string        `json:"address,omitempty"`
	MikrotikObjectID *string        `json:"mikrotik_object_id,omitempty"`
	ServiceType      ServiceType    `json:"service_type"`
	AssignedIP       *string        `json:"assigned_ip,omitempty"`
	MacAddress       *string        `json:"mac_address,omitempty"`
	Interface        *string        `json:"interface,omitempty"`
	LastOnline       *time.Time     `json:"last_online,omitempty"`
	LastIP           *string        `json:"last_ip,omitempty"`
	Status           CustomerStatus `json:"status"`
	AutoSuspension   bool           `json:"auto_suspension"`
	BillingDay       int            `json:"billing_day"`
	JoinDate         time.Time      `json:"join_date"`
	CustomerNotes    *string        `json:"customer_notes,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	// Relations
	Mikrotik       *MikrotikResponse         `json:"mikrotik,omitempty"`
	Services       []CustomerServiceResponse `json:"services,omitempty"`
	CurrentService *CustomerServiceResponse  `json:"current_service,omitempty"`
}

type CustomerServiceResponse struct {
	ID         string        `json:"id"`
	CustomerID string        `json:"customer_id"`
	ProfileID  string        `json:"profile_id"`
	Price      float64       `json:"price"`
	TaxRate    float64       `json:"tax_rate"`
	StartDate  time.Time     `json:"start_date"`
	EndDate    *time.Time    `json:"end_date,omitempty"`
	Status     ServiceStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	// Relation
	Profile ProfileResponse `json:"profile"`
}

type CustomerListResponse struct {
	Customers []CustomerResponse `json:"customers"`
	Total     int64              `json:"total"`
	Active    int64              `json:"active"`
	Inactive  int64              `json:"inactive"`
	Suspended int64              `json:"suspended"`
	Pending   int64              `json:"pending"`
}

func (c *Customer) ToResponse() *CustomerResponse {
	resp := &CustomerResponse{
		ID:               c.ID,
		TenantID:         c.TenantID,
		MikrotikID:       c.MikrotikID,
		Username:         c.ServiceUsername, // Use ServiceUsername instead of Username
		Name:             c.Name,
		Phone:            c.Phone,
		Email:            c.Email,
		Address:          c.Address,
		MikrotikObjectID: c.MikrotikObjectID,
		ServiceType:      c.ServiceType,
		AssignedIP:       c.AssignedIP,
		MacAddress:       c.MacAddress,
		Interface:        c.Interface,
		LastOnline:       c.LastOnline,
		LastIP:           c.LastIP,
		Status:           c.Status,
		AutoSuspension:   c.AutoSuspension,
		BillingDay:       c.BillingDay,
		JoinDate:         c.JoinDate,
		CustomerNotes:    c.CustomerNotes,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}

	return resp
}

func (cs *CustomerService) ToResponse() *CustomerServiceResponse {
	return &CustomerServiceResponse{
		ID:         cs.ID,
		CustomerID: cs.CustomerID,
		ProfileID:  cs.ProfileID,
		Price:      cs.Price,
		TaxRate:    cs.TaxRate,
		StartDate:  cs.StartDate,
		EndDate:    cs.EndDate,
		Status:     cs.Status,
		CreatedAt:  cs.CreatedAt,
		UpdatedAt:  cs.UpdatedAt,
	}
}

// ============================================================================
// PUBLIC REGISTRATION MODELS
// ============================================================================

// PublicRegistrationRequest - Request payload for public self-registration
type PublicRegistrationRequest struct {
	// Portal Login Credentials
	PortalEmail    string `json:"portal_email" binding:"required,email"`
	PortalPassword string `json:"portal_password" binding:"required,min=8"`

	// Customer Info
	Name    string  `json:"name" binding:"required"`
	Phone   string  `json:"phone" binding:"required"`
	Email   *string `json:"email" binding:"omitempty,email"`
	Address *string `json:"address"`

	// Service Configuration
	ProfileID   string      `json:"profile_id" binding:"required,uuid"`
	ServiceType ServiceType `json:"service_type" binding:"required,oneof=pppoe hotspot static_ip"`

	// Optional: For Hotspot only, customer can provide preferred service username
	// For PPPoE, this will be auto-generated
	PreferredServiceUsername *string `json:"preferred_service_username" binding:"omitempty,min=4,max=50"`
}

// ============================================================================
// CUSTOMER PORTAL MODELS
// ============================================================================

// CustomerLoginRequest - Customer portal login request
type CustomerLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CustomerLoginResponse - Customer portal login response
type CustomerLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// CustomerPortalProfileResponse - Customer profile response for portal
type CustomerPortalProfileResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Phone       string      `json:"phone"`
	Email       *string     `json:"email,omitempty"`
	Address     *string     `json:"address,omitempty"`
	ServiceType ServiceType `json:"service_type"`
	Status      string      `json:"status"`
	JoinDate    string      `json:"join_date"`

	// Service Credentials (visibility based on service_type)
	ServiceUsername string  `json:"service_username"`
	ServicePassword *string `json:"service_password,omitempty"` // Only shown if service_password_visible=TRUE (Hotspot)

	// Network Info
	AssignedIP *string    `json:"assigned_ip,omitempty"`
	LastOnline *time.Time `json:"last_online,omitempty"`

	// Service Info
	CurrentPlan *CustomerServiceResponse `json:"current_plan,omitempty"`
}

// CustomerSessionResetRequest - Request to reset customer PPP/Hotspot session
type CustomerSessionResetRequest struct {
	// No body needed, customer identified from JWT token
}

// CustomerSession - Customer portal session (stored in Redis)
type CustomerSession struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	TenantID     string    `json:"tenant_id"`
	TokenHash    string    `json:"token_hash"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IPAddress    *string   `json:"ip_address,omitempty"`
	UserAgent    *string   `json:"user_agent,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// ApproveProspectRequest - Request for approving prospect and provisioning to MikroTik
type ApproveProspectRequest struct {
	CustomerID     string     `json:"customer_id" binding:"required,uuid4"`
	BillingDay     *int       `json:"billing_day,omitempty" binding:"omitempty,min=1,max=31"`
	AutoSuspension *bool      `json:"auto_suspension,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	TechnicianNote *string    `json:"technician_note,omitempty"`
}
