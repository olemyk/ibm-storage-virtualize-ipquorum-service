package pool

import (
	"context"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime"
	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime/mocks"
	"go.uber.org/zap"
)

// Benchmark connection pool operations

func benchFactory() (runtime.Runtime, error) {
	return mocks.NewMockRuntime("bench", runtime.RuntimeTypeDocker), nil
}

func BenchmarkPoolGet(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     5,
		MaxSize:     10,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
			pool.Put(conn)
		}
	})
}

func BenchmarkPoolGetPut(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     5,
		MaxSize:     10,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := pool.Get(context.Background())
		if err != nil {
			b.Fatal(err)
		}
		pool.Put(conn)
	}
}

func BenchmarkPoolConcurrentGetPut(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     10,
		MaxSize:     50,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
			// Simulate some work
			time.Sleep(10 * time.Microsecond)
			pool.Put(conn)
		}
	})
}

func BenchmarkPoolStats(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     5,
		MaxSize:     10,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	// Pre-populate pool
	conns := make([]*Connection, 5)
	for i := 0; i < 5; i++ {
		conn, _ := pool.Get(context.Background())
		conns[i] = conn
	}
	for _, conn := range conns {
		pool.Put(conn)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Stats()
	}
}

func BenchmarkPoolWithTimeout(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     2,
		MaxSize:     5,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			conn, err := pool.Get(ctx)
			cancel()
			if err != nil {
				continue // Timeout is acceptable in this benchmark
			}
			pool.Put(conn)
		}
	})
}

func BenchmarkPoolHighContention(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     1,
		MaxSize:     2, // Very small pool to create contention
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			conn, err := pool.Get(ctx)
			cancel()
			if err != nil {
				continue
			}
			time.Sleep(1 * time.Microsecond) // Hold connection briefly
			pool.Put(conn)
		}
	})
}

func BenchmarkPoolCreateDestroy(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     5,
		MaxSize:     10,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool, _ := NewPool(benchFactory, config, logger)
		pool.Close()
	}
}

// Benchmark different pool sizes
func BenchmarkPoolSize_Small(b *testing.B) {
	benchmarkPoolSize(b, 2, 5)
}

func BenchmarkPoolSize_Medium(b *testing.B) {
	benchmarkPoolSize(b, 10, 20)
}

func BenchmarkPoolSize_Large(b *testing.B) {
	benchmarkPoolSize(b, 50, 100)
}

func benchmarkPoolSize(b *testing.B, minSize, maxSize int) {
	logger, _ := zap.NewDevelopment()

	config := &Config{
		MinSize:     minSize,
		MaxSize:     maxSize,
		MaxIdleTime: 5 * time.Minute,
		MaxLifetime: 30 * time.Minute,
		HealthCheck: 1 * time.Minute,
	}

	pool, _ := NewPool(benchFactory, config, logger)
	defer pool.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn, err := pool.Get(context.Background())
			if err != nil {
				b.Fatal(err)
			}
			pool.Put(conn)
		}
	})
}

// Made with Bob
