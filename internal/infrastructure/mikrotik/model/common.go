package model

// Response adalah struktur response umum
type Response struct {
	Message string      `json:"message"` // "success" atau "error"
	Data    interface{} `json:"data"`
}

// ErrorData untuk response error
type ErrorData struct {
	Error string `json:"error"`
}

// SimpleSuccessData untuk response sukses sederhana
type SimpleSuccessData struct {
	Created bool   `json:"created,omitempty"`
	Updated bool   `json:"updated,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
}

// ExpireMode tipe untuk mode expire
type ExpireMode string

const (
	ExpireModeNone ExpireMode = "0"
	ExpireModeNTF  ExpireMode = "ntf"
	ExpireModeNTFC ExpireMode = "ntfc"
	ExpireModeREM  ExpireMode = "rem"
	ExpireModeREMC ExpireMode = "remc"
)

// LockStatus untuk status lock
type LockStatus string

const (
	LockDisable LockStatus = "Disable"
	LockEnable  LockStatus = "Enable"
)