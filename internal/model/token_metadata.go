package model

import "time"

// RefreshTokenMetadata stores comprehensive information about a refresh token
type RefreshTokenMetadata struct {
	UserID          string    `json:"user_id"`
	TokenHash       string    `json:"token_hash"` // SHA-256 hash of actual token
	IssuedAt        time.Time `json:"issued_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	LastUsedAt      time.Time `json:"last_used_at"`
	RotationCount   int       `json:"rotation_count"`
	AbsoluteExpiry  time.Time `json:"absolute_expiry"`   // Cannot extend beyond this
	ParentTokenHash *string   `json:"parent_token_hash"` // For reuse detection
}

// RefreshTokenResponse includes rotation information
type RefreshTokenResponse struct {
	AccessToken      string  `json:"access_token"`
	RefreshToken     *string `json:"refresh_token,omitempty"` // Only present if rotated
	TokenType        string  `json:"token_type"`
	ExpiresIn        int64   `json:"expires_in"`
	RefreshExpiresIn *int64  `json:"refresh_expires_in,omitempty"` // Only if rotated
	Rotation         bool    `json:"rotation"`
	User             *User   `json:"user"`
}

// Enhanced LoginResponse with absolute expiry
type EnhancedLoginResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	TokenType         string `json:"token_type"`
	ExpiresIn         int64  `json:"expires_in"`
	RefreshExpiresIn  int64  `json:"refresh_expires_in"`
	AbsoluteExpiresIn int64  `json:"absolute_expires_in"` // Seconds until absolute expiry
	User              *User  `json:"user"`
}
