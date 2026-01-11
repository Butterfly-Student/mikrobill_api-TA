package inbound_port

type MonitorPort interface {
	StreamTraffic(a any) error
}

type MonitorDomain interface {
	StreamTraffic(ctx any, interfaceName string) (<-chan map[string]interface{}, func(), error)
}
