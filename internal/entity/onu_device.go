package entity

import "time"

// ONUDevice represents the onu_devices table
type ONUDevice struct {
	ID              string       `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	MikrotikID      string       `gorm:"column:mikrotik_id;type:uuid;not null"`
	Name            string       `gorm:"column:name;type:varchar(100);not null"`
	SerialNumber    string       `gorm:"column:serial_number;type:varchar(100)"`
	MACAddress      string       `gorm:"column:mac_address;type:macaddr"`
	IPAddress       string       `gorm:"column:ip_address;type:inet"`
	Status          DeviceStatus `gorm:"column:status;type:device_status;default:'online'"`
	Latitude        float64      `gorm:"column:latitude;type:decimal(10,8)"`
	Longitude       float64      `gorm:"column:longitude;type:decimal(11,8)"`
	CustomerID      *string      `gorm:"column:customer_id;type:uuid"`
	ODPID           *string      `gorm:"column:odp_id;type:uuid"`
	SSID            string       `gorm:"column:ssid;type:varchar(50)"`
	Password        string       `gorm:"column:password;type:varchar(100)"`
	Model           string       `gorm:"column:model;type:varchar(100)"`
	FirmwareVersion string       `gorm:"column:firmware_version;type:varchar(50)"`
	CreatedAt       time.Time    `gorm:"column:created_at;type:timestamptz;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time    `gorm:"column:updated_at;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Mikrotik *Mikrotik `gorm:"foreignKey:MikrotikID"`
	Customer *Customer `gorm:"foreignKey:CustomerID"`
	ODP      *ODP      `gorm:"foreignKey:ODPID"`
}

func (ONUDevice) TableName() string { return "onu_devices" }