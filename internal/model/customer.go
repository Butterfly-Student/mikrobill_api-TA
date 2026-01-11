package model

import (
	"time"

	"github.com/google/uuid"
)

// CustomerStatus represents customer status enum
type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
	CustomerStatusPending   CustomerStatus = "pending"
)

// ServiceType represents service type enum
type ServiceType string

const (
	ServiceTypePPPoE    ServiceType = "pppoe"
	ServiceTypeHotspot  ServiceType = "hotspot"
	ServiceTypeStaticIP ServiceType = "static_ip"
)

// ServiceStatus represents service status enum
type ServiceStatus string

const (
	ServiceStatusActive    ServiceStatus = "active"
	ServiceStatusInactive  ServiceStatus = "inactive"
	ServiceStatusSuspended ServiceStatus = "suspended"
	ServiceStatusExpired   ServiceStatus = "expired"
)

// Customer represents customers table
type Customer struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	MikrotikID       uuid.UUID      `json:"mikrotik_id" db:"mikrotik_id"`
	Username         string         `json:"username" db:"username"`
	Name             string         `json:"name" db:"name"`
	Phone            string         `json:"phone" db:"phone"`
	Email            *string        `json:"email,omitempty" db:"email"`
	Address          *string        `json:"address,omitempty" db:"address"`
	MikrotikObjectID string         `json:"mikrotik_object_id" db:"mikrotik_object_id"`
	ServiceType      ServiceType    `json:"service_type" db:"service_type"`
	AssignedIP       *string        `json:"assigned_ip,omitempty" db:"assigned_ip"`
	MacAddress       *string        `json:"mac_address,omitempty" db:"mac_address"`
	Interface        *string        `json:"interface,omitempty" db:"interface"`
	LastOnline       *time.Time     `json:"last_online,omitempty" db:"last_online"`
	LastIP           *string        `json:"last_ip,omitempty" db:"last_ip"`
	Status           CustomerStatus `json:"status" db:"status"`
	AutoSuspension   bool           `json:"auto_suspension" db:"auto_suspension"`
	BillingDay       int            `json:"billing_day" db:"billing_day"`
	JoinDate         time.Time      `json:"join_date" db:"join_date"`
	CustomerNotes    *string        `json:"customer_notes,omitempty" db:"customer_notes"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

// CustomerService represents customer_services table
// Keeping this as is for now, but strictly relying on Customer for core Mikrotik sync
type CustomerService struct {
	ID         uuid.UUID     `json:"id" db:"id"`
	CustomerID uuid.UUID     `json:"customer_id" db:"customer_id"`
	ProfileID  uuid.UUID     `json:"profile_id" db:"profile_id"`
	Price      float64       `json:"price" db:"price"`
	TaxRate    float64       `json:"tax_rate" db:"tax_rate"`
	StartDate  time.Time     `json:"start_date" db:"start_date"`
	EndDate    *time.Time    `json:"end_date,omitempty" db:"end_date"`
	Status     ServiceStatus `json:"status" db:"status"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}

// CustomerWithService combines Customer with Service details
type CustomerWithService struct {
	Customer
	Service *CustomerService `json:"service,omitempty"`
}

// CustomerInput is the request DTO for creating a customer
type CustomerInput struct {
	Username       string      `json:"username" binding:"required"`
	Name           string      `json:"name" binding:"required"`
	Phone          string      `json:"phone" binding:"required"`
	Email          *string     `json:"email"`
	Address        *string     `json:"address"`
	Password       string      `json:"password" binding:"required"`
	ProfileID      uuid.UUID   `json:"profile_id" binding:"required"`
	ServiceType    ServiceType `json:"service_type" binding:"required"`
	Price          float64     `json:"price" binding:"required"`
	TaxRate        *float64    `json:"tax_rate"`
	BillingDay     int         `json:"billing_day"`
	AutoSuspension bool        `json:"auto_suspension"`
	StartDate      *time.Time  `json:"start_date"`
	CustomerNotes  *string     `json:"customer_notes"`
}

// PrepareCustomerInput sets default values for optional fields
func PrepareCustomerInput(input *CustomerInput) {
	if input.TaxRate == nil {
		taxRate := 11.00
		input.TaxRate = &taxRate
	}
	if input.StartDate == nil {
		now := time.Now()
		input.StartDate = &now
	}
	if input.BillingDay == 0 {
		input.BillingDay = 15
	}
}

// PPPoEUpInput represents the payload for on-up callback
type PPPoEUpInput struct {
	User       string `json:"user" binding:"required"`
	IPAddress  string `json:"ip"`
	Interface  string `json:"interface"`
	MacAddress string `json:"mac_address"`
}

// PPPoEDownInput represents the payload for on-down callback
type PPPoEDownInput struct {
	User string `json:"user" binding:"required"`
}
