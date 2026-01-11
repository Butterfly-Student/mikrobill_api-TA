package outbound_port

type RedisPubSubPort interface {
	Publish(channel string, message string) error
}
