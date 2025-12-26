package model

// ========== IP POOL DTOs ==========

// IPPoolRequest untuk membuat IP Pool baru
type IPPoolRequest struct {
	Name     string `json:"name" binding:"required"`     // Nama pool (required)
	Ranges   string `json:"ranges" binding:"required"`   // Range IP, contoh: "192.168.1.100-192.168.1.200" (required)
	Comment  string `json:"comment"`                     // Komentar untuk pool
	NextPool string `json:"next_pool"`                   // Pool berikutnya jika pool ini penuh
}

// IPPoolUpdateRequest untuk update IP Pool
type IPPoolUpdateRequest struct {
	Name     string `json:"name"`      // Nama pool baru
	Ranges   string `json:"ranges"`    // Range IP baru
	Comment  string `json:"comment"`   // Komentar baru
	NextPool string `json:"next_pool"` // Pool berikutnya
}

// IPPoolResponse adalah response untuk operasi IP Pool
type IPPoolResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// IPPoolData adalah struktur detail IP Pool
type IPPoolData struct {
	ID       string `json:".id"`
	Name     string `json:"name"`
	Ranges   string `json:"ranges"`
	Comment  string `json:"comment,omitempty"`
	NextPool string `json:"next-pool,omitempty"`
}

// IPPoolUsedData adalah struktur untuk IP yang sedang digunakan
type IPPoolUsedData struct {
	ID      string `json:".id"`
	Address string `json:"address"`
	Pool    string `json:"pool"`
	Info    string `json:"info,omitempty"`
}
