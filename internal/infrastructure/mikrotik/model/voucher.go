package model

// CharType tipe untuk jenis karakter voucher
type CharType string

const (
	CharTypeLower   CharType = "lower"
	CharTypeUpper   CharType = "upper"
	CharTypeUppLow  CharType = "upplow"
	CharTypeMix     CharType = "mix"
	CharTypeMix1    CharType = "mix1"
	CharTypeMix2    CharType = "mix2"
	CharTypeNum     CharType = "num"
	CharTypeLower1  CharType = "lower1"
	CharTypeUpper1  CharType = "upper1"
	CharTypeUppLow1 CharType = "upplow1"
)

// UserType tipe untuk jenis user
type UserType string

const (
	UserTypeUP UserType = "up" // User + Password
	UserTypeVC UserType = "vc" // Voucher Code
)

// VoucherRequest adalah request untuk generate voucher
type VoucherRequest struct {
	Qty        int      `json:"qty" binding:"required,min=1"`
	Server     string   `json:"server,omitempty"`
	UserType   UserType `json:"userType" binding:"required"`
	UserLength int      `json:"userLength" binding:"required,min=1"`
	Prefix     string   `json:"prefix,omitempty"`
	CharType   CharType `json:"charType" binding:"required"`
	Profile    string   `json:"profile" binding:"required"`
	TimeLimit  string   `json:"timeLimit,omitempty"`
	DataLimit  string   `json:"dataLimit,omitempty"`
	Comment    string   `json:"comment,omitempty"`
	GenCode    string   `json:"genCode,omitempty"`
}

// VoucherResponse adalah response dari generate voucher
type VoucherResponse struct {
	Message string              `json:"message"`
	Data    VoucherResponseData `json:"data"`
}

// VoucherResponseData adalah data response voucher
type VoucherResponseData struct {
	Count   int          `json:"count,omitempty"`
	Comment string       `json:"comment,omitempty"`
	Profile string       `json:"profile,omitempty"`
	Users   []UserData   `json:"users,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// GeneratedVoucher adalah data voucher yang di-generate
type GeneratedVoucher struct {
	Username string `json:"username"`
	Password string `json:"password"`
}