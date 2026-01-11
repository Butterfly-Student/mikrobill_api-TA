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
	ID         uuid.UUID      `json:"id" db:"id"`
	MikrotikID uuid.UUID      `json:"mikrotik_id" db:"mikrotik_id"`
	Username   string         `json:"username" db:"username"`
	FullName   string         `json:"full_name" db:"full_name"`
	Phone      *string        `json:"phone,omitempty" db:"phone"`
	Email      *string        `json:"email,omitempty" db:"email"`
	Address    *string        `json:"address,omitempty" db:"address"`
	Status     CustomerStatus `json:"status" db:"status"`
	JoinDate   time.Time      `json:"join_date" db:"join_date"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" db:"updated_at"`
}

// CustomerService represents customer_services table
type CustomerService struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	CustomerID       uuid.UUID     `json:"customer_id" db:"customer_id"`
	ProfileID        uuid.UUID     `json:"profile_id" db:"profile_id"`
	MikrotikObjectID *string       `json:"mikrotik_object_id,omitempty" db:"mikrotik_object_id"`
	Price            float64       `json:"price" db:"price"`
	TaxRate          float64       `json:"tax_rate" db:"tax_rate"`
	StartDate        time.Time     `json:"start_date" db:"start_date"`
	EndDate          *time.Time    `json:"end_date,omitempty" db:"end_date"`
	Status           ServiceStatus `json:"status" db:"status"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
}

// CustomerWithService combines Customer with Service details
type CustomerWithService struct {
	Customer
	Service *CustomerService `json:"service,omitempty"`
}

// CustomerInput is the request DTO for creating a customer
type CustomerInput struct {
	Username  string     `json:"username" binding:"required"`
	FullName  string     `json:"full_name" binding:"required"`
	Phone     *string    `json:"phone"`
	Email     *string    `json:"email"`
	Address   *string    `json:"address"`
	Password  string     `json:"password" binding:"required"`
	ProfileID uuid.UUID  `json:"profile_id" binding:"required"`
	Price     float64    `json:"price" binding:"required"`
	TaxRate   *float64   `json:"tax_rate"`
	StartDate *time.Time `json:"start_date"`
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
}
