package inbound_port

import (
	"MikrOps/internal/model"
	"context"
)

type CustomerPort interface {
	CreateCustomer(a any) error
	GetCustomer(a any) error
	ListCustomers(a any) error
	UpdateCustomer(a any) error
	DeleteCustomer(a any) error

	// Prospect management
	PublicRegister(a any) error
	ListProspects(a any) error
	ApproveProspect(a any) error
	RejectProspect(a any) error
}

type CustomerDomain interface {
	CreateCustomer(ctx context.Context, input model.CreateCustomerRequest) (*model.Customer, error)
	GetCustomer(ctx context.Context, id string) (*model.Customer, error)
	ListCustomers(ctx context.Context) ([]model.Customer, error)
	UpdateCustomer(ctx context.Context, id string, input model.CreateCustomerRequest) (*model.Customer, error)
	// DeleteCustomer deletes a customer by ID
	DeleteCustomer(ctx context.Context, id string) error

	// HandlePPPoEUp handles on-up callback from MikroTik
	HandlePPPoEUp(ctx context.Context, input model.PPPoEEventInput) error

	// HandlePPPoEDown handles on-down callback from MikroTik
	HandlePPPoEDown(ctx context.Context, input model.PPPoEEventInput) error

	// RegisterProspect creates a prospect without MikroTik provisioning (public endpoint)
	RegisterProspect(ctx context.Context, slug string, input model.PublicRegistrationRequest) (*model.Customer, error)

	// ListProspects retrieves all prospects
	ListProspects(ctx context.Context) ([]model.Customer, error)

	// ApproveProspect approves a prospect and provisions to MikroTik
	ApproveProspect(ctx context.Context, input model.ApproveProspectRequest) (*model.Customer, error)

	// RejectProspect rejects a prospect (soft delete)
	RejectProspect(ctx context.Context, customerID string, reason *string) error
}
