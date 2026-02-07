package jwt_adapter

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/logger"
)

type jwtAdapter struct {
	secret string
}

func NewJWTAdapter(secret string) outbound_port.JWTPort {
	if secret == "" {
		logger.GetLogger().Warn("JWT_SECRET is empty, using default (not recommended for production)")
		secret = "default_secret_please_change"
	}
	if len(secret) < 32 {
		logger.GetLogger().Warn("JWT_SECRET too weak, must be at least 32 characters")
	}
	return &jwtAdapter{secret: secret}
}

func (a *jwtAdapter) GenerateAccessToken(ctx context.Context, userID string, username string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"exp":      0,
	}

	expTime := int(15 * 60)
	claims["exp"] = expTime

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (a *jwtAdapter) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(string); ok {
			return map[string]interface{}{
				"user_id": sub,
			}, nil
		}
	}

	return nil, model.ErrTokenInvalid
}

func (a *jwtAdapter) ParseToken(tokenString string) (interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, model.ErrTokenInvalid
}
