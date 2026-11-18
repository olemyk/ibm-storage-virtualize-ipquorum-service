package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/internal/api"
	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/config"
	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime"
	"go.uber.org/zap"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("IPQuorum Container Runtime Agent\n")
		fmt.Printf("Version:    %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting IPQuorum Container Runtime Agent",
		zap.String("version", version),
		zap.String("build_time", buildTime),
		zap.String("git_commit", gitCommit),
	)

	// Initialize runtime
	ctx := context.Background()
	rt, err := initRuntime(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize container runtime", zap.Error(err))
	}

	logger.Info("Container runtime initialized",
		zap.String("runtime", rt.Name()),
		zap.String("type", string(rt.Type())),
	)

	// Initialize API server
	server := api.NewServer(cfg, logger, rt)

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Starting API server",
			zap.String("address", cfg.GetServerAddress()),
		)
		serverErrors <- server.Start()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal("Server error", zap.Error(err))
	case sig := <-shutdown:
		logger.Info("Shutdown signal received",
			zap.String("signal", sig.String()),
		)

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
			os.Exit(1)
		}

		logger.Info("Server stopped gracefully")
	}
}

// loadConfig loads configuration from file or environment
func loadConfig(configPath string) (*config.Config, error) {
	if configPath != "" {
		return config.Load(configPath)
	}
	return config.LoadFromEnv()
}

// initLogger initializes the logger based on configuration
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsProductionMode() {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	switch cfg.Logging.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set encoding (json or console)
	if cfg.Logging.Format == "json" {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}

	// Set output paths
	if cfg.Logging.Output == "file" && cfg.Logging.File != "" {
		zapConfig.OutputPaths = []string{cfg.Logging.File}
		zapConfig.ErrorOutputPaths = []string{cfg.Logging.File}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	return zapConfig.Build()
}

// initRuntime initializes the container runtime
func initRuntime(ctx context.Context, cfg *config.Config, logger *zap.Logger) (runtime.Runtime, error) {
	// Create runtime factory
	factoryConfig := &runtime.FactoryConfig{
		PreferredRuntime: cfg.Runtime.PreferredRuntime,
		DockerSocketPath: cfg.Runtime.DockerSocketPath,
		PodmanSocketPath: cfg.Runtime.PodmanSocketPath,
		DetectionTimeout: cfg.Runtime.DetectionTimeout,
	}

	factory := runtime.NewFactory(factoryConfig)

	// Detect available runtimes
	available, err := factory.DetectAvailableRuntimes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect runtimes: %w", err)
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no container runtime available (Docker or Podman required)")
	}

	logger.Info("Available container runtimes",
		zap.Int("count", len(available)),
		zap.Any("runtimes", available),
	)

	// Create runtime instance
	rt, err := factory.CreateRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	// Validate runtime
	if err := factory.ValidateRuntime(ctx, rt.Type()); err != nil {
		return nil, fmt.Errorf("runtime validation failed: %w", err)
	}

	// Get runtime info
	info, err := rt.Info(ctx)
	if err != nil {
		logger.Warn("Failed to get runtime info", zap.Error(err))
	} else {
		logger.Info("Runtime information",
			zap.String("name", info.Name),
			zap.String("version", info.Version),
			zap.String("api_version", info.APIVersion),
			zap.Int("containers_running", info.ContainersRunning),
			zap.Int("containers_stopped", info.ContainersStopped),
		)
	}

	return rt, nil
}

// Made with Bob
