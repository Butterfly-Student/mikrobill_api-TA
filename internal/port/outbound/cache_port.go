package outbound_port

import (
	"context"
	"time"

	"MikrOps/internal/model"
)

//go:generate mockgen -source=cache_port.go -destination=./../../../tests/mocks/port/mock_cache_port.go

type PPPCachePort interface {
	// PPP Active/Inactive dengan hash untuk change detection
	GetPPPActive(ctx context.Context, tenantId string) (data []model.PPPActive, hash string, err error)
	SetPPPActive(ctx context.Context, tenantId string, data []model.PPPActive, hash string) error

	GetPPPInactive(ctx context.Context, tenantId string) (data []model.PPPActive, hash string, err error)
	SetPPPInactive(ctx context.Context, tenantId string, data []model.PPPActive, hash string) error

	// Generic cache methods untuk resources yang jarang berubah (profiles, pools, queues)
	GetCached(ctx context.Context, key string) (string, error)
	SetCached(ctx context.Context, key, value string, ttl time.Duration) error
	InvalidateCache(ctx context.Context, key string) error
}
