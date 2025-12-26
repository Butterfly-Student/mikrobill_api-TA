// ============================================

// File: internal/port/service/jwt.go
package service

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
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, tokenDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a new JWT token
func (s *JWTService) GenerateToken(userID int64, email string, role string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.tokenDuration)

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, utils.ErrInvalidToken
			}
			return []byte(s.secretKey), nil
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
func (s *JWTService) RefreshToken(oldToken string) (string, int64, error) {
	claims, err := s.ValidateToken(oldToken)
	if err != nil {
		return "", 0, err
	}

	return s.GenerateToken(claims.UserID, claims.Email, claims.Role)
}

// GetTokenDuration returns the configured token duration
func (s *JWTService) GetTokenDuration() time.Duration {
	return s.tokenDuration
}

