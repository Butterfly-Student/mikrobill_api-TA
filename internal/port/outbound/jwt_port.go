package outbound_port

//go:generate mockgen -source=jwt_port.go -destination=./../../../tests/mocks/port/mock_jwt_port.go

import (
	"context"
)

// JWTPort handles JWT token generation and validation
type JWTPort interface {
	// GenerateAccessToken creates a short-lived access token
	GenerateAccessToken(ctx context.Context, userID string, username string, role string) (string, error)

	// ValidateToken validates a JWT token and returns claims
	ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error)

	// ParseToken parses a token without validation
	ParseToken(tokenString string) (interface{}, error)
}
