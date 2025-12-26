package model

// ProfileRequest adalah request untuk membuat/update profile
type ProfileRequest struct {
	Name              string      `json:"name" binding:"required"`
	SharedUsers       *int        `json:"sharedUsers,omitempty"`
	RateLimit         string      `json:"rateLimit,omitempty"`
	ExpMode           ExpireMode  `json:"expMode,omitempty"`
	Validity          string      `json:"validity,omitempty"`
	Price             string      `json:"price,omitempty"`
	SellingPrice      string      `json:"sellingPrice,omitempty"`
	AddressPool       string      `json:"addressPool,omitempty"`
	LockUser          LockStatus  `json:"lockUser,omitempty"`
	LockServer        LockStatus  `json:"lockServer,omitempty"`
	ParentQueue       string      `json:"parentQueue,omitempty"`
	StatusAutoRefresh string      `json:"statusAutorefresh,omitempty"`
	OnLogin           string      `json:"onLogin,omitempty"`
	Bandwidth         string      `json:"bandwidth,omitempty"`
	SessionTimeout    string      `json:"sessionTimeout,omitempty"`
	IdleTimeout       string      `json:"idleTimeout,omitempty"`
	DownloadLimit     string      `json:"downloadLimit,omitempty"`
	UploadLimit       string      `json:"uploadLimit,omitempty"`
	MaxSessions       string      `json:"maxSessions,omitempty"`
}

// ProfileUpdateRequest untuk update profile
type ProfileUpdateRequest struct {
	RateLimit   string      `json:"rateLimit,omitempty"`
	SharedUsers *int        `json:"sharedUsers,omitempty"`
	AddressPool string      `json:"addressPool,omitempty"`
	ParentQueue string      `json:"parentQueue,omitempty"`
	ExpMode     ExpireMode  `json:"expMode,omitempty"`
	Price       string      `json:"price,omitempty"`
	LockUser    LockStatus  `json:"lockUser,omitempty"`
	LockServer  LockStatus  `json:"lockServer,omitempty"`
	Validity    string      `json:"validity,omitempty"`
	Name        string      `json:"name,omitempty"`
}

// ProfileResponse adalah response dari operasi profile
type ProfileResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ProfileData adalah data profile dari MikroTik
type ProfileData struct {
	ID                string `json:".id"`
	Name              string `json:"name"`
	AddressPool       string `json:"address-pool,omitempty"`
	RateLimit         string `json:"rate-limit,omitempty"`
	SharedUsers       string `json:"shared-users,omitempty"`
	ParentQueue       string `json:"parent-queue,omitempty"`
	OnLogin           string `json:"on-login,omitempty"`
	StatusAutoRefresh string `json:"status-autorefresh,omitempty"`
}