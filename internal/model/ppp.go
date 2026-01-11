package model

import "github.com/google/uuid"

type PPPSecret struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Password      string    `json:"password"`
	Profile       string    `json:"profile"`
	Service       string    `json:"service"` // pppoe, any, etc.
	RemoteAddress string    `json:"remote_address,omitempty"`
	LocalAddress  string    `json:"local_address,omitempty"`
	Comment       string    `json:"comment,omitempty"`
	Disabled      bool      `json:"disabled"`
}

type PPPProfile struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	LocalAddress  string    `json:"local_address,omitempty"`
	RemoteAddress string    `json:"remote_address,omitempty"`
	RateLimit     string    `json:"rate_limit,omitempty"`
	Comment       string    `json:"comment,omitempty"`
	OnlyOne       string    `json:"only_one,omitempty"` // default, yes, no
}

type PPPSecretInput struct {
	Name          string `json:"name" binding:"required"`
	Password      string `json:"password" binding:"required"`
	Profile       string `json:"profile" binding:"required"`
	Service       string `json:"service" binding:"required"`
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	Comment       string `json:"comment"`
}

type PPPSecretUpdateInput struct {
	Password      *string `json:"password"`
	Profile       *string `json:"profile"`
	Service       *string `json:"service"`
	LocalAddress  *string `json:"local_address"`
	RemoteAddress *string `json:"remote_address"`
	Comment       *string `json:"comment"`
	Disabled      *bool   `json:"disabled"`
}

type PPPProfileInput struct {
	Name          string `json:"name" binding:"required"`
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	RateLimit     string `json:"rate_limit"`
	OnlyOne       string `json:"only_one"`
	Comment       string `json:"comment"`
}
