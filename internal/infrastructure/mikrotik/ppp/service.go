package ppp

import (
	"mikrobill/internal/infrastructure/mikrotik"
)

// Service adalah service untuk operasi MikroTik PPP
type Service struct {
	client *mikrotik.Client
}

// NewService membuat instance baru dari PPP Service
func NewService(client *mikrotik.Client) *Service {
	return &Service{
		client: client,
	}
}
