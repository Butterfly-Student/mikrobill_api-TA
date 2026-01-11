package inbound_port

type HttpPort interface {
	Mikrotik() MikrotikHttpPort
	Testing() TestingHttpPort
	Middleware() MiddlewareHttpPort
	Ping() PingHttpPort
	Client() ClientHttpPort
	Auth() AuthHttpPort
	PPP() PPPPort
	Monitor() MonitorPort
	Profile() ProfilePort
	Customer() CustomerPort
	Callback() CallbackHttpPort
}
