package outbound_port

import (
	"context"

	"MikrOps/internal/model"
)

//go:generate mockgen -source=auth_cache.go -destination=./../../../tests/mocks/port/mock_auth_cache.go

type AuthCachePort interface {
	// StoreRefreshToken stores refresh token with full metadata
	StoreRefreshToken(ctx context.Context, metadata model.RefreshTokenMetadata) error

	// GetRefreshTokenMetadata retrieves metadata by token hash
	GetRefreshTokenMetadata(ctx context.Context, tokenHash string) (*model.RefreshTokenMetadata, error)

	// MarkTokenUsed updates last_used timestamp
	MarkTokenUsed(ctx context.Context, tokenHash string) error

	// InvalidateToken removes specific token
	InvalidateToken(ctx context.Context, tokenHash string) error

	// InvalidateAllUserTokens removes all user tokens (logout/security breach)
	InvalidateAllUserTokens(ctx context.Context, userID string) error

	// MarkTokenRotated flags token as rotated for reuse detection (5 min TTL)
	MarkTokenRotated(ctx context.Context, tokenHash string) error

	// IsTokenRotated checks if token was already rotated
	IsTokenRotated(ctx context.Context, tokenHash string) (bool, error)
}
