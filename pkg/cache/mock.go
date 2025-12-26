// pkg/cache/mock.go
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockCache implementasi in-memory cache untuk testing
type MockCache struct {
	mu    sync.RWMutex
	data  map[string]cacheItem
}

type cacheItem struct {
	value     interface{}
	expiresAt *time.Time
}

// NewMockCache membuat instance mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]cacheItem),
	}
}

// Get mengambil value dari mock cache
func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return fmt.Errorf("cache miss: key not found")
	}

	if item.expiresAt != nil && time.Now().After(*item.expiresAt) {
		delete(m.data, key)
		return fmt.Errorf("cache miss: key expired")
	}

	// Type assertion untuk copy value
	switch v := dest.(type) {
	case *string:
		if s, ok := item.value.(string); ok {
			*v = s
		}
	case *int:
		if i, ok := item.value.(int); ok {
			*v = i
		}
	case *interface{}:
		*v = item.value
	default:
		return fmt.Errorf("unsupported destination type")
	}

	return nil
}

// Set menyimpan value ke mock cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		exp := time.Now().Add(ttl)
		expiresAt = &exp
	}

	m.data[key] = cacheItem{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

// Delete menghapus key dari mock cache
func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
	}

	return nil
}

// Exists mengecek apakah key ada
func (m *MockCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			count++
		}
	}

	return count, nil
}

// Clear menghapus semua key dengan pattern
func (m *MockCache) Clear(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simple implementation: clear all if pattern is "*"
	if pattern == "*" {
		m.data = make(map[string]cacheItem)
	}

	return nil
}

// GetStats mengambil statistik mock cache
func (m *MockCache) GetStats(ctx context.Context) (*Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &Stats{
		Keys:        int64(len(m.data)),
		UsedMemory:  "N/A (mock)",
		RetrievedAt: time.Now(),
	}, nil
}