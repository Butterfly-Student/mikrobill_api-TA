package entity

type ProfileStaticIP struct {
	ProfileID            string   `gorm:"primaryKey;type:uuid;column:profile_id"`
	IPPool               string   `gorm:"column:ip_pool;type:varchar(50)"`
	Gateway              string   `gorm:"column:gateway;type:varchar(50);not null"`
	Netmask              string   `gorm:"column:netmask;type:varchar(50);not null;default:'255.255.255.0'"`
	AllowedMACAddresses  []string `gorm:"column:allowed_mac_addresses;type:text[]"`
	FirewallChain        string   `gorm:"column:firewall_chain;type:varchar(50)"`
	VLANID               int      `gorm:"column:vlan_id"`
	VLANPriority         int      `gorm:"column:vlan_priority"`
	RouteDistance        int      `gorm:"column:route_distance;default:1"`
	RoutingMark          string   `gorm:"column:routing_mark;type:varchar(50)"`

	// Relations
	Profile *MikrotikProfile `gorm:"foreignKey:ProfileID"`
}

func (ProfileStaticIP) TableName() string { return "mikrotik_profile_static_ip" }