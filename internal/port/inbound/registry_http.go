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
	MikrotikPPPActive() MikrotikPPPActivePort
	MikrotikPPPInactive() MikrotikPPPInactivePort
	MikrotikPool() MikrotikPoolPort
	MikrotikQueue() MikrotikQueuePort
	MikrotikLog() MikrotikLogPort
	Monitor() MonitorPort
	Profile() ProfilePort
	Customer() CustomerPort
	Callback() CallbackHttpPort
	Tenant() TenantPort
	User() UserHttpPort
	DirectMonitor() DirectMonitorPort
	PPPRealtime() PPPRealtimePort
}
