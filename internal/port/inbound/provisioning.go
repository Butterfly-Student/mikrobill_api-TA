package inbound_port

// ProvisioningMessagePort defines the interface for handling provisioning messages from RabbitMQ
type ProvisioningMessagePort interface {
	// ProcessProvisioning handles a provisioning message
	// Returns true if message was processed successfully and should be ACKed
	ProcessProvisioning(msg any) bool

	// ProcessDisconnect handles a disconnect message
	// Returns true if message was processed successfully and should be ACKed
	ProcessDisconnect(msg any) bool
}
