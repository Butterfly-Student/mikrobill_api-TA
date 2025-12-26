package model

// ========== FIREWALL NAT ==========

// NATAction tipe untuk action NAT
type NATAction string

const (
	NATActionSrcNAT     NATAction = "src-nat"
	NATActionMasquerade NATAction = "masquerade"
	NATActionDstNAT     NATAction = "dst-nat"
	NATActionRedirect   NATAction = "redirect"
	NATActionNetmap     NATAction = "netmap"
	NATActionSame       NATAction = "same"
)

// NATChain tipe untuk chain
type NATChain string

const (
	NATChainSrcNAT NATChain = "srcnat"
	NATChainDstNAT NATChain = "dstnat"
)

// NATRequest adalah request untuk membuat/update NAT rule
type NATRequest struct {
	Chain              NATChain  `json:"chain" binding:"required"`
	Action             NATAction `json:"action" binding:"required"`
	SrcAddress         string    `json:"srcAddress,omitempty"`
	SrcAddressList     string    `json:"srcAddressList,omitempty"`
	DstAddress         string    `json:"dstAddress,omitempty"`
	DstAddressList     string    `json:"dstAddressList,omitempty"`
	Protocol           string    `json:"protocol,omitempty"` // tcp, udp, icmp, etc
	SrcPort            string    `json:"srcPort,omitempty"`
	DstPort            string    `json:"dstPort,omitempty"`
	InInterface        string    `json:"inInterface,omitempty"`
	InInterfaceList    string    `json:"inInterfaceList,omitempty"`
	OutInterface       string    `json:"outInterface,omitempty"`
	OutInterfaceList   string    `json:"outInterfaceList,omitempty"`
	ToAddresses        string    `json:"toAddresses,omitempty"`     // untuk src-nat/dst-nat
	ToPorts            string    `json:"toPorts,omitempty"`         // untuk dst-nat/redirect
	Comment            string    `json:"comment,omitempty"`
	Disabled           bool      `json:"disabled,omitempty"`
	Log                bool      `json:"log,omitempty"`
	LogPrefix          string    `json:"logPrefix,omitempty"`
}

// NATUpdateRequest untuk update NAT rule
type NATUpdateRequest struct {
	Action             NATAction `json:"action,omitempty"`
	SrcAddress         string    `json:"srcAddress,omitempty"`
	SrcAddressList     string    `json:"srcAddressList,omitempty"`
	DstAddress         string    `json:"dstAddress,omitempty"`
	DstAddressList     string    `json:"dstAddressList,omitempty"`
	Protocol           string    `json:"protocol,omitempty"`
	SrcPort            string    `json:"srcPort,omitempty"`
	DstPort            string    `json:"dstPort,omitempty"`
	InInterface        string    `json:"inInterface,omitempty"`
	InInterfaceList    string    `json:"inInterfaceList,omitempty"`
	OutInterface       string    `json:"outInterface,omitempty"`
	OutInterfaceList   string    `json:"outInterfaceList,omitempty"`
	ToAddresses        string    `json:"toAddresses,omitempty"`
	ToPorts            string    `json:"toPorts,omitempty"`
	Comment            string    `json:"comment,omitempty"`
	Disabled           *bool     `json:"disabled,omitempty"`
	Log                *bool     `json:"log,omitempty"`
	LogPrefix          string    `json:"logPrefix,omitempty"`
}

// NATResponse adalah response dari operasi NAT
type NATResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NATData adalah data NAT rule dari MikroTik
type NATData struct {
	ID               string `json:".id"`
	Chain            string `json:"chain"`
	Action           string `json:"action"`
	SrcAddress       string `json:"src-address,omitempty"`
	SrcAddressList   string `json:"src-address-list,omitempty"`
	DstAddress       string `json:"dst-address,omitempty"`
	DstAddressList   string `json:"dst-address-list,omitempty"`
	Protocol         string `json:"protocol,omitempty"`
	SrcPort          string `json:"src-port,omitempty"`
	DstPort          string `json:"dst-port,omitempty"`
	InInterface      string `json:"in-interface,omitempty"`
	InInterfaceList  string `json:"in-interface-list,omitempty"`
	OutInterface     string `json:"out-interface,omitempty"`
	OutInterfaceList string `json:"out-interface-list,omitempty"`
	ToAddresses      string `json:"to-addresses,omitempty"`
	ToPorts          string `json:"to-ports,omitempty"`
	Comment          string `json:"comment,omitempty"`
	Disabled         string `json:"disabled"`
	Invalid          string `json:"invalid,omitempty"`
	Dynamic          string `json:"dynamic,omitempty"`
	Bytes            string `json:"bytes,omitempty"`
	Packets          string `json:"packets,omitempty"`
	Log              string `json:"log,omitempty"`
	LogPrefix        string `json:"log-prefix,omitempty"`
}

// NATMasqueradeRequest untuk quick masquerade setup
type NATMasqueradeRequest struct {
	OutInterface     string `json:"outInterface" binding:"required"` // Interface WAN
	SrcAddress       string `json:"srcAddress,omitempty"`            // LAN network (e.g., 192.168.88.0/24)
	Comment          string `json:"comment,omitempty"`
}