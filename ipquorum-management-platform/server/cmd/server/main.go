package main

import (
	"fmt"
	"os"
	"time"

	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/api"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"github.com/olemyk/ipquorum-platform/server/pkg/config"
	"github.com/olemyk/ipquorum-platform/server/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version = "3.0.0-dev"
	commit  = "dev"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ipquorum-server",
	Short: "IP Quorum Management Platform Server",
	Long: `IP Quorum Management Platform Server
	
A modern management platform for IBM Storage Virtualize IP Quorum services.
Features:
  - Web Dashboard
  - REST API
  - Multi-server orchestration
  - Real-time monitoring
  - Health checks and metrics`,
	RunE: runServer,
}

var (
	configFile string
	debug      bool
)

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "/etc/ipquorum-platform/config.yaml", "Config file path")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ipquorum-server version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	})

	// Init command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize configuration and database",
		RunE:  initServer,
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Initialize logger
	log, err := logger.New(debug)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer log.Sync()

	log.Info("Starting IP Quorum Management Server",
		zap.String("version", version),
		zap.String("commit", commit),
	)

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Error("Failed to load configuration", zap.Error(err))
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	log.Info("Initializing database", zap.String("path", cfg.Database.Path))
	db, err := storage.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Run migrations
	log.Info("Running database migrations")
	if err := db.Migrate(); err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create default admin user
	log.Info("Creating default admin user if not exists")
	if err := db.CreateDefaultAdmin(); err != nil {
		log.Error("Failed to create default admin", zap.Error(err))
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	// Get scripts directory from config or use default
	scriptsDir := cfg.Agent.ScriptsDir
	if scriptsDir == "" {
		scriptsDir = "/opt/ipquorum/scripts"
	}

	// Initialize agent registry
	log.Info("Initializing agent registry")
	agentRegistry := agent.NewRegistry(log)

	// Load registered servers from database and register agents
	log.Info("Loading registered servers from database")
	servers, err := db.ListServers()
	if err != nil {
		log.Error("Failed to list servers", zap.Error(err))
		return fmt.Errorf("failed to list servers: %w", err)
	}

	registeredCount := 0
	for _, srv := range servers {
		// Only register servers that are marked as online
		if srv.Status == "online" {
			agentCfg := agent.Config{
				Host:       srv.IPAddress,
				Port:       srv.AgentPort,
				APIKey:     "ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7", // TODO: Store in database per server
				TLSEnabled: false,                                                              // Agent is running on HTTP
				TLSVerify:  false,
				Timeout:    30 * time.Second,
			}

			if err := agentRegistry.Register(srv.ID, agentCfg); err != nil {
				log.Warn("Failed to register agent on startup",
					zap.String("server_id", srv.ID),
					zap.String("hostname", srv.Hostname),
					zap.String("ip", srv.IPAddress),
					zap.Error(err),
				)
				// Don't fail startup, just log the warning
				// The server will be marked offline by health checks
			} else {
				registeredCount++
				log.Info("Agent registered on startup",
					zap.String("server_id", srv.ID),
					zap.String("hostname", srv.Hostname),
					zap.String("ip", srv.IPAddress),
				)
			}
		}
	}

	log.Info("Agent registry initialization complete",
		zap.Int("total_servers", len(servers)),
		zap.Int("registered_agents", registeredCount),
	)

	// Initialize API server
	log.Info("Initializing API server", zap.Int("port", cfg.Server.Port))
	server := api.NewServer(cfg, db, log, scriptsDir, agentRegistry)

	// Start server
	log.Info("Server starting", zap.String("address", fmt.Sprintf(":%d", cfg.Server.Port)))
	if err := server.Run(); err != nil {
		log.Error("Server failed", zap.Error(err))
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func initServer(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing IP Quorum Management Platform...")

	// Create default configuration
	cfg := config.Default()

	// Create config directory
	if err := os.MkdirAll("/etc/ipquorum-platform", 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration
	if err := cfg.Save("/etc/ipquorum-platform/config.yaml"); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create data directory
	if err := os.MkdirAll("/var/lib/ipquorum-platform", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize database
	db, err := storage.NewDatabase(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create default admin user
	if err := db.CreateDefaultAdmin(); err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	fmt.Println("✓ Configuration created: /etc/ipquorum-platform/config.yaml")
	fmt.Println("✓ Database initialized: /var/lib/ipquorum-platform/ipquorum.db")
	fmt.Println("✓ Default admin user created")
	fmt.Println("")
	fmt.Println("Default credentials:")
	fmt.Println("  Username: admin")
	fmt.Println("  Password: changeme")
	fmt.Println("")
	fmt.Println("⚠️  Please change the default password immediately!")
	fmt.Println("")
	fmt.Println("Start the server with:")
	fmt.Println("  sudo systemctl start ipquorum-server")
	fmt.Println("")
	fmt.Println("Access the web dashboard at:")
	fmt.Println("  https://localhost:8443")

	return nil
}
