package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/server/internal/agent"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// handleListServers lists all registered servers/agents
func (s *Server) handleListServers(c *gin.Context) {
	servers, err := s.db.ListServers()
	if err != nil {
		s.log.Error("Failed to list servers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// handleGetServer gets a specific server
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

// handleRegisterServer registers a new agent/server
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
		req.AgentPort = 8081 // Default to agent API port, not metrics port
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

// handleTestServer tests connection to a server
func (s *Server) handleTestServer(c *gin.Context) {
	id := c.Param("id")

	client, err := s.agentRegistry.Get(id)
	if err != nil {
		s.log.Error("Agent not found", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	ctx := c.Request.Context()
	if err := client.Health(ctx); err != nil {
		s.log.Warn("Agent health check failed",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	s.log.Info("Agent health check passed", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// handleDeleteServer unregisters a server
func (s *Server) handleDeleteServer(c *gin.Context) {
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

// handleUpdateServer updates a server's configuration
func (s *Server) handleUpdateServer(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Hostname   string `json:"hostname" binding:"required"`
		IPAddress  string `json:"ip_address" binding:"required"`
		AgentPort  int    `json:"agent_port"`
		APIKey     string `json:"api_key"`
		TLSEnabled bool   `json:"tls_enabled"`
		TLSVerify  bool   `json:"tls_verify"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AgentPort == 0 {
		req.AgentPort = 8081 // Default to agent API port
	}

	// Get existing server
	server, err := s.db.GetServer(id)
	if err != nil {
		s.log.Error("Failed to get server", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Update server fields
	server.Hostname = req.Hostname
	server.IPAddress = req.IPAddress
	server.AgentPort = req.AgentPort

	// Update in database
	if err := s.db.UpdateServer(server); err != nil {
		s.log.Error("Failed to update server", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If API key or TLS settings provided, update agent registry
	if req.APIKey != "" {
		// Unregister old agent
		s.agentRegistry.Unregister(id)

		// Register with new config
		cfg := agent.Config{
			Host:       req.IPAddress,
			Port:       req.AgentPort,
			APIKey:     req.APIKey,
			TLSEnabled: req.TLSEnabled,
			TLSVerify:  req.TLSVerify,
			Timeout:    30 * time.Second,
		}

		if err := s.agentRegistry.Register(id, cfg); err != nil {
			s.log.Error("Failed to reconnect to agent",
				zap.String("server_id", id),
				zap.String("host", req.IPAddress),
				zap.Error(err),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Server updated but failed to reconnect to agent: " + err.Error(),
				"server":  server,
				"warning": true,
			})
			return
		}
	}

	s.log.Info("Server updated successfully",
		zap.String("server_id", id),
		zap.String("hostname", server.Hostname),
		zap.String("ip", server.IPAddress),
	)

	c.JSON(http.StatusOK, server)
}

// handleGetServerStatus gets detailed status of a server including containers
func (s *Server) handleGetServerStatus(c *gin.Context) {
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

	// Try to get containers list
	instances := []interface{}{}
	if healthErr == nil {
		if containerList, err := client.ListContainers(ctx); err == nil {
			instances = make([]interface{}, len(containerList))
			for i, container := range containerList {
				instances[i] = container
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"server":    server,
		"status":    status,
		"instances": instances,
	})
}

// handleGetContainerLogs gets logs from a container on a specific server
func (s *Server) handleGetContainerLogs(c *gin.Context) {
	serverID := c.Param("id")
	containerID := c.Param("container_id")

	// Get tail parameter (default 100 lines)
	tail := 100
	if tailStr := c.Query("tail"); tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	// Get agent client
	client, err := s.agentRegistry.Get(serverID)
	if err != nil {
		s.log.Error("Agent not found",
			zap.String("server_id", serverID),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Get container logs from agent
	ctx := c.Request.Context()
	logs, err := client.GetContainerLogs(ctx, containerID, tail)
	if err != nil {
		s.log.Error("Failed to get container logs",
			zap.String("server_id", serverID),
			zap.String("container_id", containerID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return logs as plain text or JSON based on Accept header
	if c.GetHeader("Accept") == "text/plain" {
		c.String(http.StatusOK, logs)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"container_id": containerID,
			"server_id":    serverID,
			"tail":         tail,
			"logs":         logs,
		})
	}
}
