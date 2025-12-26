package entity

import "time"


type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusSuspended CustomerStatus = "suspended"
	CustomerStatusInactive  CustomerStatus = "inactive"
)

// Customer represents the customers table
type Customer struct {
	ID                  string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID          string         `gorm:"column:mikrotik_id;type:uuid;not null"`
	Username            string         `gorm:"column:username;type:varchar(100);not null"`
	Name                string         `gorm:"column:name;type:varchar(255);not null"`
	Phone               string         `gorm:"column:phone;type:varchar(20);not null"`
	Email               string         `gorm:"column:email;type:varchar(255)"`
	Address             string         `gorm:"column:address;type:text"`
	Latitude            float64        `gorm:"column:latitude;type:decimal(10,8)"`
	Longitude           float64        `gorm:"column:longitude;type:decimal(11,8)"`
	ServiceType         ServiceType    `gorm:"column:service_type;type:service_type;not null"`
	PackageID           *string        `gorm:"column:package_id;type:uuid"`
	PPPoEUsername       string         `gorm:"column:pppoe_username;type:varchar(100)"`
	PPPoEPassword       string         `gorm:"column:pppoe_password;type:varchar(100)"`
	PPPoEProfileID      *string        `gorm:"column:pppoe_profile_id;type:uuid"`
	HotspotUsername     string         `gorm:"column:hotspot_username;type:varchar(100)"`
	HotspotPassword     string         `gorm:"column:hotspot_password;type:varchar(100)"`
	HotspotProfileID    *string        `gorm:"column:hotspot_profile_id;type:uuid"`
	HotspotMACAddress   string         `gorm:"column:hotspot_mac_address;type:macaddr"`
	HotspotIPAddress    string         `gorm:"column:hotspot_ip_address;type:inet"`
	StaticIP            string         `gorm:"column:static_ip;type:inet"`
	StaticIPNetmask     string         `gorm:"column:static_ip_netmask;type:varchar(20)"`
	StaticIPGateway     string         `gorm:"column:static_ip_gateway;type:inet"`
	StaticIPDNS1        string         `gorm:"column:static_ip_dns1;type:inet"`
	StaticIPDNS2        string         `gorm:"column:static_ip_dns2;type:inet"`
	AssignedIP          string         `gorm:"column:assigned_ip;type:inet"`
	MACAddress          string         `gorm:"column:mac_address;type:macaddr"`
	LastOnline          *time.Time     `gorm:"column:last_online;type:timestamptz"`
	LastIP              string         `gorm:"column:last_ip;type:inet"`
	ODPID               *string        `gorm:"column:odp_id;type:uuid"`
	Status              CustomerStatus `gorm:"column:status;type:customer_status;default:'active'"`
	AutoSuspension      bool           `gorm:"column:auto_suspension;default:true"`
	BillingDay          int            `gorm:"column:billing_day;default:15"`
	JoinDate            time.Time      `gorm:"column:join_date;type:timestamptz;default:CURRENT_TIMESTAMP"`
	CableType           string         `gorm:"column:cable_type;type:varchar(50)"`
	CableLength         int            `gorm:"column:cable_length"`
	PortNumber          int            `gorm:"column:port_number"`
	CableStatus         CableStatus    `gorm:"column:cable_status;type:cable_status;default:'connected'"`
	CableNotes          string         `gorm:"column:cable_notes;type:text"`
	FUPQuotaUsed        int            `gorm:"column:fup_quota_used;default:0"`
	FUPResetDate        *time.Time     `gorm:"column:fup_reset_date;type:date"`
	IsFUPActive         bool           `gorm:"column:is_fup_active;default:false"`
	CustomerNotes       string         `gorm:"column:customer_notes;type:text"`
	CreatedAt           time.Time      `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik       *Mikrotik        `gorm:"foreignKey:MikrotikID"`
	Package        *Package         `gorm:"foreignKey:PackageID"`
	PPPoEProfile   *MikrotikProfile `gorm:"foreignKey:PPPoEProfileID"`
	HotspotProfile *MikrotikProfile `gorm:"foreignKey:HotspotProfileID"`
	ODP            *ODP             `gorm:"foreignKey:ODPID"`
	Invoices       []Invoice        `gorm:"foreignKey:CustomerID"`
	CollectorPayments []CollectorPayment `gorm:"foreignKey:CustomerID"`
	AgentMonthlyPayments []AgentMonthlyPayment `gorm:"foreignKey:CustomerID"`
	AgentPayments       []AgentPayment        `gorm:"foreignKey:CustomerID"`
	CableRoutes         []CableRoute          `gorm:"foreignKey:CustomerID"`
	ONUDevices          []ONUDevice           `gorm:"foreignKey:CustomerID"`
}

func (Customer) TableName() string { return "customers" }