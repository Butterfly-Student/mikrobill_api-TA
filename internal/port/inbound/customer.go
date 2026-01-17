package inbound_port

import (
	"context"
	"MikrOps/internal/model"
)

type CustomerPort interface {
	CreateCustomer(a any) error
	GetCustomer(a any) error
	ListCustomers(a any) error
	UpdateCustomer(a any) error
	DeleteCustomer(a any) error
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
}

