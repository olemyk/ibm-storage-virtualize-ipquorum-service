package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/config"
	"github.com/olemyk/ipquorum-management-platform/container-runtime-agent/pkg/runtime"
	"go.uber.org/zap"
)

// Server represents the HTTP API server
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	runtime runtime.Runtime
	router  *gin.Engine
	server  *http.Server
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, logger *zap.Logger, rt runtime.Runtime) *Server {
	// Set Gin mode based on config
	if cfg.IsProductionMode() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	s := &Server{
		config:  cfg,
		logger:  logger,
		runtime: rt,
		router:  gin.New(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures global middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(s.loggerMiddleware())

	// CORS middleware
	s.router.Use(s.corsMiddleware())

	// Request ID middleware
	s.router.Use(s.requestIDMiddleware())
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no auth required)
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/ready", s.handleReady)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Runtime info (no auth for now, will add auth middleware later)
		v1.GET("/runtime/info", s.handleRuntimeInfo)
		v1.GET("/runtime/version", s.handleRuntimeVersion)
		v1.GET("/runtime/ping", s.handleRuntimePing)

		// Container management
		containers := v1.Group("/containers")
		{
			containers.POST("", s.handleCreateContainer)
			containers.GET("", s.handleListContainers)
			containers.GET("/:id", s.handleInspectContainer)
			containers.POST("/:id/start", s.handleStartContainer)
			containers.POST("/:id/stop", s.handleStopContainer)
			containers.POST("/:id/restart", s.handleRestartContainer)
			containers.DELETE("/:id", s.handleRemoveContainer)
			containers.GET("/:id/logs", s.handleContainerLogs)
			containers.GET("/:id/stats", s.handleContainerStats)
		}

		// Image management
		images := v1.Group("/images")
		{
			images.GET("", s.handleListImages)
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.GetServerAddress()

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	s.logger.Info("Starting API server",
		zap.String("address", addr),
		zap.Bool("tls", s.config.Server.TLS.Enabled),
	)

	if s.config.Server.TLS.Enabled {
		return s.server.ListenAndServeTLS(
			s.config.Server.TLS.CertFile,
			s.config.Server.TLS.KeyFile,
		)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.server.Shutdown(ctx)
}

// Health check handlers
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(c *gin.Context) {
	// Check if runtime is accessible
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := s.runtime.Ping(ctx); err != nil {
		s.logger.Error("Runtime ping failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "runtime not accessible",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// Runtime info handlers
func (s *Server) handleRuntimeInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	info, err := s.runtime.Info(ctx)
	if err != nil {
		s.logger.Error("Failed to get runtime info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get runtime info: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (s *Server) handleRuntimeVersion(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	version, err := s.runtime.Version(ctx)
	if err != nil {
		s.logger.Error("Failed to get runtime version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get runtime version: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

func (s *Server) handleRuntimePing(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := s.runtime.Ping(ctx); err != nil {
		s.logger.Error("Runtime ping failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": fmt.Sprintf("runtime ping failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"runtime": s.runtime.Name(),
	})
}

// Container management handlers
func (s *Server) handleCreateContainer(c *gin.Context) {
	var req runtime.ContainerConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	// Debug logging
	s.logger.Info("Received container create request",
		zap.String("name", req.Name),
		zap.String("image", req.Image),
		zap.String("working_dir", req.WorkingDir),
		zap.Any("env", req.Env),
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	container, err := s.runtime.Create(ctx, &req)
	if err != nil {
		s.logger.Error("Failed to create container",
			zap.String("image", req.Image),
			zap.String("working_dir", req.WorkingDir),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create container: %v", err),
		})
		return
	}

	s.logger.Info("Container created",
		zap.String("id", container.ID),
		zap.String("image", req.Image),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":   container.ID,
		"name": container.Name,
	})
}

func (s *Server) handleListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "false") == "true"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Build filters map
	filters := make(map[string]string)
	if !all {
		filters["status"] = "running"
	}

	containers, err := s.runtime.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list containers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to list containers: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"containers": containers,
		"count":      len(containers),
	})
}

func (s *Server) handleInspectContainer(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	container, err := s.runtime.Inspect(ctx, id)
	if err != nil {
		s.logger.Error("Failed to inspect container",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("container not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, container)
}

func (s *Server) handleStartContainer(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := s.runtime.Start(ctx, id); err != nil {
		s.logger.Error("Failed to start container",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to start container: %v", err),
		})
		return
	}

	s.logger.Info("Container started", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{
		"status": "started",
		"id":     id,
	})
}

func (s *Server) handleStopContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10 * time.Second // default timeout

	ctx, cancel := context.WithTimeout(c.Request.Context(), 40*time.Second)
	defer cancel()

	if err := s.runtime.Stop(ctx, id, timeout); err != nil {
		s.logger.Error("Failed to stop container",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to stop container: %v", err),
		})
		return
	}

	s.logger.Info("Container stopped", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{
		"status": "stopped",
		"id":     id,
	})
}

func (s *Server) handleRestartContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10 * time.Second // default timeout

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	if err := s.runtime.Restart(ctx, id, timeout); err != nil {
		s.logger.Error("Failed to restart container",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to restart container: %v", err),
		})
		return
	}

	s.logger.Info("Container restarted", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{
		"status": "restarted",
		"id":     id,
	})
}

func (s *Server) handleRemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := s.runtime.Remove(ctx, id, force); err != nil {
		s.logger.Error("Failed to remove container",
			zap.String("id", id),
			zap.Bool("force", force),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to remove container: %v", err),
		})
		return
	}

	s.logger.Info("Container removed",
		zap.String("id", id),
		zap.Bool("force", force),
	)
	c.JSON(http.StatusOK, gin.H{
		"status": "removed",
		"id":     id,
	})
}

func (s *Server) handleContainerLogs(c *gin.Context) {
	id := c.Param("id")

	// Parse tail parameter
	tailStr := c.DefaultQuery("tail", "100")
	tail := 100
	if tailStr != "all" {
		if n, err := fmt.Sscanf(tailStr, "%d", &tail); err != nil || n != 1 {
			tail = 100
		}
	}

	opts := runtime.LogOptions{
		Follow:     c.DefaultQuery("follow", "false") == "true",
		Timestamps: c.DefaultQuery("timestamps", "false") == "true",
		Tail:       tail,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	logsReader, err := s.runtime.Logs(ctx, id, opts)
	if err != nil {
		s.logger.Error("Failed to get container logs",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get logs: %v", err),
		})
		return
	}
	defer logsReader.Close()

	// Read logs from reader
	logs, err := io.ReadAll(logsReader)
	if err != nil {
		s.logger.Error("Failed to read container logs",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to read logs: %v", err),
		})
		return
	}

	c.String(http.StatusOK, string(logs))
}

func (s *Server) handleContainerStats(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := s.runtime.Stats(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get container stats",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get stats: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
func (s *Server) handleListImages(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	images, err := s.runtime.ListImages(ctx)
	if err != nil {
		s.logger.Error("Failed to list images", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to list images: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
		"count":  len(images),
	})
}

// Made with Bob
