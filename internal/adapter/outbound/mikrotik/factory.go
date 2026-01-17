package mikrotik_outbound_adapter

import (
	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"time"
)

type mikrotikClientFactory struct{}

func NewMikrotikClientFactory() outbound_port.MikrotikClientFactory {
	return &mikrotikClientFactory{}
}

func (f *mikrotikClientFactory) NewClient(mikrotik *model.Mikrotik) (outbound_port.MikrotikClientPort, error) {
	password := ""
	if mikrotik.APIEncryptedPassword != nil {
		password = *mikrotik.APIEncryptedPassword
	}

	cfg := Config{
		Host:     mikrotik.Host,
		Port:     mikrotik.Port,
		Username: mikrotik.APIUsername,
		Password: password,
		Timeout:  5 * time.Second,       // Default timeout
		UseTLS:   mikrotik.Port == 8729, // Auto-detect TLS based on port
	}
	return NewClient(cfg)
}

