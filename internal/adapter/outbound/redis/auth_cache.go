package redis_outbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/redis"

	"github.com/palantir/stacktrace"
)

type authCacheAdapter struct {
}

func NewAuthCacheAdapter() outbound_port.AuthCachePort {
	return &authCacheAdapter{}
}

// StoreRefreshToken stores refresh token metadata in Redis
func (a *authCacheAdapter) StoreRefreshToken(ctx context.Context, metadata model.RefreshTokenMetadata) error {
	// Serialize metadata to JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal token metadata")
	}

	// Calculate TTL from expires_at
	ttl := time.Until(metadata.ExpiresAt)
	if ttl <= 0 {
		return stacktrace.NewError("token already expired")
	}

	// Store token metadata: refresh_token:{hash} -> JSON
	tokenKey := fmt.Sprintf("refresh_token:%s", metadata.TokenHash)
	if err := redis.SetWithExpiry(ctx, tokenKey, string(data), ttl); err != nil {
		return stacktrace.Propagate(err, "failed to store refresh token metadata")
	}

	// Add to user's token set: user_tokens:{userID} -> Set of hashes
	userTokensKey := fmt.Sprintf("user_tokens:%s", metadata.UserID)
	if err := redis.SAdd(ctx, userTokensKey, metadata.TokenHash); err != nil {
		return stacktrace.Propagate(err, "failed to add token to user set")
	}

	// Set TTL on user tokens set
	if err := redis.Expire(ctx, userTokensKey, ttl); err != nil {
		return stacktrace.Propagate(err, "failed to set TTL on user tokens")
	}

	return nil
}

// GetRefreshTokenMetadata retrieves token metadata by hash
func (a *authCacheAdapter) GetRefreshTokenMetadata(ctx context.Context, tokenHash string) (*model.RefreshTokenMetadata, error) {
	tokenKey := fmt.Sprintf("refresh_token:%s", tokenHash)
	data, err := redis.Get(ctx, tokenKey)
	if err != nil {
		return nil, stacktrace.NewError("refresh token not found or expired")
	}

	var metadata model.RefreshTokenMetadata
	if err := json.Unmarshal([]byte(data), &metadata); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal token metadata")
	}

	return &metadata, nil
}

// MarkTokenUsed updates last_used timestamp
func (a *authCacheAdapter) MarkTokenUsed(ctx context.Context, tokenHash string) error {
	metadata, err := a.GetRefreshTokenMetadata(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Update last_used_at
	metadata.LastUsedAt = time.Now()

	// Re-store with updated metadata
	return a.StoreRefreshToken(ctx, *metadata)
}

// InvalidateToken removes specific token
func (a *authCacheAdapter) InvalidateToken(ctx context.Context, tokenHash string) error {
	// Get metadata first to get userID
	metadata, err := a.GetRefreshTokenMetadata(ctx, tokenHash)
	if err != nil {
		// Token doesn't exist, that's okay
		return nil
	}

	// Delete token metadata
	tokenKey := fmt.Sprintf("refresh_token:%s", tokenHash)
	if err := redis.Del(ctx, tokenKey); err != nil {
		return stacktrace.Propagate(err, "failed to delete refresh token")
	}

	// Remove from user tokens set
	userTokensKey := fmt.Sprintf("user_tokens:%s", metadata.UserID)
	if err := redis.SRem(ctx, userTokensKey, tokenHash); err != nil {
		return stacktrace.Propagate(err, "failed to remove token from user set")
	}

	return nil
}

// InvalidateAllUserTokens removes all user tokens
func (a *authCacheAdapter) InvalidateAllUserTokens(ctx context.Context, userID string) error {
	userTokensKey := fmt.Sprintf("user_tokens:%s", userID)

	// Get all token hashes for the user
	hashes, err := redis.SMembers(ctx, userTokensKey)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get user tokens")
	}

	// Delete each token
	for _, hash := range hashes {
		tokenKey := fmt.Sprintf("refresh_token:%s", hash)
		if err := redis.Del(ctx, tokenKey); err != nil {
			return stacktrace.Propagate(err, "failed to delete token")
		}
	}

	// Delete the user tokens set
	if err := redis.Del(ctx, userTokensKey); err != nil {
		return stacktrace.Propagate(err, "failed to delete user tokens set")
	}

	return nil
}

// MarkTokenRotated flags token as rotated (5 min TTL for reuse detection)
func (a *authCacheAdapter) MarkTokenRotated(ctx context.Context, tokenHash string) error {
	rotatedKey := fmt.Sprintf("rotated_token:%s", tokenHash)
	// Store with 5 minute TTL
	if err := redis.SetWithExpiry(ctx, rotatedKey, "1", 5*time.Minute); err != nil {
		return stacktrace.Propagate(err, "failed to mark token as rotated")
	}
	return nil
}

// IsTokenRotated checks if token was already rotated
func (a *authCacheAdapter) IsTokenRotated(ctx context.Context, tokenHash string) (bool, error) {
	rotatedKey := fmt.Sprintf("rotated_token:%s", tokenHash)
	_, err := redis.Get(ctx, rotatedKey)
	if err != nil {
		// Key doesn't exist = not rotated
		return false, nil
	}
	return true, nil
}
