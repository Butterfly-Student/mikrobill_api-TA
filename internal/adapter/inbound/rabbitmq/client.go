package rabbitmq_inbound_adapter

import (
	"context"
	"encoding/json"
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
) inbound_port.ClientMessagePort {
	return &clientAdapter{
		domain: domain,
	}
}

func (h *clientAdapter) Upsert(a any) bool {
	msg := a.([]byte)
	ctx := activity.NewContext(context.Background(), "rabbitmq_client_sync")
	var payload []model.ClientInput
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		log.WithContext(ctx).Errorf("client upsert error %s: %s", err.Error(), string(msg))
		return true
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	results, err := h.domain.Client().Upsert(ctx, payload)
	if err != nil {
		log.WithContext(ctx).Errorf("client upsert error %s: %s", err.Error(), string(msg))
	}
	ctx = context.WithValue(ctx, activity.Result, results)

	log.WithContext(ctx).Info("client upsert success")
	return true
}

