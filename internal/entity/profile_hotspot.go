package entity

type ProfileHotspot struct {
	ProfileID           string `gorm:"primaryKey;type:uuid;column:profile_id"`
	SharedUsers         int    `gorm:"column:shared_users;default:1"`
	HotspotAddressPool  string `gorm:"column:hotspot_address_pool;type:varchar(50)"`
	TransparentProxy    bool   `gorm:"column:transparent_proxy;default:false"`
	SMTPServer          string `gorm:"column:smtp_server;type:varchar(100)"`
	HTTPProxy           string `gorm:"column:http_proxy;type:varchar(100)"`
	HTTPCookieLifetime  string `gorm:"column:http_cookie_lifetime;type:varchar(20)"`
	MacAuth             bool   `gorm:"column:mac_auth;default:false"`
	MacAuthMode         string `gorm:"column:mac_auth_mode;type:varchar(20);default:'none'"`
	TrialUserProfile    string `gorm:"column:trial_user_profile;type:varchar(50)"`
	MacCookieTimeout    string `gorm:"column:mac_cookie_timeout;type:varchar(20)"`
	LoginTimeout        string `gorm:"column:login_timeout;type:varchar(20)"`

	// Relations
	Profile *MikrotikProfile `gorm:"foreignKey:ProfileID"`
}

func (ProfileHotspot) TableName() string { return "mikrotik_profile_hotspot" }