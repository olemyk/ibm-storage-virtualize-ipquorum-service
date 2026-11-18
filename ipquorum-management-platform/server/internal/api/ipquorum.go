package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/server/internal/service"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// IPQuorumHandler handles IPQuorum container API requests
type IPQuorumHandler struct {
	service *service.IPQuorumService
	logger  *zap.Logger
}

// NewIPQuorumHandler creates a new IPQuorum handler
func NewIPQuorumHandler(s *service.IPQuorumService, logger *zap.Logger) *IPQuorumHandler {
	return &IPQuorumHandler{
		service: s,
		logger:  logger,
	}
}

// CreateIPQuorumContainer creates a new IPQuorum container
// POST /api/ipquorum/containers
func (h *IPQuorumHandler) CreateIPQuorumContainer(c *gin.Context) {
	var req storage.CreateIPQuorumContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	h.logger.Info("Creating IPQuorum container",
		zap.String("name", req.Name),
		zap.String("agent_id", req.AgentID),
		zap.String("api_endpoint", req.IPQuorumConfig.APIEndpoint))

	container, err := h.service.CreateIPQuorumContainer(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create IPQuorum container", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, container)
}

// GetIPQuorumContainer retrieves an IPQuorum container by ID
// GET /api/ipquorum/containers/:id
func (h *IPQuorumHandler) GetIPQuorumContainer(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Getting IPQuorum container", zap.String("id", containerID))

	container, err := h.service.GetIPQuorumContainer(c.Request.Context(), containerID)
	if err != nil {
		h.logger.Error("Failed to get IPQuorum container", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, container)
}

// ListIPQuorumContainers lists all IPQuorum containers
// GET /api/ipquorum/containers
func (h *IPQuorumHandler) ListIPQuorumContainers(c *gin.Context) {
	// Optional agent_id filter
	agentID := c.Query("agent_id")

	h.logger.Info("Listing IPQuorum containers", zap.String("agent_id", agentID))

	containers, err := h.service.ListIPQuorumContainers(c.Request.Context(), agentID)
	if err != nil {
		h.logger.Error("Failed to list IPQuorum containers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// DeleteIPQuorumContainer deletes an IPQuorum container
// DELETE /api/ipquorum/containers/:id
func (h *IPQuorumHandler) DeleteIPQuorumContainer(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Deleting IPQuorum container", zap.String("id", containerID))

	if err := h.service.DeleteIPQuorumContainer(c.Request.Context(), containerID); err != nil {
		h.logger.Error("Failed to delete IPQuorum container", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartIPQuorumContainer starts an IPQuorum container
// POST /api/ipquorum/containers/:id/start
func (h *IPQuorumHandler) StartIPQuorumContainer(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Starting IPQuorum container", zap.String("id", containerID))

	if err := h.service.StartIPQuorumContainer(c.Request.Context(), containerID); err != nil {
		h.logger.Error("Failed to start IPQuorum container", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "started"})
}

// StopIPQuorumContainer stops an IPQuorum container
// POST /api/ipquorum/containers/:id/stop
func (h *IPQuorumHandler) StopIPQuorumContainer(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Stopping IPQuorum container", zap.String("id", containerID))

	if err := h.service.StopIPQuorumContainer(c.Request.Context(), containerID); err != nil {
		h.logger.Error("Failed to stop IPQuorum container", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stopped"})
}

// RestartIPQuorumContainer restarts an IPQuorum container
// POST /api/ipquorum/containers/:id/restart
func (h *IPQuorumHandler) RestartIPQuorumContainer(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Restarting IPQuorum container", zap.String("id", containerID))

	if err := h.service.RestartIPQuorumContainer(c.Request.Context(), containerID); err != nil {
		h.logger.Error("Failed to restart IPQuorum container", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "restarted"})
}

// GetIPQuorumContainerLogs retrieves logs from an IPQuorum container
// GET /api/ipquorum/containers/:id/logs
func (h *IPQuorumHandler) GetIPQuorumContainerLogs(c *gin.Context) {
	containerID := c.Param("id")

	// Parse tail parameter (default: 100)
	tail := 100
	if tailStr := c.Query("tail"); tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	h.logger.Info("Getting IPQuorum container logs",
		zap.String("id", containerID),
		zap.Int("tail", tail))

	logs, err := h.service.GetIPQuorumContainerLogs(c.Request.Context(), containerID, tail)
	if err != nil {
		h.logger.Error("Failed to get IPQuorum container logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetIPQuorumContainerStats retrieves stats from an IPQuorum container
// GET /api/ipquorum/containers/:id/stats
func (h *IPQuorumHandler) GetIPQuorumContainerStats(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Getting IPQuorum container stats", zap.String("id", containerID))

	stats, err := h.service.GetIPQuorumContainerStats(c.Request.Context(), containerID)
	if err != nil {
		h.logger.Error("Failed to get IPQuorum container stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportIPQuorumConfig exports IPQuorum configuration as JSON
// GET /api/ipquorum/containers/:id/config/export
func (h *IPQuorumHandler) ExportIPQuorumConfig(c *gin.Context) {
	containerID := c.Param("id")

	h.logger.Info("Exporting IPQuorum config", zap.String("id", containerID))

	configJSON, err := h.service.ExportIPQuorumConfig(c.Request.Context(), containerID)
	if err != nil {
		h.logger.Error("Failed to export IPQuorum config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=ipquorum-config.json")
	c.Data(http.StatusOK, "application/json", configJSON)
}

// RegisterRoutes registers IPQuorum API routes with Gin router
func (h *IPQuorumHandler) RegisterRoutes(r *gin.RouterGroup) {
	ipquorum := r.Group("/ipquorum")
	{
		// IPQuorum container management
		ipquorum.POST("/containers", h.CreateIPQuorumContainer)
		ipquorum.GET("/containers", h.ListIPQuorumContainers)
		ipquorum.GET("/containers/:id", h.GetIPQuorumContainer)
		ipquorum.DELETE("/containers/:id", h.DeleteIPQuorumContainer)

		// Container lifecycle
		ipquorum.POST("/containers/:id/start", h.StartIPQuorumContainer)
		ipquorum.POST("/containers/:id/stop", h.StopIPQuorumContainer)
		ipquorum.POST("/containers/:id/restart", h.RestartIPQuorumContainer)

		// Container monitoring
		ipquorum.GET("/containers/:id/logs", h.GetIPQuorumContainerLogs)
		ipquorum.GET("/containers/:id/stats", h.GetIPQuorumContainerStats)

		// Configuration export
		ipquorum.GET("/containers/:id/config/export", h.ExportIPQuorumConfig)
	}
}

// Made with Bob
