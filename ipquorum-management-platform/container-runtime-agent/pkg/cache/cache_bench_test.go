package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Benchmark cache operations

func BenchmarkCacheSet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key-%d", i%1000))
	}
}

func BenchmarkCacheGetMiss(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("missing-key-%d", i))
	}
}

func BenchmarkCacheGetOrSet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	counter := 0
	loader := func(ctx context.Context) (interface{}, error) {
		counter++
		return fmt.Sprintf("value-%d", counter), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetOrSet(context.Background(), fmt.Sprintf("key-%d", i%100), 5*time.Minute, loader)
	}
}

func BenchmarkCacheDelete(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(fmt.Sprintf("key-%d", i))
	}
}

func BenchmarkCacheDeletePrefix(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pre-populate for each iteration
		for j := 0; j < 100; j++ {
			cache.Set(fmt.Sprintf("prefix-%d-%d", i, j), "value", 5*time.Minute)
		}
		cache.DeletePrefix(fmt.Sprintf("prefix-%d", i))
	}
}

func BenchmarkCacheClear(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pre-populate for each iteration
		for j := 0; j < 1000; j++ {
			cache.Set(fmt.Sprintf("key-%d", j), "value", 5*time.Minute)
		}
		cache.Clear()
	}
}

func BenchmarkCacheStats(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
	}

	// Generate some hits and misses
	for i := 0; i < 100; i++ {
		cache.Get(fmt.Sprintf("key-%d", i))
		cache.Get(fmt.Sprintf("missing-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Stats()
	}
}

func BenchmarkCacheConcurrentSet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
			i++
		}
	})
}

func BenchmarkCacheConcurrentGet(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(fmt.Sprintf("key-%d", i%1000))
			i++
		}
	})
}

func BenchmarkCacheConcurrentMixed(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				cache.Get(fmt.Sprintf("key-%d", i%1000))
			case 1:
				cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
			case 2:
				cache.Delete(fmt.Sprintf("key-%d", i%1000))
			case 3:
				_ = cache.Stats()
			}
			i++
		}
	})
}

func BenchmarkCacheWithTTL(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	ttls := []time.Duration{
		1 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		1 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ttl := ttls[i%len(ttls)]
		cache.Set(fmt.Sprintf("key-%d", i), "value", ttl)
	}
}

func BenchmarkCacheExpiration(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Set items with very short TTL
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), "value", 1*time.Millisecond)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key-%d", i%1000))
	}
}

// Benchmark different cache sizes
func BenchmarkCacheSize_Small(b *testing.B) {
	benchmarkCacheSize(b, 100)
}

func BenchmarkCacheSize_Medium(b *testing.B) {
	benchmarkCacheSize(b, 1000)
}

func BenchmarkCacheSize_Large(b *testing.B) {
	benchmarkCacheSize(b, 10000)
}

func benchmarkCacheSize(b *testing.B, size int) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate cache
	for i := 0; i < size; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(fmt.Sprintf("key-%d", i%size))
			i++
		}
	})
}

// Benchmark memory usage
func BenchmarkCacheMemory(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	b.Run("SmallValues", func(b *testing.B) {
		cache := New(nil, logger)
		defer cache.Close()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key-%d", i), "small", 5*time.Minute)
		}
	})

	b.Run("LargeValues", func(b *testing.B) {
		cache := New(nil, logger)
		defer cache.Close()
		largeValue := make([]byte, 1024*1024) // 1MB
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key-%d", i), largeValue, 5*time.Minute)
		}
	})
}

// Benchmark cleanup performance
func BenchmarkCacheCleanup(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add expired items
		for j := 0; j < 1000; j++ {
			cache.Set(fmt.Sprintf("key-%d-%d", i, j), "value", 1*time.Nanosecond)
		}

		// Trigger cleanup by accessing cache
		time.Sleep(1 * time.Millisecond)
		cache.Get("trigger-cleanup")
	}
}

// Benchmark concurrent access patterns
func BenchmarkCacheReadHeavy(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 90% reads, 10% writes
			if i%10 == 0 {
				cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
			} else {
				cache.Get(fmt.Sprintf("key-%d", i%1000))
			}
			i++
		}
	})
}

func BenchmarkCacheWriteHeavy(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 90% writes, 10% reads
			if i%10 == 0 {
				cache.Get(fmt.Sprintf("key-%d", i%1000))
			} else {
				cache.Set(fmt.Sprintf("key-%d", i), "value", 5*time.Minute)
			}
			i++
		}
	})
}

// Benchmark GetOrSet with concurrent access
func BenchmarkCacheGetOrSetConcurrent(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cache := New(nil, logger)
	defer cache.Close()

	var counter int
	var mu sync.Mutex

	loader := func(ctx context.Context) (interface{}, error) {
		mu.Lock()
		counter++
		val := counter
		mu.Unlock()
		time.Sleep(1 * time.Millisecond) // Simulate slow load
		return fmt.Sprintf("value-%d", val), nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Multiple goroutines trying to load the same key
			cache.GetOrSet(context.Background(), fmt.Sprintf("key-%d", i%10), 5*time.Minute, loader)
			i++
		}
	})
}

// Made with Bob
