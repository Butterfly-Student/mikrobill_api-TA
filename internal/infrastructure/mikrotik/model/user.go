package model

// UserRequest adalah request untuk membuat/update user
type UserRequest struct {
	Server      string `json:"server,omitempty"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password,omitempty"`
	Profile     string `json:"profile,omitempty"`
	MacAddress  string `json:"macAddress,omitempty"`
	TimeLimit   string `json:"timeLimit,omitempty"`
	DataLimit   string `json:"dataLimit,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// UserUpdateRequest untuk update user
type UserUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	Password   string `json:"password,omitempty"`
	Profile    string `json:"profile,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
	TimeLimit  string `json:"timeLimit,omitempty"`
	DataLimit  string `json:"dataLimit,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Disabled   *bool  `json:"disabled,omitempty"`
}

// UserResponse adalah response dari operasi user
type UserResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UserData adalah data user dari MikroTik
type UserData struct {
	ID              string `json:".id"`
	Name            string `json:"name"`
	Password        string `json:"password,omitempty"`
	Profile         string `json:"profile,omitempty"`
	Server          string `json:"server,omitempty"`
	MacAddress      string `json:"mac-address,omitempty"`
	LimitUptime     string `json:"limit-uptime,omitempty"`
	LimitBytesTotal string `json:"limit-bytes-total,omitempty"`
	Comment         string `json:"comment,omitempty"`
	Disabled        string `json:"disabled,omitempty"`
}

// ActiveUserData adalah data user aktif
type ActiveUserData struct {
	ID         string `json:".id"`
	User       string `json:"user"`
	Address    string `json:"address"`
	MacAddress string `json:"mac-address"`
	LoginBy    string `json:"login-by"`
	Uptime     string `json:"uptime"`
	BytesIn    string `json:"bytes-in"`
	BytesOut   string `json:"bytes-out"`
	Server     string `json:"server"`
}

// HostData adalah data host
type HostData struct {
	ID         string `json:".id"`
	MacAddress string `json:"mac-address"`
	Address    string `json:"address"`
	ToAddress  string `json:"to-address"`
	Server     string `json:"server"`
	Uptime     string `json:"uptime"`
	BytesIn    string `json:"bytes-in"`
	BytesOut   string `json:"bytes-out"`
}