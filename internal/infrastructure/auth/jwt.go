package auth

import (
	"errors"
	"mikrobill/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims contains the JWT payload
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // Single role instead of array
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token
func (m *JWTManager) GenerateToken(userID int64, email string, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.tokenDuration)

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role, // Single role
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates and parses a JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, utils.ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, utils.ErrTokenExpired
		}
		return nil, utils.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, utils.ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (m *JWTManager) RefreshToken(oldToken string) (string, int64, error) {
	claims, err := m.ValidateToken(oldToken)
	if err != nil {
		return "", 0, err
	}

	// Generate new token with same claims but new expiration
	return m.GenerateToken(claims.UserID, claims.Email, claims.Role)
}

// GetTokenDuration returns the configured token duration
func (m *JWTManager) GetTokenDuration() time.Duration {
	return m.tokenDuration
}
