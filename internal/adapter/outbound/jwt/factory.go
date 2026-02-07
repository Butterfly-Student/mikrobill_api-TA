package jwt_adapter

import (
	"os"

	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/logger"
)

func NewJWTAdapterFactory() outbound_port.JWTPort {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.GetLogger().Warn("JWT_SECRET environment variable not set")
		return nil
	}
	return NewJWTAdapter(secret)
}
