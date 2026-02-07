package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/logger"
)

const (
	CacheKeyRouterList      = "router:list:%s"
	CacheKeyCustomerProfile = "customer:profile:%s"
	CacheKeyUserPermissions = "user:permissions:%s"
	CacheKeyTenantConfig    = "tenant:config:%s"
)

const (
	RouterListTTL      = 5 * time.Minute
	CustomerProfileTTL = 10 * time.Minute
	UserPermissionsTTL = 15 * time.Minute
	TenantConfigTTL    = 30 * time.Minute
)

type CacheService struct {
	cachePort outbound_port.CachePort
}

func NewCacheService(cachePort outbound_port.CachePort) *CacheService {
	return &CacheService{
		cachePort: cachePort,
	}
}

func (s *CacheService) CacheRouterList(ctx context.Context, tenantID string, routers []model.Mikrotik) error {
	key := fmt.Sprintf(CacheKeyRouterList, tenantID)

	data, err := json.Marshal(routers)
	if err != nil {
		return fmt.Errorf("failed to marshal router list: %w", err)
	}

	logger.GetLogger().Info("Caching router list",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(routers)))

	pppCache := s.cachePort.PPP()
	if err := pppCache.SetCached(ctx, key, string(data), RouterListTTL); err != nil {
		return fmt.Errorf("failed to cache router list: %w", err)
	}

	return nil
}

func (s *CacheService) GetRouterList(ctx context.Context, tenantID string) ([]model.Mikrotik, error) {
	key := fmt.Sprintf(CacheKeyRouterList, tenantID)

	pppCache := s.cachePort.PPP()
	data, err := pppCache.GetCached(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get router list from cache: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var routers []model.Mikrotik
	if err := json.Unmarshal([]byte(data), &routers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal router list: %w", err)
	}

	logger.GetLogger().Debug("Router list cache hit",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(routers)))

	return routers, nil
}

func (s *CacheService) InvalidateRouterList(ctx context.Context, tenantID string) error {
	key := fmt.Sprintf(CacheKeyRouterList, tenantID)

	logger.GetLogger().Info("Invalidating router list cache",
		zap.String("tenant_id", tenantID))

	pppCache := s.cachePort.PPP()
	return pppCache.InvalidateCache(ctx, key)
}

func (s *CacheService) CacheCustomerProfile(ctx context.Context, customerID uuid.UUID, profile interface{}) error {
	key := fmt.Sprintf(CacheKeyCustomerProfile, customerID.String())

	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal customer profile: %w", err)
	}

	logger.GetLogger().Debug("Caching customer profile",
		zap.String("customer_id", customerID.String()))

	pppCache := s.cachePort.PPP()
	if err := pppCache.SetCached(ctx, key, string(data), CustomerProfileTTL); err != nil {
		return fmt.Errorf("failed to cache customer profile: %w", err)
	}

	return nil
}

func (s *CacheService) GetCustomerProfile(ctx context.Context, customerID uuid.UUID) (map[string]interface{}, error) {
	key := fmt.Sprintf(CacheKeyCustomerProfile, customerID.String())

	pppCache := s.cachePort.PPP()
	data, err := pppCache.GetCached(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer profile from cache: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var profile map[string]interface{}
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal customer profile: %w", err)
	}

	logger.GetLogger().Debug("Customer profile cache hit",
		zap.String("customer_id", customerID.String()))

	return profile, nil
}

func (s *CacheService) InvalidateCustomerProfile(ctx context.Context, customerID uuid.UUID) error {
	key := fmt.Sprintf(CacheKeyCustomerProfile, customerID.String())

	logger.GetLogger().Debug("Invalidating customer profile cache",
		zap.String("customer_id", customerID.String()))

	pppCache := s.cachePort.PPP()
	return pppCache.InvalidateCache(ctx, key)
}

func (s *CacheService) CacheUserPermissions(ctx context.Context, userID string, permissions []string) error {
	key := fmt.Sprintf(CacheKeyUserPermissions, userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal user permissions: %w", err)
	}

	logger.GetLogger().Debug("Caching user permissions",
		zap.String("user_id", userID),
		zap.Int("count", len(permissions)))

	pppCache := s.cachePort.PPP()
	if err := pppCache.SetCached(ctx, key, string(data), UserPermissionsTTL); err != nil {
		return fmt.Errorf("failed to cache user permissions: %w", err)
	}

	return nil
}

func (s *CacheService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	key := fmt.Sprintf(CacheKeyUserPermissions, userID)

	pppCache := s.cachePort.PPP()
	data, err := pppCache.GetCached(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions from cache: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var permissions []string
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user permissions: %w", err)
	}

	logger.GetLogger().Debug("User permissions cache hit",
		zap.String("user_id", userID),
		zap.Int("count", len(permissions)))

	return permissions, nil
}

func (s *CacheService) InvalidateUserPermissions(ctx context.Context, userID string) error {
	key := fmt.Sprintf(CacheKeyUserPermissions, userID)

	logger.GetLogger().Debug("Invalidating user permissions cache",
		zap.String("user_id", userID))

	pppCache := s.cachePort.PPP()
	return pppCache.InvalidateCache(ctx, key)
}

func (s *CacheService) CacheTenantConfig(ctx context.Context, tenantID string, config map[string]interface{}) error {
	key := fmt.Sprintf(CacheKeyTenantConfig, tenantID)

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant config: %w", err)
	}

	logger.GetLogger().Info("Caching tenant config",
		zap.String("tenant_id", tenantID))

	pppCache := s.cachePort.PPP()
	if err := pppCache.SetCached(ctx, key, string(data), TenantConfigTTL); err != nil {
		return fmt.Errorf("failed to cache tenant config: %w", err)
	}

	return nil
}

func (s *CacheService) GetTenantConfig(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	key := fmt.Sprintf(CacheKeyTenantConfig, tenantID)

	pppCache := s.cachePort.PPP()
	data, err := pppCache.GetCached(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant config from cache: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tenant config: %w", err)
	}

	logger.GetLogger().Debug("Tenant config cache hit",
		zap.String("tenant_id", tenantID))

	return config, nil
}

func (s *CacheService) InvalidateTenantConfig(ctx context.Context, tenantID string) error {
	key := fmt.Sprintf(CacheKeyTenantConfig, tenantID)

	logger.GetLogger().Info("Invalidating tenant config cache",
		zap.String("tenant_id", tenantID))

	pppCache := s.cachePort.PPP()
	return pppCache.InvalidateCache(ctx, key)
}

func (s *CacheService) InvalidateAllForTenant(ctx context.Context, tenantID string) error {
	logger.GetLogger().Info("Invalidating all tenant caches",
		zap.String("tenant_id", tenantID))

	keys := []string{
		fmt.Sprintf(CacheKeyRouterList, tenantID),
		fmt.Sprintf(CacheKeyTenantConfig, tenantID),
	}

	pppCache := s.cachePort.PPP()
	for _, key := range keys {
		if err := pppCache.InvalidateCache(ctx, key); err != nil {
			logger.GetLogger().Error("Failed to invalidate cache",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	return nil
}
