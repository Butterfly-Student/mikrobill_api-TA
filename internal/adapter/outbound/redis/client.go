package redis_outbound_adapter

import (
	"context"
	"encoding/json"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/redis"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientCachePort {
	return &clientAdapter{}
}

func (a *clientAdapter) SetClient(ctx context.Context, bearerKey string, client model.Client) error {
	bytes, err := json.Marshal(client)
	if err != nil {
		return err
	}
	return redis.Set(ctx, bearerKey, string(bytes))
}

func (a *clientAdapter) GetClient(ctx context.Context, bearerKey string) (model.Client, bool) {
	var client model.Client
	result, err := redis.Get(ctx, bearerKey)
	if err != nil {
		return model.Client{}, false
	}

	err = json.Unmarshal([]byte(result), &client)
	if err != nil {
		return model.Client{}, false
	}

	return client, true
}
