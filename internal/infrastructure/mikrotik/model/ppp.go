package model

// ========== PPP SECRET ==========

// PPPSecretRequest adalah request untuk membuat/update PPP secret
type PPPSecretRequest struct {
	Name           string `json:"name" binding:"required"`
	Password       string `json:"password" binding:"required"`
	Service        string `json:"service,omitempty"`        // any, pppoe, pptp, l2tp, ovpn, sstp
	Profile        string `json:"profile,omitempty"`        // Profile name
	LocalAddress   string `json:"localAddress,omitempty"`   // IP untuk server
	RemoteAddress  string `json:"remoteAddress,omitempty"`  // IP untuk client
	CallerID       string `json:"callerID,omitempty"`       // MAC address
	Routes         string `json:"routes,omitempty"`         // Static routes
	Comment        string `json:"comment,omitempty"`
	Disabled       bool   `json:"disabled,omitempty"`
	LimitBytesIn   string `json:"limitBytesIn,omitempty"`   // Upload limit
	LimitBytesOut  string `json:"limitBytesOut,omitempty"`  // Download limit
}

// PPPSecretUpdateRequest untuk update PPP secret
type PPPSecretUpdateRequest struct {
	Password       string `json:"password,omitempty"`
	Service        string `json:"service,omitempty"`
	Profile        string `json:"profile,omitempty"`
	LocalAddress   string `json:"localAddress,omitempty"`
	RemoteAddress  string `json:"remoteAddress,omitempty"`
	CallerID       string `json:"callerID,omitempty"`
	Routes         string `json:"routes,omitempty"`
	Comment        string `json:"comment,omitempty"`
	Disabled       *bool  `json:"disabled,omitempty"`
	LimitBytesIn   string `json:"limitBytesIn,omitempty"`
	LimitBytesOut  string `json:"limitBytesOut,omitempty"`
}

// PPPSecretResponse adalah response dari operasi PPP secret
type PPPSecretResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PPPSecretData adalah data PPP secret dari MikroTik
type PPPSecretData struct {
	ID            string `json:".id"`
	Name          string `json:"name"`
	Password      string `json:"password,omitempty"`
	Service       string `json:"service"`
	Profile       string `json:"profile"`
	LocalAddress  string `json:"local-address"`
	RemoteAddress string `json:"remote-address"`
	CallerID      string `json:"caller-id,omitempty"`
	Routes        string `json:"routes,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Disabled      string `json:"disabled"`
	LastLoggedOut string `json:"last-logged-out,omitempty"`
}

// ========== PPP ACTIVE ==========

// PPPActiveData adalah data PPP active connection
type PPPActiveData struct {
	ID            string `json:".id"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	CallerID      string `json:"caller-id"`
	Address       string `json:"address"`
	Uptime        string `json:"uptime"`
	Encoding      string `json:"encoding,omitempty"`
	SessionID     string `json:"session-id,omitempty"`
	LimitBytesIn  string `json:"limit-bytes-in,omitempty"`
	LimitBytesOut string `json:"limit-bytes-out,omitempty"`
	Radius        string `json:"radius,omitempty"`
}

// PPPActiveResponse adalah response dari operasi PPP active
type PPPActiveResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ========== PPP PROFILE ==========

// PPPProfileRequest adalah request untuk membuat/update PPP profile
type PPPProfileRequest struct {
	Name                 string `json:"name" binding:"required"`
	LocalAddress         string `json:"localAddress,omitempty"`
	RemoteAddress        string `json:"remoteAddress,omitempty"`
	DNSServer            string `json:"dnsServer,omitempty"`
	WINSServer           string `json:"winsServer,omitempty"`
	RateLimit            string `json:"rateLimit,omitempty"`
	SessionTimeout       string `json:"sessionTimeout,omitempty"`
	IdleTimeout          string `json:"idleTimeout,omitempty"`
	OnlyOne              string `json:"onlyOne,omitempty"`           // yes, no, default
	ChangeTCP_MSS       string `json:"changeTcpMss,omitempty"`      // yes, no, default
	UseEncryption        string `json:"useEncryption,omitempty"`     // yes, no, default, required
	UseCompression       string `json:"useCompression,omitempty"`    // yes, no, default
	UseVJ_Compression    string `json:"useVjCompression,omitempty"`  // yes, no, default
	UseMPLS              string `json:"useMpls,omitempty"`           // yes, no, default, required
	UseIPv6              string `json:"useIpv6,omitempty"`           // yes, no, default, required
	AddressList          string `json:"addressList,omitempty"`
	IncomingFilter       string `json:"incomingFilter,omitempty"`
	OutgoingFilter       string `json:"outgoingFilter,omitempty"`
	Comment              string `json:"comment,omitempty"`
}

// PPPProfileUpdateRequest untuk update PPP profile
type PPPProfileUpdateRequest struct {
	LocalAddress         string `json:"localAddress,omitempty"`
	RemoteAddress        string `json:"remoteAddress,omitempty"`
	DNSServer            string `json:"dnsServer,omitempty"`
	WINSServer           string `json:"winsServer,omitempty"`
	RateLimit            string `json:"rateLimit,omitempty"`
	SessionTimeout       string `json:"sessionTimeout,omitempty"`
	IdleTimeout          string `json:"idleTimeout,omitempty"`
	OnlyOne              string `json:"onlyOne,omitempty"`
	ChangeTC_MSS         string `json:"changeTcpMss,omitempty"`
	UseEncryption        string `json:"useEncryption,omitempty"`
	UseCompression       string `json:"useCompression,omitempty"`
	UseVJ_Compression    string `json:"useVjCompression,omitempty"`
	UseMPLS              string `json:"useMpls,omitempty"`
	UseIPv6              string `json:"useIpv6,omitempty"`
	AddressList          string `json:"addressList,omitempty"`
	IncomingFilter       string `json:"incomingFilter,omitempty"`
	OutgoingFilter       string `json:"outgoingFilter,omitempty"`
	Comment              string `json:"comment,omitempty"`
}

// PPPProfileResponse adalah response dari operasi PPP profile
type PPPProfileResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PPPProfileData adalah data PPP profile dari MikroTik
type PPPProfileData struct {
	ID                string `json:".id"`
	Name              string `json:"name"`
	LocalAddress      string `json:"local-address"`
	RemoteAddress     string `json:"remote-address"`
	DNSServer         string `json:"dns-server,omitempty"`
	WINSServer        string `json:"wins-server,omitempty"`
	RateLimit         string `json:"rate-limit,omitempty"`
	SessionTimeout    string `json:"session-timeout,omitempty"`
	IdleTimeout       string `json:"idle-timeout,omitempty"`
	OnlyOne           string `json:"only-one"`
	ChangeTC_MSS      string `json:"change-tcp-mss"`
	UseEncryption     string `json:"use-encryption"`
	UseCompression    string `json:"use-compression"`
	UseVJ_Compression string `json:"use-vj-compression"`
	UseMPLS           string `json:"use-mpls"`
	UseIPv6           string `json:"use-ipv6"`
	AddressList       string `json:"address-list,omitempty"`
	Comment           string `json:"comment,omitempty"`
}