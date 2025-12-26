package firewall

import "mikrobill/internal/infrastructure/mikrotik"

// Service adalah service untuk operasi MikroTik Firewall
type Service struct {
	client *mikrotik.Client
}

// NewService membuat instance baru dari Firewall Service
func NewService(client *mikrotik.Client) *Service {
	return &Service{
		client: client,
	}
}
