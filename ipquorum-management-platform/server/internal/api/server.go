package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/auth"
	"github.com/olemyk/ipquorum-platform/server/internal/executor"
	"github.com/olemyk/ipquorum-platform/server/internal/health"
	"github.com/olemyk/ipquorum-platform/server/internal/service"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"github.com/olemyk/ipquorum-platform/server/pkg/config"
	"github.com/olemyk/ipquorum-platform/server/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Server represents the API server
type Server struct {
	config          *config.Config
	db              *storage.Database
	log             *zap.Logger
	jwtManager      *auth.JWTManager
	router          *gin.Engine
	executor        *executor.Executor
	instanceService *service.InstanceService
	healthChecker   *health.HealthChecker
	metrics         *metrics.Metrics
	agentRegistry   *agent.Registry
	startTime       time.Time
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, db *storage.Database, log *zap.Logger, scriptsDir string, agentRegistry *agent.Registry) *Server {
	// Set Gin mode
	if cfg.Server.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.Auth.JWTSecret,
		cfg.Auth.TokenExpiry,
		cfg.Auth.RefreshExpiry,
	)

	// Create executor for bash scripts (with agent registry for remote execution)
	exec := executor.NewExecutor(log, scriptsDir, 5*time.Minute, agentRegistry)

	// Create instance service (with agent registry for remote execution)
	instanceService := service.NewInstanceService(db, exec, agentRegistry, log)

	// Create health checker (with agent registry for remote health checks)
	healthCheckInterval := 30 * time.Second
	if cfg.Server.HealthCheckInterval > 0 {
		healthCheckInterval = time.Duration(cfg.Server.HealthCheckInterval) * time.Second
	}
	healthChecker := health.NewHealthChecker(db, exec, agentRegistry, log, healthCheckInterval)

	// Create metrics
	metricsCollector := metrics.New("ipquorum")

	server := &Server{
		config:          cfg,
		db:              db,
		log:             log,
		jwtManager:      jwtManager,
		router:          gin.New(),
		executor:        exec,
		instanceService: instanceService,
		healthChecker:   healthChecker,
		metrics:         metricsCollector,
		agentRegistry:   agentRegistry,
		startTime:       time.Now(),
	}

	server.setupMiddleware()
	server.setupRoutes()

	// Start health checker
	healthChecker.Start()

	return server
}

// setupMiddleware sets up global middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Metrics middleware (before logger to capture all requests)
	s.router.Use(s.metrics.Middleware())

	// Logger middleware
	s.router.Use(func(c *gin.Context) {
		c.Next()
		s.log.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
		)
	})

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Health check (no auth required)
	s.router.GET("/health", s.handleHealth)

	// Metrics endpoint (no auth required)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// System health check (no auth required)
		v1.GET("/health", s.handleHealth)

		// Authentication routes (no auth required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/login", s.handleLogin)
			authRoutes.POST("/logout", s.handleLogout)
			authRoutes.POST("/refresh", s.handleRefresh)
			authRoutes.GET("/me", auth.AuthMiddleware(s.jwtManager), s.handleMe)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(auth.AuthMiddleware(s.jwtManager))
		{
			// User management routes
			users := protected.Group("/users")
			users.Use(auth.RequireRole("admin"))
			{
				users.GET("", s.handleListUsers)
				users.POST("", s.handleCreateUser)
				users.GET("/:id", s.handleGetUser)
				users.PUT("/:id", s.handleUpdateUser)
				users.DELETE("/:id", s.handleDeleteUser)
			}

			// Instance routes
			instances := protected.Group("/instances")
			{
				instances.GET("", s.handleListInstances)
				instances.POST("", auth.RequireRole("admin", "operator"), s.handleCreateInstance)
				instances.GET("/:id", s.handleGetInstance)
				instances.PUT("/:id", auth.RequireRole("admin", "operator"), s.handleUpdateInstance)
				instances.DELETE("/:id", auth.RequireRole("admin", "operator"), s.handleDeleteInstance)
				instances.POST("/:id/start", auth.RequireRole("admin", "operator"), s.handleStartInstance)
				instances.POST("/:id/stop", auth.RequireRole("admin", "operator"), s.handleStopInstance)
				instances.POST("/:id/restart", auth.RequireRole("admin", "operator"), s.handleRestartInstance)
				instances.GET("/:id/status", s.handleGetInstanceStatus)
				instances.GET("/:id/logs", s.handleGetInstanceLogs)
			}

			// Health monitoring routes
			health := protected.Group("/health")
			{
				health.GET("/instances", s.handleGetAllInstancesHealth)
				health.GET("/instances/:id", s.handleGetInstanceHealth)
			}

			// Metrics routes
			metrics := protected.Group("/metrics")
			{
				metrics.GET("/instances/:id", s.handleGetInstanceMetrics)
			}

			// Server routes (for multi-server support)
			servers := protected.Group("/servers")
			servers.Use(auth.RequireRole("admin"))
			{
				servers.GET("", s.handleListServers)
				servers.POST("", s.handleRegisterServer)
				servers.GET("/:id", s.handleGetServer)
				servers.DELETE("/:id", s.handleUnregisterServer)
				servers.GET("/:id/instances", s.handleGetServerInstances)
			}

			// Configuration routes
			configRoutes := protected.Group("/config")
			configRoutes.Use(auth.RequireRole("admin"))
			{
				configRoutes.GET("", s.handleGetConfig)
				configRoutes.PUT("", s.handleUpdateConfig)
			}
		}
	}

	// Serve web dashboard (if built)
	s.router.Static("/dashboard", "./web/dist")
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})
}

// Run starts the API server
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	// Check if TLS is configured
	if s.config.Server.TLSCert != "" && s.config.Server.TLSKey != "" {
		s.log.Info("Starting HTTPS server", zap.String("address", addr))
		return s.router.RunTLS(addr, s.config.Server.TLSCert, s.config.Server.TLSKey)
	}

	s.log.Warn("Starting HTTP server (TLS not configured)", zap.String("address", addr))
	return s.router.Run(addr)
}

// Stop stops the API server and health checker
func (s *Server) Stop() {
	s.log.Info("Stopping server...")
	if s.healthChecker != nil {
		s.healthChecker.Stop()
	}
}

// Health check handler
func (s *Server) handleHealth(c *gin.Context) {
	// Check database connection
	dbStatus := "healthy"
	if err := s.db.Ping(); err != nil {
		dbStatus = "unhealthy"
	}

	// Get instance counts
	instances, err := s.db.ListInstances()
	if err != nil {
		s.log.Error("Failed to list instances for health check", zap.Error(err))
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, instance := range instances {
		switch instance.Health {
		case "healthy":
			healthyCount++
		case "degraded":
			degradedCount++
		case "unhealthy":
			unhealthyCount++
		}
	}

	// Calculate uptime in seconds
	uptime := int64(time.Since(s.startTime).Seconds())

	response := storage.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime,
		Components: map[string]interface{}{
			"database": dbStatus,
			"instances": map[string]interface{}{
				"total":     len(instances),
				"healthy":   healthyCount,
				"degraded":  degradedCount,
				"unhealthy": unhealthyCount,
			},
		},
	}

	if dbStatus != "healthy" {
		response.Status = "unhealthy"
	}

	c.JSON(http.StatusOK, response)
}

// Login handler
func (s *Server) handleLogin(c *gin.Context) {
	var req storage.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get user from database
	user, err := s.db.GetUserByUsername(req.Username)
	if err != nil {
		s.log.Warn("Login attempt for non-existent user", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.log.Warn("Failed login attempt", zap.String("username", req.Username))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		s.log.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		s.log.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Update last login
	if err := s.db.UpdateUserLastLogin(user.ID); err != nil {
		s.log.Error("Failed to update last login", zap.Error(err))
	}

	s.log.Info("User logged in", zap.String("username", user.Username))

	response := storage.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtManager.GetTokenExpiry(),
		User:         user,
	}

	c.JSON(http.StatusOK, response)
}

// Logout handler
func (s *Server) handleLogout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// by removing the token. We just log the event.
	username := auth.GetUsername(c)
	s.log.Info("User logged out", zap.String("username", username))
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Refresh token handler
func (s *Server) handleRefresh(c *gin.Context) {
	// TODO: Implement refresh token logic
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// Get current user handler
func (s *Server) handleMe(c *gin.Context) {
	username := auth.GetUsername(c)
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// List instances handler
func (s *Server) handleListInstances(c *gin.Context) {
	instances, err := s.db.ListInstances()
	if err != nil {
		s.log.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list instances"})
		return
	}

	c.JSON(http.StatusOK, instances)
}

// Create instance handler
func (s *Server) handleCreateInstance(c *gin.Context) {
	var req storage.CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Convert pointer booleans to regular booleans with defaults
	enableDownload := true // Default to true
	if req.EnableDownload != nil {
		enableDownload = *req.EnableDownload
	}
	enableMkquorumapp := false // Default to false
	if req.EnableMkquorumapp != nil {
		enableMkquorumapp = *req.EnableMkquorumapp
	}
	ip6 := false
	if req.IP6 != nil {
		ip6 = *req.IP6
	}
	partnerIP6 := false
	if req.PartnerIP6 != nil {
		partnerIP6 = *req.PartnerIP6
	}
	noMetadata := false
	if req.NoMetadata != nil {
		noMetadata = *req.NoMetadata
	}

	// Default to "local" if server_id is not provided or empty
	serverID := req.ServerID
	if serverID == "" {
		serverID = "local"
	}

	instance := &storage.Instance{
		Name:              req.Name,
		ServerID:          serverID,
		APIEndpoint:       req.APIEndpoint,
		Username:          req.Username,
		Password:          req.Password,
		StorageSystem:     req.StorageSystem,
		Description:       req.Description,
		Location:          req.Location,
		EnableDownload:    enableDownload,
		EnableMkquorumapp: enableMkquorumapp,
		Partnersystem:     req.Partnersystem,
		IPQuorumName:      req.IPQuorumName,
		IP6:               ip6,
		PartnerIP6:        partnerIP6,
		NoMetadata:        noMetadata,
	}

	// Use instance service to create instance (calls bash script)
	if err := s.instanceService.CreateInstance(c.Request.Context(), instance); err != nil {
		s.log.Error("Failed to create instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "create", "instance", instance.ID, fmt.Sprintf("Created instance: %s", instance.Name))

	s.log.Info("Instance created", zap.String("name", instance.Name), zap.String("id", instance.ID))
	c.JSON(http.StatusCreated, instance)
}

// Get instance handler
func (s *Server) handleGetInstance(c *gin.Context) {
	id := c.Param("id")
	instance, err := s.db.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// Update instance handler
func (s *Server) handleUpdateInstance(c *gin.Context) {
	id := c.Param("id")

	// Get existing instance
	instance, err := s.db.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	var req storage.UpdateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update fields if provided
	if req.APIEndpoint != "" {
		instance.APIEndpoint = req.APIEndpoint
	}
	if req.Username != "" {
		instance.Username = req.Username
	}
	if req.Password != "" {
		instance.Password = req.Password
	}
	if req.StorageSystem != "" {
		instance.StorageSystem = req.StorageSystem
	}
	if req.Description != "" {
		instance.Description = req.Description
	}
	if req.Location != "" {
		instance.Location = req.Location
	}
	if req.Partnersystem != "" {
		instance.Partnersystem = req.Partnersystem
	}
	if req.IPQuorumName != "" {
		instance.IPQuorumName = req.IPQuorumName
	}

	// Update boolean fields if provided
	if req.EnableDownload != nil {
		instance.EnableDownload = *req.EnableDownload
	}
	if req.EnableMkquorumapp != nil {
		instance.EnableMkquorumapp = *req.EnableMkquorumapp
	}
	if req.IP6 != nil {
		instance.IP6 = *req.IP6
	}
	if req.PartnerIP6 != nil {
		instance.PartnerIP6 = *req.PartnerIP6
	}
	if req.NoMetadata != nil {
		instance.NoMetadata = *req.NoMetadata
	}

	// Update instance in database
	if err := s.db.UpdateInstance(instance); err != nil {
		s.log.Error("Failed to update instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "update", "instance", id, fmt.Sprintf("Updated instance: %s", instance.Name))

	s.log.Info("Instance updated", zap.String("name", instance.Name), zap.String("id", id))
	c.JSON(http.StatusOK, instance)
}

// Delete instance handler
func (s *Server) handleDeleteInstance(c *gin.Context) {
	id := c.Param("id")

	// Get instance for logging
	instance, err := s.db.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	// Check for force parameter
	force := c.Query("force") == "true"

	// Use instance service to delete instance (calls bash script)
	if err := s.instanceService.DeleteInstance(c.Request.Context(), id, force); err != nil {
		s.log.Error("Failed to delete instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "delete", "instance", id, fmt.Sprintf("Deleted instance: %s", instance.Name))

	s.log.Info("Instance deleted", zap.String("name", instance.Name), zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Instance deleted successfully"})
}

// Start instance handler
func (s *Server) handleStartInstance(c *gin.Context) {
	id := c.Param("id")

	if err := s.instanceService.StartInstance(c.Request.Context(), id); err != nil {
		s.log.Error("Failed to start instance", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "start", "instance", id, "Started instance")

	c.JSON(http.StatusOK, gin.H{"message": "Instance started successfully"})
}

// Stop instance handler
func (s *Server) handleStopInstance(c *gin.Context) {
	id := c.Param("id")

	if err := s.instanceService.StopInstance(c.Request.Context(), id); err != nil {
		s.log.Error("Failed to stop instance", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "stop", "instance", id, "Stopped instance")

	c.JSON(http.StatusOK, gin.H{"message": "Instance stopped successfully"})
}

// Restart instance handler
func (s *Server) handleRestartInstance(c *gin.Context) {
	id := c.Param("id")

	if err := s.instanceService.RestartInstance(c.Request.Context(), id); err != nil {
		s.log.Error("Failed to restart instance", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "restart", "instance", id, "Restarted instance")

	c.JSON(http.StatusOK, gin.H{"message": "Instance restarted successfully"})
}

// Get instance status handler
func (s *Server) handleGetInstanceStatus(c *gin.Context) {
	id := c.Param("id")

	instance, err := s.instanceService.GetInstanceStatus(c.Request.Context(), id)
	if err != nil {
		s.log.Error("Failed to get instance status", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      instance.ID,
		"name":    instance.Name,
		"status":  instance.Status,
		"health":  instance.Health,
		"updated": instance.UpdatedAt,
	})
}

// Get instance logs handler
func (s *Server) handleGetInstanceLogs(c *gin.Context) {
	id := c.Param("id")

	// Get lines parameter (default: 50)
	linesStr := c.DefaultQuery("lines", "50")
	lines, err := strconv.Atoi(linesStr)
	if err != nil || lines < 1 || lines > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lines parameter (must be 1-10000)"})
		return
	}

	// Get instance from database
	instance, err := s.db.GetInstance(id)
	if err != nil {
		s.log.Error("Failed to get instance", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	// Get logs based on server_id
	var result *executor.ExecuteResult
	if instance.ServerID == "local" {
		// Local execution
		result = s.executor.GetLogs(c.Request.Context(), instance.Name, lines)
	} else {
		// Remote execution via agent
		result = s.executor.GetLogsRemote(c.Request.Context(), instance.ServerID, instance.Name, lines)
	}

	if result.Error != nil {
		s.log.Error("Failed to get logs",
			zap.String("instance", instance.Name),
			zap.String("server_id", instance.ServerID),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get logs",
			"details": result.Stderr,
		})
		return
	}

	// Return logs as plain text
	c.String(http.StatusOK, result.Stdout)
}

// Get all instances health handler
func (s *Server) handleGetAllInstancesHealth(c *gin.Context) {
	instances, err := s.db.ListInstances()
	if err != nil {
		s.log.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list instances"})
		return
	}

	healthSummary := make([]gin.H, 0, len(instances))
	for _, instance := range instances {
		healthSummary = append(healthSummary, gin.H{
			"id":      instance.ID,
			"name":    instance.Name,
			"status":  instance.Status,
			"health":  instance.Health,
			"updated": instance.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": healthSummary,
		"timestamp": time.Now(),
	})
}

// Get instance health handler
func (s *Server) handleGetInstanceHealth(c *gin.Context) {
	id := c.Param("id")

	// Get limit parameter (default 10)
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Check if immediate check is requested
	if c.Query("check") == "now" {
		healthCheck, err := s.healthChecker.CheckInstanceNow(id)
		if err != nil {
			s.log.Error("Failed to check instance health", zap.String("id", id), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, healthCheck)
		return
	}

	// Get health history
	history, err := s.healthChecker.GetInstanceHealthHistory(id, limit)
	if err != nil {
		s.log.Error("Failed to get health history", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get current instance status
	instance, err := s.db.GetInstance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance": gin.H{
			"id":      instance.ID,
			"name":    instance.Name,
			"status":  instance.Status,
			"health":  instance.Health,
			"updated": instance.UpdatedAt,
		},
		"history": history,
	})
}

// Get instance metrics handler
func (s *Server) handleGetInstanceMetrics(c *gin.Context) {
	// TODO: Implement get instance metrics logic
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// List servers handler
func (s *Server) handleListServers(c *gin.Context) {
	servers, err := s.db.ListServers()
	if err != nil {
		s.log.Error("Failed to list servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// Register server handler
func (s *Server) handleRegisterServer(c *gin.Context) {
	var req struct {
		Hostname   string `json:"hostname" binding:"required"`
		IPAddress  string `json:"ip_address" binding:"required"`
		AgentPort  int    `json:"agent_port"`
		APIKey     string `json:"api_key" binding:"required"`
		TLSEnabled bool   `json:"tls_enabled"`
		TLSVerify  bool   `json:"tls_verify"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AgentPort == 0 {
		req.AgentPort = 9090
	}

	// Create server record
	server := &storage.Server{
		Hostname:  req.Hostname,
		IPAddress: req.IPAddress,
		AgentPort: req.AgentPort,
		Status:    "online",
	}

	if err := s.db.CreateServer(server); err != nil {
		s.log.Error("Failed to create server", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Register in agent registry
	cfg := agent.Config{
		Host:       req.IPAddress,
		Port:       req.AgentPort,
		APIKey:     req.APIKey,
		TLSEnabled: req.TLSEnabled,
		TLSVerify:  req.TLSVerify,
		Timeout:    30 * time.Second,
	}

	if err := s.agentRegistry.Register(server.ID, cfg); err != nil {
		// Rollback database entry
		s.db.DeleteServer(server.ID)
		s.log.Error("Failed to connect to agent",
			zap.String("server_id", server.ID),
			zap.String("host", req.IPAddress),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to agent: " + err.Error()})
		return
	}

	s.log.Info("Server registered successfully",
		zap.String("server_id", server.ID),
		zap.String("hostname", server.Hostname),
		zap.String("ip", server.IPAddress),
	)

	c.JSON(http.StatusCreated, server)
}

// Get server handler
func (s *Server) handleGetServer(c *gin.Context) {
	id := c.Param("id")

	server, err := s.db.GetServer(id)
	if err != nil {
		s.log.Error("Failed to get server", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// Unregister server handler
func (s *Server) handleUnregisterServer(c *gin.Context) {
	id := c.Param("id")

	// Unregister from agent registry
	s.agentRegistry.Unregister(id)

	// Delete from database
	if err := s.db.DeleteServer(id); err != nil {
		s.log.Error("Failed to delete server", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.log.Info("Server deleted", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}

// Get server instances handler
func (s *Server) handleGetServerInstances(c *gin.Context) {
	id := c.Param("id")

	// Get server info
	server, err := s.db.GetServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Get agent client
	client, err := s.agentRegistry.Get(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"server":    server,
			"status":    "offline",
			"instances": []interface{}{},
		})
		return
	}

	// Check health
	ctx := c.Request.Context()
	healthErr := client.Health(ctx)

	status := "online"
	if healthErr != nil {
		status = "offline"
	}

	// Try to get instances list
	instances := []interface{}{}
	if healthErr == nil {
		if instanceList, err := client.ListInstances(ctx); err == nil {
			instances = make([]interface{}, len(instanceList))
			for i, inst := range instanceList {
				instances[i] = inst
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"server":    server,
		"status":    status,
		"instances": instances,
	})
}

// Get config handler
func (s *Server) handleGetConfig(c *gin.Context) {
	// TODO: Implement get config logic
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// Update config handler
func (s *Server) handleUpdateConfig(c *gin.Context) {
	// TODO: Implement update config logic
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// User management handlers

// List users handler
func (s *Server) handleListUsers(c *gin.Context) {
	users, err := s.db.ListUsers()
	if err != nil {
		s.log.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Create user handler
func (s *Server) handleCreateUser(c *gin.Context) {
	var req storage.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user := &storage.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		Active:       true,
	}

	if err := s.db.CreateUser(user); err != nil {
		s.log.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "create", "user", user.ID, fmt.Sprintf("Created user: %s", user.Username))

	s.log.Info("User created", zap.String("username", user.Username), zap.String("id", user.ID))
	c.JSON(http.StatusCreated, user)
}

// Get user handler
func (s *Server) handleGetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := s.db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update user handler
func (s *Server) handleUpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req storage.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get existing user
	user, err := s.db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.log.Error("Failed to hash password", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := s.db.UpdateUser(user); err != nil {
		s.log.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "update", "user", id, fmt.Sprintf("Updated user: %s", user.Username))

	s.log.Info("User updated", zap.String("username", user.Username), zap.String("id", id))
	c.JSON(http.StatusOK, user)
}

// Delete user handler
func (s *Server) handleDeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Get user for logging
	user, err := s.db.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Prevent deleting yourself
	currentUserID := auth.GetUserID(c)
	if currentUserID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	if err := s.db.DeleteUser(id); err != nil {
		s.log.Error("Failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit event
	s.logAudit(c, "delete", "user", id, fmt.Sprintf("Deleted user: %s", user.Username))

	s.log.Info("User deleted", zap.String("username", user.Username), zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// logAudit logs an audit event
func (s *Server) logAudit(c *gin.Context, action, resourceType, resourceID, details string) {
	userID := auth.GetUserID(c)
	auditLog := &storage.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}

	if err := s.db.CreateAuditLog(auditLog); err != nil {
		s.log.Error("Failed to create audit log", zap.Error(err))
	}
}
