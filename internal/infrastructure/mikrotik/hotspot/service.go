package hotspot

import "mikrobill/internal/infrastructure/mikrotik"

// Service adalah service untuk operasi MikroTik Hotspot
type Service struct {
	client *mikrotik.Client
}

// NewService membuat instance baru dari Service
func NewService(client *mikrotik.Client) *Service {
	return &Service{
		client: client,
	}
}
