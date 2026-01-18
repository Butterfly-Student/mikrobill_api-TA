package inbound_port

type HttpPort interface {
	Mikrotik() MikrotikHttpPort
	Testing() TestingHttpPort
	Middleware() MiddlewareHttpPort
	Ping() PingHttpPort
	Client() ClientHttpPort
	Auth() AuthHttpPort
	MikrotikPPPSecret() MikrotikPPPSecretPort
	MikrotikPPPProfile() MikrotikPPPProfilePort
	Monitor() MonitorPort
	Profile() ProfilePort
	Customer() CustomerPort
	Callback() CallbackHttpPort
	Tenant() TenantPort
	User() UserHttpPort
	DirectMonitor() DirectMonitorPort
}
