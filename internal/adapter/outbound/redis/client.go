package redis_outbound_adapter

import (
	"context"
	"encoding/json"
	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/redis"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientCachePort {
	return &clientAdapter{}
}

func (adapter *clientAdapter) Set(ctx context.Context, data model.Client) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return redis.Set(ctx, data.BearerKey, string(bytes))
}

func (adapter *clientAdapter) Get(ctx context.Context, bearerKey string) (model.Client, error) {
	var client model.Client
	result, err := redis.Get(ctx, bearerKey)
	if err != nil {
		return model.Client{}, err
	}

	err = json.Unmarshal([]byte(result), &client)
	if err != nil {
		return model.Client{}, err
	}

	return client, nil
}

