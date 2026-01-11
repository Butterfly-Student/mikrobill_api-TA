package inbound_port

import "prabogo/internal/model"

type CustomerPort interface {
	CreateCustomer(a any) error
	GetCustomer(a any) error
	ListCustomers(a any) error
	UpdateCustomer(a any) error
	DeleteCustomer(a any) error
}

type CustomerDomain interface {
	CreateCustomer(ctx any, input model.CustomerInput) (*model.CustomerWithService, error)
	GetCustomer(ctx any, id string) (*model.CustomerWithService, error)
	ListCustomers(ctx any) ([]model.CustomerWithService, error)
	UpdateCustomer(ctx any, id string, input model.CustomerInput) (*model.CustomerWithService, error)
	DeleteCustomer(ctx any, id string) error
}
