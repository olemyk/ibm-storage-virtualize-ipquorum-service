package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/agent/internal/config"
	"github.com/olemyk/ipquorum-platform/agent/internal/executor"
	"go.uber.org/zap"
)

// Server represents the agent API server
type Server struct {
	router   *gin.Engine
	executor *executor.Executor
	config   *config.Config
	log      *zap.Logger
}

// NewServer creates a new agent API server
func NewServer(cfg *config.Config, exec *executor.Executor, log *zap.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		router:   gin.New(),
		executor: exec,
		config:   cfg,
		log:      log,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logging middleware
	s.router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		s.log.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", c.ClientIP()))
	})

	// API Key authentication middleware
	s.router.Use(func(c *gin.Context) {
		// Skip auth for health endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != s.config.Agent.APIKey {
			s.log.Warn("Unauthorized request", zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	})
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check (no auth required)
	s.router.GET("/health", s.handleHealth)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Agent info
		v1.GET("/info", s.handleInfo)

		// Instance management
		instances := v1.Group("/instances")
		{
			instances.GET("", s.handleListInstances)
			instances.POST("", s.handleCreateInstance)
			instances.GET("/:name/status", s.handleGetStatus)
			instances.GET("/:name/logs", s.handleGetLogs)
			instances.POST("/:name/start", s.handleStartInstance)
			instances.POST("/:name/stop", s.handleStopInstance)
			instances.POST("/:name/restart", s.handleRestartInstance)
			instances.DELETE("/:name", s.handleDeleteInstance)
		}
	}
}

// Run starts the API server
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Agent.Host, s.config.Agent.Port)
	s.log.Info("Starting agent API server", zap.String("address", addr))
	return s.router.Run(addr)
}

// Health check handler
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"agent":     s.config.Agent.InstanceName,
	})
}

// Info handler
func (s *Server) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"agent_name": s.config.Agent.InstanceName,
		"version":    "1.0.0",
		"script":     s.config.Script.Path,
	})
}

// List instances handler
func (s *Server) handleListInstances(c *gin.Context) {
	output, err := s.executor.ListInstances(c.Request.Context())
	if err != nil {
		s.log.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output": output,
	})
}

// Create instance handler
func (s *Server) handleCreateInstance(c *gin.Context) {
	var req struct {
		// Required fields
		Name        string `json:"name" binding:"required"`
		APIEndpoint string `json:"api_endpoint" binding:"required"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`

		// Documentation fields (optional)
		StorageSystem      string `json:"storage_system"`
		StorageDescription string `json:"storage_description"`
		StorageLocation    string `json:"storage_location"`
		IPQuorumName       string `json:"ipquorum_name"`

		// Download configuration (optional, defaults to true)
		DownloadEnabled *bool `json:"download_enabled"`

		// mkquorumapp configuration (optional)
		MkQuorumAppEnabled *bool  `json:"mkquorumapp_enabled"`
		PartnerSystem      string `json:"partnersystem"`
		IP6                *bool  `json:"ip6"`
		PartnerIP6         *bool  `json:"partnerip6"`
		NoMetadata         *bool  `json:"nometadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate: partnersystem is mandatory when mkquorumapp_enabled is true
	if req.MkQuorumAppEnabled != nil && *req.MkQuorumAppEnabled {
		if req.PartnerSystem == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "partnersystem is required when mkquorumapp_enabled is true",
			})
			return
		}
	}

	// Set defaults for optional fields
	downloadEnabled := true
	if req.DownloadEnabled != nil {
		downloadEnabled = *req.DownloadEnabled
	}

	mkquorumappEnabled := false
	if req.MkQuorumAppEnabled != nil {
		mkquorumappEnabled = *req.MkQuorumAppEnabled
	}

	ip6 := false
	if req.IP6 != nil {
		ip6 = *req.IP6
	}

	partnerip6 := false
	if req.PartnerIP6 != nil {
		partnerip6 = *req.PartnerIP6
	}

	nometadata := false
	if req.NoMetadata != nil {
		nometadata = *req.NoMetadata
	}

	// Set default storage system if not provided
	storageSystem := req.StorageSystem
	if storageSystem == "" {
		storageSystem = "N/A"
	}

	// Set default IPQuorum name if not provided (sanitize instance name)
	ipquorumName := req.IPQuorumName
	if ipquorumName == "" {
		// Remove dashes and underscores, truncate to 20 chars
		ipquorumName = strings.Map(func(r rune) rune {
			if r == '-' || r == '_' {
				return -1
			}
			return r
		}, req.Name)
		if len(ipquorumName) > 20 {
			ipquorumName = ipquorumName[:20]
		}
	}

	// Create instance with full request
	if err := s.executor.CreateInstance(c.Request.Context(), executor.CreateInstanceRequest{
		Name:               req.Name,
		APIEndpoint:        req.APIEndpoint,
		Username:           req.Username,
		Password:           req.Password,
		StorageSystem:      storageSystem,
		StorageDescription: req.StorageDescription,
		StorageLocation:    req.StorageLocation,
		IPQuorumName:       ipquorumName,
		DownloadEnabled:    downloadEnabled,
		MkQuorumAppEnabled: mkquorumappEnabled,
		PartnerSystem:      req.PartnerSystem,
		IP6:                ip6,
		PartnerIP6:         partnerip6,
		NoMetadata:         nometadata,
	}); err != nil {
		s.log.Error("Failed to create instance", zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance created successfully"})
}

// Get status handler
func (s *Server) handleGetStatus(c *gin.Context) {
	name := c.Param("name")

	output, err := s.executor.GetStatus(c.Request.Context(), name)

	// Handle different exit codes from ipqm status command:
	// - Exit 0: running/active (success)
	// - Exit 1: instance not found (error)
	// - Exit 3: inactive/stopped (valid state, not an error)
	if err != nil {
		errStr := err.Error()

		// Check if it's exit status 3 (inactive/stopped)
		if strings.Contains(errStr, "exit status 3") {
			// Instance is stopped/inactive - this is a valid state
			s.log.Info("Instance is inactive", zap.String("name", name))
			c.JSON(http.StatusOK, gin.H{
				"name":           name,
				"status":         "inactive",
				"uptime_seconds": 0,
			})
			return
		}

		// Check if it's exit status 1 (not found)
		if strings.Contains(errStr, "exit status 1") {
			s.log.Warn("Instance not found", zap.String("name", name))
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Instance %s not found", name),
			})
			return
		}

		// Other errors
		s.log.Error("Failed to get status", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Success - parse the status and uptime from systemd output
	statusInfo := s.parseSystemdStatus(output)

	response := gin.H{
		"name":   name,
		"status": statusInfo.Status,
	}

	if statusInfo.Uptime > 0 {
		response["uptime_seconds"] = statusInfo.Uptime
		response["started_at"] = statusInfo.StartedAt
	}

	c.JSON(http.StatusOK, response)
}

// SystemdStatusInfo holds parsed systemd status information
type SystemdStatusInfo struct {
	Status    string
	Uptime    int64  // seconds
	StartedAt string // ISO 8601 timestamp
}

// parseSystemdStatus extracts the status and uptime from systemd status output
func (s *Server) parseSystemdStatus(output string) SystemdStatusInfo {
	info := SystemdStatusInfo{
		Status: "unknown",
	}

	// Look for the "Active:" line in systemd output
	// Format: "Active: active (running) since Wed 2026-05-06 15:52:02 CEST; 2min 33s ago"
	// or "Active: inactive (dead)"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Active:") {
			// Extract the status: "Active: active (running)" -> "running"
			// or "Active: inactive (dead)" -> "inactive"
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// parts[1] is "active" or "inactive"
				// parts[2] is "(running)" or "(dead)"
				statusInParens := strings.Trim(parts[2], "()")

				// Map systemd states to our states
				switch statusInParens {
				case "running":
					info.Status = "running"
					// Parse start time and calculate uptime
					// Format: "Active: active (running) since Wed 2026-05-06 15:52:02 CEST; 2min 33s ago"
					if sinceIdx := strings.Index(line, "since "); sinceIdx != -1 {
						afterSince := line[sinceIdx+6:] // Skip "since "
						// Find the semicolon that separates timestamp from "ago" part
						if semiIdx := strings.Index(afterSince, ";"); semiIdx != -1 {
							timestampStr := strings.TrimSpace(afterSince[:semiIdx])
							// Parse the timestamp
							startTime, err := s.parseSystemdTimestamp(timestampStr)
							if err == nil {
								info.StartedAt = startTime.Format(time.RFC3339)
								info.Uptime = int64(time.Since(startTime).Seconds())
							}
						}
					}
				case "dead":
					info.Status = "inactive"
				default:
					info.Status = statusInParens
				}
			}
			break
		}
	}

	return info
}

// parseSystemdTimestamp parses systemd timestamp format
// Example: "Wed 2026-05-06 15:52:02 CEST"
func (s *Server) parseSystemdTimestamp(timestamp string) (time.Time, error) {
	// Systemd format: "Day YYYY-MM-DD HH:MM:SS TZ"
	// We'll try multiple formats
	formats := []string{
		"Mon 2006-01-02 15:04:05 MST",
		"Mon 2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

// Get logs handler
func (s *Server) handleGetLogs(c *gin.Context) {
	name := c.Param("name")
	lines := 100

	if l := c.Query("lines"); l != "" {
		fmt.Sscanf(l, "%d", &lines)
	}

	output, err := s.executor.GetLogs(c.Request.Context(), name, lines)
	if err != nil {
		s.log.Error("Failed to get logs", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name": name,
		"logs": output,
	})
}

// Start instance handler
func (s *Server) handleStartInstance(c *gin.Context) {
	name := c.Param("name")

	if err := s.executor.StartInstance(c.Request.Context(), name); err != nil {
		s.log.Error("Failed to start instance", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance started successfully"})
}

// Stop instance handler
func (s *Server) handleStopInstance(c *gin.Context) {
	name := c.Param("name")

	if err := s.executor.StopInstance(c.Request.Context(), name); err != nil {
		s.log.Error("Failed to stop instance", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance stopped successfully"})
}

// Restart instance handler
func (s *Server) handleRestartInstance(c *gin.Context) {
	name := c.Param("name")

	if err := s.executor.RestartInstance(c.Request.Context(), name); err != nil {
		s.log.Error("Failed to restart instance", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance restarted successfully"})
}

// Delete instance handler
func (s *Server) handleDeleteInstance(c *gin.Context) {
	name := c.Param("name")
	force := c.Query("force") == "true"

	if err := s.executor.DeleteInstance(c.Request.Context(), name, force); err != nil {
		s.log.Error("Failed to delete instance", zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instance deleted successfully"})
}
