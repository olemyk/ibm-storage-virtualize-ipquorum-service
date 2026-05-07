package api

import (
	"net/http"
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

// handleGetServerStatus gets detailed status of a server including instances
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
