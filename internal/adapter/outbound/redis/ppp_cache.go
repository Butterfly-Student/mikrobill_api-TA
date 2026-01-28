package redis_outbound_adapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/redis"

	"github.com/palantir/stacktrace"
)

type pppCacheAdapter struct{}

func NewPPPCacheAdapter() outbound_port.PPPCachePort {
	return &pppCacheAdapter{}
}

// ============= PPP Active =============

func (a *pppCacheAdapter) GetPPPActive(ctx context.Context, tenantId string) ([]model.PPPActive, string, error) {
	key := fmt.Sprintf(CachePPPActiveKey, tenantId)
	hashKey := key + CacheHashSuffix

	// Get data
	dataStr, err := redis.Get(ctx, key)
	if err != nil {
		return nil, "", err // Cache miss
	}

	// Get hash
	hash, err := redis.Get(ctx, hashKey)
	if err != nil {
		hash = "" // Hash might not exist
	}

	// Unmarshal
	var data []model.PPPActive
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, "", stacktrace.Propagate(err, "failed to unmarshal cached PPP active")
	}

	return data, hash, nil
}

func (a *pppCacheAdapter) SetPPPActive(ctx context.Context, tenantId string, data []model.PPPActive, hash string) error {
	key := fmt.Sprintf(CachePPPActiveKey, tenantId)
	hashKey := key + CacheHashSuffix

	// Marshal data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal PPP active data")
	}

	// Set data with TTL (5 minutes for frequently changing data)
	if err := redis.SetWithExpiry(ctx, key, string(dataBytes), 5*time.Minute); err != nil {
		return stacktrace.Propagate(err, "failed to set PPP active cache")
	}

	// Set hash
	if err := redis.SetWithExpiry(ctx, hashKey, hash, 5*time.Minute); err != nil {
		return stacktrace.Propagate(err, "failed to set PPP active hash")
	}

	return nil
}

// ============= PPP Inactive =============

func (a *pppCacheAdapter) GetPPPInactive(ctx context.Context, tenantId string) ([]model.PPPActive, string, error) {
	key := fmt.Sprintf(CachePPPInactiveKey, tenantId)
	hashKey := key + CacheHashSuffix

	dataStr, err := redis.Get(ctx, key)
	if err != nil {
		return nil, "", err
	}

	hash, err := redis.Get(ctx, hashKey)
	if err != nil {
		hash = ""
	}

	var data []model.PPPActive
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, "", stacktrace.Propagate(err, "failed to unmarshal cached PPP inactive")
	}

	return data, hash, nil
}

func (a *pppCacheAdapter) SetPPPInactive(ctx context.Context, tenantId string, data []model.PPPActive, hash string) error {
	key := fmt.Sprintf(CachePPPInactiveKey, tenantId)
	hashKey := key + CacheHashSuffix

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to marshal PPP inactive data")
	}

	if err := redis.SetWithExpiry(ctx, key, string(dataBytes), 5*time.Minute); err != nil {
		return stacktrace.Propagate(err, "failed to set PPP inactive cache")
	}

	if err := redis.SetWithExpiry(ctx, hashKey, hash, 5*time.Minute); err != nil {
		return stacktrace.Propagate(err, "failed to set PPP inactive hash")
	}

	return nil
}

// ============= Generic Cache Methods =============

func (a *pppCacheAdapter) GetCached(ctx context.Context, key string) (string, error) {
	return redis.Get(ctx, key)
}

func (a *pppCacheAdapter) SetCached(ctx context.Context, key, value string, ttl time.Duration) error {
	return redis.SetWithExpiry(ctx, key, value, ttl)
}

func (a *pppCacheAdapter) InvalidateCache(ctx context.Context, key string) error {
	return redis.Del(ctx, key)
}

// ============= Utility Functions =============

// GenerateHash creates SHA256 hash dari data untuk change detection
func GenerateHash(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:]), nil
}
