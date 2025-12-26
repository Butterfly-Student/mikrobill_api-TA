package entity

import "time"

// VoucherUsageLog represents the voucher_usage_logs table
type VoucherUsageLog struct {
	ID             string     `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	VoucherID      string     `gorm:"column:voucher_id;type:uuid;not null"`
	EventType      string     `gorm:"column:event_type;type:varchar(50);not null"`
	IPAddress      string     `gorm:"column:ip_address;type:inet"`
	MACAddress     string     `gorm:"column:mac_address;type:macaddr"`
	SessionID      string     `gorm:"column:session_id;type:varchar(100)"`
	UploadBytes    int64      `gorm:"column:upload_bytes;default:0"`
	DownloadBytes  int64      `gorm:"column:download_bytes;default:0"`
	TotalBytes     int64      `gorm:"column:total_bytes;default:0"`
	SessionTime    int        `gorm:"column:session_time"`
	EventTime      time.Time  `gorm:"column:event_time;type:timestamptz;default:CURRENT_TIMESTAMP"`

	// Relations
	Voucher *HotspotVoucher `gorm:"foreignKey:VoucherID"`
}

func (VoucherUsageLog) TableName() string { return "voucher_usage_logs" }