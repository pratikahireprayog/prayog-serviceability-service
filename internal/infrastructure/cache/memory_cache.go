package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements an in-memory cache with TTL and LRU eviction
type MemoryCache struct {
	mu             sync.RWMutex
	data           map[string]*CacheEntry
	accessOrder    []string // For LRU tracking
	config         CacheConfig
	stats          CacheStats
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
	cleanupRunning bool
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config CacheConfig) *MemoryCache {
	// Set defaults if not provided
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 1 * time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	cache := &MemoryCache{
		data:        make(map[string]*CacheEntry),
		accessOrder: make([]string, 0),
		config:      config,
		stats: CacheStats{
			MaxSize:     config.MaxSize,
			LastUpdated: time.Now(),
		},
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.startCleanup()

	return cache
}

// Get retrieves a value from cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, exists := mc.data[key]
	if !exists {
		mc.stats.Misses++
		mc.updateHitRatio()
		return nil, false
	}

	// Check if expired
	if entry.IsExpired() {
		delete(mc.data, key)
		mc.removeFromAccessOrder(key)
		mc.stats.Misses++
		mc.stats.Evictions++
		mc.updateHitRatio()
		return nil, false
	}

	// Update access information
	entry.AccessCount++
	entry.LastAccess = time.Now()
	mc.updateAccessOrder(key)

	mc.stats.Hits++
	mc.updateHitRatio()

	return entry.Value, true
}

// Set stores a value in cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Use default TTL if not specified
	if ttl <= 0 {
		ttl = mc.config.DefaultTTL
	}

	now := time.Now()
	entry := &CacheEntry{
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		AccessCount: 0,
		LastAccess:  now,
	}

	// Check if we need to evict entries
	if len(mc.data) >= mc.config.MaxSize && mc.data[key] == nil {
		mc.evictLRU()
	}

	mc.data[key] = entry
	mc.updateAccessOrder(key)
	mc.stats.Sets++
	mc.stats.Size = len(mc.data)
	mc.stats.LastUpdated = now

	return nil
}

// Delete removes a value from cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.data[key]; exists {
		delete(mc.data, key)
		mc.removeFromAccessOrder(key)
		mc.stats.Deletes++
		mc.stats.Size = len(mc.data)
		mc.stats.LastUpdated = time.Now()
	}

	return nil
}

// Clear removes all values from cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*CacheEntry)
	mc.accessOrder = make([]string, 0)
	mc.stats.Size = 0
	mc.stats.LastUpdated = time.Now()

	return nil
}

// Exists checks if a key exists in cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return false
	}

	return !entry.IsExpired()
}

// GetTTL returns the remaining TTL for a key
func (mc *MemoryCache) GetTTL(ctx context.Context, key string) (time.Duration, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return 0, false
	}

	if entry.IsExpired() {
		return 0, false
	}

	return entry.TTL(), true
}

// GetStats returns cache statistics
func (mc *MemoryCache) GetStats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	statsCopy := mc.stats
	statsCopy.Size = len(mc.data)
	return statsCopy
}

// Close closes the cache and releases resources
func (mc *MemoryCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.cleanupRunning {
		close(mc.stopCleanup)
		mc.cleanupTicker.Stop()
		mc.cleanupRunning = false
	}

	return nil
}

// startCleanup starts the background cleanup goroutine
func (mc *MemoryCache) startCleanup() {
	mc.cleanupTicker = time.NewTicker(mc.config.CleanupInterval)
	mc.cleanupRunning = true

	go func() {
		for {
			select {
			case <-mc.cleanupTicker.C:
				mc.cleanup()
			case <-mc.stopCleanup:
				return
			}
		}
	}()
}

// cleanup removes expired entries
func (mc *MemoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range mc.data {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(mc.data, key)
		mc.removeFromAccessOrder(key)
		mc.stats.Evictions++
	}

	if len(expiredKeys) > 0 {
		mc.stats.Size = len(mc.data)
		mc.stats.LastUpdated = now
	}
}

// evictLRU removes the least recently used entry
func (mc *MemoryCache) evictLRU() {
	if len(mc.accessOrder) == 0 {
		return
	}

	// Remove the least recently used (first in access order)
	lruKey := mc.accessOrder[0]
	delete(mc.data, lruKey)
	mc.accessOrder = mc.accessOrder[1:]
	mc.stats.Evictions++
}

// updateAccessOrder updates the access order for LRU tracking
func (mc *MemoryCache) updateAccessOrder(key string) {
	// Remove key from current position
	mc.removeFromAccessOrder(key)

	// Add to end (most recently used)
	mc.accessOrder = append(mc.accessOrder, key)
}

// removeFromAccessOrder removes a key from the access order
func (mc *MemoryCache) removeFromAccessOrder(key string) {
	for i, k := range mc.accessOrder {
		if k == key {
			mc.accessOrder = append(mc.accessOrder[:i], mc.accessOrder[i+1:]...)
			break
		}
	}
}

// updateHitRatio calculates and updates the hit ratio
func (mc *MemoryCache) updateHitRatio() {
	total := mc.stats.Hits + mc.stats.Misses
	if total > 0 {
		mc.stats.HitRatio = float64(mc.stats.Hits) / float64(total)
	}
}

// Verify that MemoryCache implements the Cache interface
var _ Cache = (*MemoryCache)(nil)
