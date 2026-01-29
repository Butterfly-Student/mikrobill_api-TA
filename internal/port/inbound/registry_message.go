package inbound_port

type MessagePort interface {
	Client() ClientMessagePort
	Provisioning() ProvisioningMessagePort
}
