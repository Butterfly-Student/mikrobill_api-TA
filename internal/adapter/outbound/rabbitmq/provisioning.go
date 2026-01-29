package rabbitmq_outbound_adapter

import (
	"context"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/rabbitmq"
)

type provisioningAdapter struct{}

func NewProvisioningAdapter() outbound_port.ProvisioningMessagePort {
	return &provisioningAdapter{}
}

func (a *provisioningAdapter) PublishProvisioning(ctx context.Context, msg model.ProvisioningMessage) error {
	return rabbitmq.Publish(
		ctx,
		model.ProvisionCustomerMessage,
		rabbitmq.KindTopic,
		"customer.provision",
		msg,
	)
}

func (a *provisioningAdapter) PublishDisconnect(ctx context.Context, msg model.DisconnectMessage) error {
	return rabbitmq.Publish(
		ctx,
		model.DisconnectCustomerMessage,
		rabbitmq.KindTopic,
		"customer.disconnect",
		msg,
	)
}
