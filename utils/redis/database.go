package redis

import (
	"context"
	"os"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var dbClient *redis.Client

func InitDatabase() {
	addr := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	pass := os.Getenv("CACHE_PASSWORD")
	if port == "" {
		port = "6379"
	}
	dbClient = redis.NewClient(&redis.Options{
		Addr:     addr + ":" + port,
		Password: pass,
	})
}

func Set(ctx context.Context, key string, value interface{}) error {
	return dbClient.Set(ctx, key, value, 24*60*60*1e9).Err() // 1 day in nanoseconds
}

func SetWithExpiry(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return dbClient.Set(ctx, key, value, ttl).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return dbClient.Get(ctx, key).Result()
}

func Del(ctx context.Context, key string) error {
	return dbClient.Del(ctx, key).Err()
}

func SAdd(ctx context.Context, key string, value interface{}) error {
	return dbClient.SAdd(ctx, key, value).Err()
}

func SRem(ctx context.Context, key string, value interface{}) error {
	return dbClient.SRem(ctx, key, value).Err()
}

func SMembers(ctx context.Context, key string) ([]string, error) {
	return dbClient.SMembers(ctx, key).Result()
}

func Expire(ctx context.Context, key string, ttl time.Duration) error {
	return dbClient.Expire(ctx, key, ttl).Err()
}
