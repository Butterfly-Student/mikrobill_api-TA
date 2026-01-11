package model

import (
	"time"

	"github.com/google/uuid"
)

type MikrotikStatus string

const (
	MikrotikStatusOnline  MikrotikStatus = "online"
	MikrotikStatusOffline MikrotikStatus = "offline"
	MikrotikStatusError   MikrotikStatus = "error"
)

type Mikrotik struct {
	ID                   uuid.UUID      `json:"id" db:"id"`
	Name                 string         `json:"name" db:"name"`
	Host                 string         `json:"host" db:"host"`
	Port                 int            `json:"port" db:"port"`
	APIUsername          string         `json:"api_username" db:"api_username"`
	APIEncryptedPassword *string        `json:"api_encrypted_password,omitempty" db:"api_encrypted_password"`
	Keepalive            bool           `json:"keepalive" db:"keepalive"`
	Timeout              int            `json:"timeout" db:"timeout"`
	Location             *string        `json:"location,omitempty" db:"location"`
	Description          *string        `json:"description,omitempty" db:"description"`
	IsActive             bool           `json:"is_active" db:"is_active"`
	Status               MikrotikStatus `json:"status" db:"status"`
	LastSync             *time.Time     `json:"last_sync,omitempty" db:"last_sync"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

type MikrotikInput struct {
	Name                 string  `json:"name" binding:"required"`
	Host                 string  `json:"host" binding:"required"`
	Port                 int     `json:"port"`
	APIUsername          string  `json:"api_username" binding:"required"`
	APIEncryptedPassword *string `json:"api_encrypted_password,omitempty"`
	Keepalive            *bool   `json:"keepalive,omitempty"`
	Timeout              *int    `json:"timeout,omitempty"`
	Location             *string `json:"location,omitempty"`
	Description          *string `json:"description,omitempty"`
}

type MikrotikUpdateInput struct {
	Name                 *string `json:"name,omitempty"`
	Host                 *string `json:"host,omitempty"`
	Port                 *int    `json:"port,omitempty"`
	APIUsername          *string `json:"api_username,omitempty"`
	APIEncryptedPassword *string `json:"api_encrypted_password,omitempty"`
	Keepalive            *bool   `json:"keepalive,omitempty"`
	Timeout              *int    `json:"timeout,omitempty"`
	Location             *string `json:"location,omitempty"`
	Description          *string `json:"description,omitempty"`
}

type MikrotikFilter struct {
	IDs      []uuid.UUID      `json:"ids,omitempty"`
	Hosts    []string         `json:"hosts,omitempty"`
	Statuses []MikrotikStatus `json:"statuses,omitempty"`
	IsActive *bool            `json:"is_active,omitempty"`
}

func MikrotikPrepare(input *MikrotikInput) {
	if input.Port == 0 {
		input.Port = 8728
	}
	if input.Keepalive == nil {
		keepalive := true
		input.Keepalive = &keepalive
	}
	if input.Timeout == nil {
		timeout := 300000
		input.Timeout = &timeout
	}
}

func (f MikrotikFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Hosts) == 0 && len(f.Statuses) == 0 && f.IsActive == nil
}
