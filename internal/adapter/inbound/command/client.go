package command_inbound_adapter

import (
	"context"
	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
	"MikrOps/utils/log"
)

type clientAdapter struct {
	domain domain.Domain
}

func NewClientAdapter(
	domain domain.Domain,
) inbound_port.ClientCommandPort {
	return &clientAdapter{
		domain: domain,
	}
}

func (h *clientAdapter) PublishUpsert(name string) {
	ctx := activity.NewContext(context.Background(), "command_client_sync")
	ctx = context.WithValue(ctx, activity.Payload, name)
	payload := []model.ClientInput{{Name: name}}
	err := h.domain.Client().PublishUpsert(ctx, payload)
	if err != nil {
		log.WithContext(ctx).Errorf("client upsert error %s: %s", err.Error(), name)
	}
	log.WithContext(ctx).Info("client upsert success")
}

