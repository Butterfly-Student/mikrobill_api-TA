package redis_outbound_adapter

import (
	"context"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/redis"
)

type pubSubAdapter struct{}

func NewPubSubAdapter() outbound_port.RedisPubSubPort {
	return &pubSubAdapter{}
}

func (a *pubSubAdapter) Publish(channel string, message string) error {
	return redis.Publish(context.Background(), channel, message)
}

