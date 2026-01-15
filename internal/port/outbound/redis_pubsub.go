package outbound_port

//go:generate mockgen -source=redis_pubsub.go -destination=./../../../tests/mocks/port/mock_redis_pubsub.go

type RedisPubSubPort interface {
	Publish(channel string, message string) error
}
