package cache

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Cache provides a thread-safe in-memory cache with TTL support
type Cache struct {
	mu     sync.RWMutex
	items  map[string]*item
	logger *zap.Logger
	stopCh chan struct{}

	// Metrics
	hits    int64
	misses  int64
	evicted int64
}

// item represents a cached item
type item struct {
	value      interface{}
	expiration time.Time
	created    time.Time
}

// Config holds cache configuration
type Config struct {
	CleanupInterval time.Duration // How often to clean expired items
	DefaultTTL      time.Duration // Default TTL for items
}

// DefaultConfig returns default cache configuration
func DefaultConfig() *Config {
	return &Config{
		CleanupInterval: 1 * time.Minute,
		DefaultTTL:      5 * time.Minute,
	}
}

// New creates a new cache instance
func New(config *Config, logger *zap.Logger) *Cache {
	if config == nil {
		config = DefaultConfig()
	}

	c := &Cache{
		items:  make(map[string]*item),
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanup(config.CleanupInterval)

	return c
}

// Set stores a value in the cache with the specified TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl)
	c.items[key] = &item{
		value:      value,
		expiration: expiration,
		created:    time.Now(),
	}

	c.logger.Debug("Cache set",
		zap.String("key", key),
		zap.Duration("ttl", ttl))
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		c.misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		c.misses++
		return nil, false
	}

	c.hits++
	c.logger.Debug("Cache hit", zap.String("key", key))
	return item.value, true
}

// GetOrSet retrieves a value from cache or sets it using the provided function
func (c *Cache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Not in cache, call function
	value, err := fn(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, value, ttl)
	return value, nil
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	c.logger.Debug("Cache delete", zap.String("key", key))
}

// DeletePrefix removes all keys with the specified prefix
func (c *Cache) DeletePrefix(prefix string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
			count++
		}
	}

	c.logger.Debug("Cache delete prefix",
		zap.String("prefix", prefix),
		zap.Int("count", count))

	return count
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*item)
	c.logger.Debug("Cache cleared")
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return CacheStats{
		Items:   len(c.items),
		Hits:    c.hits,
		Misses:  c.misses,
		Evicted: c.evicted,
		HitRate: hitRate,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	Items   int
	Hits    int64
	Misses  int64
	Evicted int64
	HitRate float64
}

// Close stops the cleanup goroutine
func (c *Cache) Close() {
	close(c.stopCh)
}

// Private methods

func (c *Cache) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0

	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
			count++
			c.evicted++
		}
	}

	if count > 0 {
		c.logger.Debug("Cache cleanup",
			zap.Int("evicted", count),
			zap.Int("remaining", len(c.items)))
	}
}

// RuntimeCache provides caching for runtime operations
type RuntimeCache struct {
	cache  *Cache
	logger *zap.Logger

	// TTL configurations
	containerListTTL time.Duration
	imageListTTL     time.Duration
	runtimeInfoTTL   time.Duration
	versionTTL       time.Duration
}

// NewRuntimeCache creates a new runtime cache
func NewRuntimeCache(logger *zap.Logger) *RuntimeCache {
	return &RuntimeCache{
		cache:            New(DefaultConfig(), logger),
		logger:           logger,
		containerListTTL: 30 * time.Second,
		imageListTTL:     5 * time.Minute,
		runtimeInfoTTL:   1 * time.Minute,
		versionTTL:       10 * time.Minute,
	}
}

// GetContainerList retrieves cached container list or calls the function
func (rc *RuntimeCache) GetContainerList(ctx context.Context, runtimeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	key := "containers:" + runtimeType
	return rc.cache.GetOrSet(ctx, key, rc.containerListTTL, fn)
}

// InvalidateContainerList invalidates the container list cache
func (rc *RuntimeCache) InvalidateContainerList(runtimeType string) {
	key := "containers:" + runtimeType
	rc.cache.Delete(key)
}

// GetImageList retrieves cached image list or calls the function
func (rc *RuntimeCache) GetImageList(ctx context.Context, runtimeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	key := "images:" + runtimeType
	return rc.cache.GetOrSet(ctx, key, rc.imageListTTL, fn)
}

// InvalidateImageList invalidates the image list cache
func (rc *RuntimeCache) InvalidateImageList(runtimeType string) {
	key := "images:" + runtimeType
	rc.cache.Delete(key)
}

// GetRuntimeInfo retrieves cached runtime info or calls the function
func (rc *RuntimeCache) GetRuntimeInfo(ctx context.Context, runtimeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	key := "runtime:info:" + runtimeType
	return rc.cache.GetOrSet(ctx, key, rc.runtimeInfoTTL, fn)
}

// GetVersion retrieves cached version info or calls the function
func (rc *RuntimeCache) GetVersion(ctx context.Context, runtimeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	key := "runtime:version:" + runtimeType
	return rc.cache.GetOrSet(ctx, key, rc.versionTTL, fn)
}

// InvalidateRuntime invalidates all runtime-related caches
func (rc *RuntimeCache) InvalidateRuntime(runtimeType string) {
	rc.cache.DeletePrefix("runtime:" + runtimeType)
}

// InvalidateAll invalidates all caches
func (rc *RuntimeCache) InvalidateAll() {
	rc.cache.Clear()
}

// Stats returns cache statistics
func (rc *RuntimeCache) Stats() CacheStats {
	return rc.cache.Stats()
}

// Close closes the cache
func (rc *RuntimeCache) Close() {
	rc.cache.Close()
}

// SetTTL configures TTL for different cache types
func (rc *RuntimeCache) SetTTL(containerList, imageList, runtimeInfo, version time.Duration) {
	if containerList > 0 {
		rc.containerListTTL = containerList
	}
	if imageList > 0 {
		rc.imageListTTL = imageList
	}
	if runtimeInfo > 0 {
		rc.runtimeInfoTTL = runtimeInfo
	}
	if version > 0 {
		rc.versionTTL = version
	}
}

// Made with Bob
