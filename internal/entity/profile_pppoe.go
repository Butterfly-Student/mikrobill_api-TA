package entity

type ProfilePPPoE struct {
	ProfileID       string `gorm:"primaryKey;type:uuid;column:profile_id"`
	LocalAddress    string `gorm:"column:local_address;type:varchar(50);not null"`
	RemoteAddress   string `gorm:"column:remote_address;type:varchar(50)"`
	AddressPool     string `gorm:"column:address_pool;type:varchar(50);not null"`
	MTU             string `gorm:"column:mtu;type:varchar(10);default:'1480'"`
	MRU             string `gorm:"column:mru;type:varchar(10);default:'1480'"`
	ServiceName     string `gorm:"column:service_name;type:varchar(50)"`
	MaxMTU          string `gorm:"column:max_mtu;type:varchar(10)"`
	MaxMRU          string `gorm:"column:max_mru;type:varchar(10)"`
	UseMPLS         bool   `gorm:"column:use_mpls;default:false"`
	UseCompression  bool   `gorm:"column:use_compression;default:false"`
	UseEncryption   bool   `gorm:"column:use_encryption;default:false"`

	// Relations
	Profile *MikrotikProfile `gorm:"foreignKey:ProfileID"`
}

func (ProfilePPPoE) TableName() string { return "mikrotik_profile_pppoe" }