// pkg/cache/cache.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache interface untuk caching operations
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	Clear(ctx context.Context, pattern string) error
	GetStats(ctx context.Context) (*Stats, error)
}

// RedisCache implementasi Cache dengan Redis
type RedisCache struct {
	client *redis.Client
	prefix string
}

// Config untuk Redis cache
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Prefix   string
}

// Stats statistik cache
type Stats struct {
	Keys         int64     `json:"keys"`
	UsedMemory   string    `json:"used_memory"`
	HitRate      float64   `json:"hit_rate,omitempty"`
	RetrievedAt  time.Time `json:"retrieved_at"`
}

// NewRedisCache membuat instance baru Redis cache
func NewRedisCache(cfg Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

// prefixKey menambahkan prefix pada key
func (r *RedisCache) prefixKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return r.prefix + ":" + key
}

// Get mengambil value dari cache
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	prefixedKey := r.prefixKey(key)
	
	val, err := r.client.Get(ctx, prefixedKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss: key not found")
	}
	if err != nil {
		return fmt.Errorf("redis get error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("unmarshal cache value: %w", err)
	}

	return nil
}

// Set menyimpan value ke cache dengan TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	prefixedKey := r.prefixKey(key)
	
	jsonVal, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}

	if err := r.client.Set(ctx, prefixedKey, jsonVal, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

// Delete menghapus key dari cache
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	if err := r.client.Del(ctx, prefixedKeys...).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}

	return nil
}

// Exists mengecek apakah key ada di cache
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.prefixKey(key)
	}

	count, err := r.client.Exists(ctx, prefixedKeys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists error: %w", err)
	}

	return count, nil
}

// Clear menghapus semua key dengan pattern tertentu
func (r *RedisCache) Clear(ctx context.Context, pattern string) error {
	prefixedPattern := r.prefixKey(pattern)
	
	iter := r.client.Scan(ctx, 0, prefixedPattern, 0).Iterator()
	keys := make([]string, 0)
	
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis scan error: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("redis delete error: %w", err)
		}
	}

	return nil
}

// GetStats mengambil statistik cache
func (r *RedisCache) GetStats(ctx context.Context) (*Stats, error) {
	_, err := r.client.Info(ctx, "stats", "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("redis info error: %w", err)
	}

	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis dbsize error: %w", err)
	}

	// Parse used memory dari info string (simplified)
	usedMemory := "N/A"
	
	return &Stats{
		Keys:        dbSize,
		UsedMemory:  usedMemory,
		RetrievedAt: time.Now(),
	}, nil
}

// Close menutup koneksi Redis
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetClient mengembalikan Redis client (untuk advanced usage)
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

