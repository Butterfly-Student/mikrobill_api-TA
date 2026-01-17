package redis_outbound_adapter

import (
	outbound_port "MikrOps/internal/port/outbound"
)

type adapter struct {
}

func NewAdapter() outbound_port.CachePort {
	return &adapter{}
}

func (s *adapter) Client() outbound_port.ClientCachePort {
	return NewClientAdapter()
}

func (s *adapter) PubSub() outbound_port.RedisPubSubPort {
	return NewPubSubAdapter()
}

func (s *adapter) AuthCache() outbound_port.AuthCachePort {
	return NewAuthCacheAdapter()
}
