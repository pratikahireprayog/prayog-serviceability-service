package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, bool)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all values from cache
	Clear(ctx context.Context) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) bool

	// GetTTL returns the remaining TTL for a key
	GetTTL(ctx context.Context, key string) (time.Duration, bool)

	// GetStats returns cache statistics
	GetStats() CacheStats

	// Close closes the cache and releases resources
	Close() error
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Sets        int64     `json:"sets"`
	Deletes     int64     `json:"deletes"`
	Evictions   int64     `json:"evictions"`
	Size        int       `json:"size"`
	MaxSize     int       `json:"max_size"`
	HitRatio    float64   `json:"hit_ratio"`
	LastUpdated time.Time `json:"last_updated"`
}

// CacheConfig holds configuration for cache implementations
type CacheConfig struct {
	MaxSize         int           `yaml:"max_size" json:"max_size"`
	DefaultTTL      time.Duration `yaml:"default_ttl" json:"default_ttl"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	EnableMetrics   bool          `yaml:"enable_metrics" json:"enable_metrics"`
}

// CacheEntry represents a cached entry with metadata
type CacheEntry struct {
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
	LastAccess  time.Time   `json:"last_access"`
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// TTL returns the remaining time to live for the entry
func (e *CacheEntry) TTL() time.Duration {
	if e.ExpiresAt.IsZero() {
		return 0 // No expiration
	}
	remaining := time.Until(e.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}
