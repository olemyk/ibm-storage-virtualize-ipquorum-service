package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/olemyk/ipquorum-platform/agent/internal/api"
	"github.com/olemyk/ipquorum-platform/agent/internal/config"
	"github.com/olemyk/ipquorum-platform/agent/internal/executor"
	"github.com/olemyk/ipquorum-platform/agent/pkg/logger"
	"go.uber.org/zap"
)

var (
	version = "1.0.0"
	commit  = "dev"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "/etc/ipquorum-agent/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("IPQuorum Agent v%s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Agent.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting IPQuorum Agent",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("instance", cfg.Agent.InstanceName))

	// Validate configuration
	if cfg.Agent.APIKey == "" {
		log.Fatal("API key is required in configuration")
	}

	if cfg.Script.Path == "" {
		log.Fatal("Script path is required in configuration")
	}

	// Check if script exists
	if _, err := os.Stat(cfg.Script.Path); os.IsNotExist(err) {
		log.Fatal("Script not found", zap.String("path", cfg.Script.Path))
	}

	// Initialize executor
	exec := executor.NewExecutor(cfg.Script.Path, cfg.Script.Timeout, log)

	// Initialize API server
	server := api.NewServer(cfg, exec, log)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := server.Run(); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("Received shutdown signal", zap.String("signal", sig.String()))
	log.Info("Agent stopped")
}
