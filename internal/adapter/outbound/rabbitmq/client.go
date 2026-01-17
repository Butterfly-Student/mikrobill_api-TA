package rabbitmq_outbound_adapter

import (
	"context"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/rabbitmq"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientMessagePort {
	return &clientAdapter{}
}

func (adapter *clientAdapter) PublishUpsert(ctx context.Context, datas []model.ClientInput) error {
	err := rabbitmq.Publish(ctx, model.UpsertClientMessage, rabbitmq.KindFanOut, "", datas)
	if err != nil {
		return err
	}

	return nil
}

