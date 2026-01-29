package outbound_port

import (
	"MikrOps/internal/model"
	"context"
)

//go:generate mockgen -source=provisioning.go -destination=./../../../tests/mocks/port/mock_provisioning.go
type ProvisioningMessagePort interface {
	// PublishProvisioning sends a message to provision a customer on MikroTik
	PublishProvisioning(ctx context.Context, msg model.ProvisioningMessage) error

	// PublishDisconnect sends a message to disconnect a customer from MikroTik
	PublishDisconnect(ctx context.Context, msg model.DisconnectMessage) error
}
