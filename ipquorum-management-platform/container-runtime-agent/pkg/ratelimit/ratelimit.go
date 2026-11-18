package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Limiter provides rate limiting functionality
type Limiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
	logger   *zap.Logger

	// Cleanup
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// Config holds rate limiter configuration
type Config struct {
	RequestsPerSecond float64       // Requests per second
	Burst             int           // Burst size
	CleanupInterval   time.Duration // How often to clean old limiters
}

// DefaultConfig returns default rate limiter configuration
func DefaultConfig() *Config {
	return &Config{
		RequestsPerSecond: 10.0, // 10 requests per second
		Burst:             20,   // Allow burst of 20
		CleanupInterval:   5 * time.Minute,
	}
}

// New creates a new rate limiter
func New(config *Config, logger *zap.Logger) *Limiter {
	if config == nil {
		config = DefaultConfig()
	}

	return &Limiter{
		limiters:        make(map[string]*rate.Limiter),
		rate:            rate.Limit(config.RequestsPerSecond),
		burst:           config.Burst,
		logger:          logger,
		lastCleanup:     time.Now(),
		cleanupInterval: config.CleanupInterval,
	}
}

// GetLimiter returns a rate limiter for the given key
func (l *Limiter) GetLimiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[key] = limiter
	}

	// Periodic cleanup
	if time.Since(l.lastCleanup) > l.cleanupInterval {
		go l.cleanup()
		l.lastCleanup = time.Now()
	}

	return limiter
}

// Allow checks if a request is allowed for the given key
func (l *Limiter) Allow(key string) bool {
	limiter := l.GetLimiter(key)
	return limiter.Allow()
}

// Wait waits until a request is allowed for the given key
func (l *Limiter) Wait(key string) error {
	limiter := l.GetLimiter(key)
	return limiter.Wait(nil)
}

// cleanup removes inactive limiters
func (l *Limiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Remove limiters that haven't been used recently
	// This is a simple implementation - in production, you might want
	// to track last access time for each limiter
	if len(l.limiters) > 1000 {
		l.limiters = make(map[string]*rate.Limiter)
		l.logger.Debug("Rate limiter cleanup performed")
	}
}

// Middleware returns a Gin middleware for rate limiting
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as key
		key := c.ClientIP()

		limiter := l.GetLimiter(key)

		if !limiter.Allow() {
			// Rate limit exceeded
			l.logger.Warn("Rate limit exceeded",
				zap.String("ip", key),
				zap.String("path", c.Request.URL.Path))

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(l.burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMITED",
				"details": gin.H{
					"limit":  l.burst,
					"window": "1s",
				},
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(l.burst))
		// Note: Getting exact remaining count from rate.Limiter is not straightforward
		// This is an approximation
		c.Header("X-RateLimit-Remaining", strconv.Itoa(l.burst-1))

		c.Next()
	}
}

// PerEndpointLimiter provides different rate limits for different endpoints
type PerEndpointLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*Limiter
	logger   *zap.Logger

	// Default limiter
	defaultLimiter *Limiter
}

// EndpointConfig holds configuration for a specific endpoint
type EndpointConfig struct {
	Path              string
	RequestsPerSecond float64
	Burst             int
}

// NewPerEndpointLimiter creates a new per-endpoint rate limiter
func NewPerEndpointLimiter(defaultConfig *Config, endpoints []EndpointConfig, logger *zap.Logger) *PerEndpointLimiter {
	if defaultConfig == nil {
		defaultConfig = DefaultConfig()
	}

	pel := &PerEndpointLimiter{
		limiters:       make(map[string]*Limiter),
		logger:         logger,
		defaultLimiter: New(defaultConfig, logger),
	}

	// Create limiters for specific endpoints
	for _, ep := range endpoints {
		config := &Config{
			RequestsPerSecond: ep.RequestsPerSecond,
			Burst:             ep.Burst,
			CleanupInterval:   5 * time.Minute,
		}
		pel.limiters[ep.Path] = New(config, logger)
	}

	return pel
}

// Middleware returns a Gin middleware for per-endpoint rate limiting
func (pel *PerEndpointLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Get limiter for this endpoint
		pel.mu.RLock()
		limiter, exists := pel.limiters[path]
		pel.mu.RUnlock()

		if !exists {
			limiter = pel.defaultLimiter
		}

		// Use client IP as key
		key := c.ClientIP() + ":" + path

		if !limiter.Allow(key) {
			// Rate limit exceeded
			pel.logger.Warn("Rate limit exceeded",
				zap.String("ip", c.ClientIP()),
				zap.String("path", path))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMITED",
				"details": gin.H{
					"endpoint": path,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	logger     *zap.Logger
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(capacity float64, refillRate float64, logger *zap.Logger) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
		logger:     logger,
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow(tokens float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

// Tokens returns the current number of tokens
func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	return tb.tokens
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// SlidingWindowLimiter implements a sliding window rate limiter
type SlidingWindowLimiter struct {
	mu      sync.RWMutex
	windows map[string]*window
	limit   int
	window  time.Duration
	logger  *zap.Logger
}

type window struct {
	requests []time.Time
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(limit int, windowDuration time.Duration, logger *zap.Logger) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string]*window),
		limit:   limit,
		window:  windowDuration,
		logger:  logger,
	}
}

// Allow checks if a request is allowed for the given key
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swl.window)

	// Get or create window
	w, exists := swl.windows[key]
	if !exists {
		w = &window{requests: make([]time.Time, 0)}
		swl.windows[key] = w
	}

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, t := range w.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	w.requests = validRequests

	// Check if we're under the limit
	if len(w.requests) < swl.limit {
		w.requests = append(w.requests, now)
		return true
	}

	return false
}

// Middleware returns a Gin middleware for sliding window rate limiting
func (swl *SlidingWindowLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if !swl.Allow(key) {
			swl.logger.Warn("Rate limit exceeded (sliding window)",
				zap.String("ip", key),
				zap.String("path", c.Request.URL.Path))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMITED",
				"details": gin.H{
					"limit":  swl.limit,
					"window": swl.window.String(),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Cleanup removes old windows
func (swl *SlidingWindowLimiter) Cleanup() {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swl.window * 2)

	for key, w := range swl.windows {
		if len(w.requests) == 0 {
			delete(swl.windows, key)
			continue
		}

		// Check if all requests are old
		allOld := true
		for _, t := range w.requests {
			if t.After(cutoff) {
				allOld = false
				break
			}
		}

		if allOld {
			delete(swl.windows, key)
		}
	}

	swl.logger.Debug("Sliding window cleanup",
		zap.Int("remaining_windows", len(swl.windows)))
}

// Made with Bob
