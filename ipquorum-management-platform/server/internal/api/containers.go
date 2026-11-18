package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/server/internal/service"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ContainerHandlers handles container-related HTTP requests
type ContainerHandlers struct {
	containerService *service.ContainerService
	logger           *zap.Logger
}

// NewContainerHandlers creates a new container handlers instance
func NewContainerHandlers(containerService *service.ContainerService, logger *zap.Logger) *ContainerHandlers {
	return &ContainerHandlers{
		containerService: containerService,
		logger:           logger,
	}
}

// RegisterRoutes registers container routes
func (h *ContainerHandlers) RegisterRoutes(r *gin.RouterGroup) {
	containers := r.Group("/containers")
	{
		containers.POST("", h.CreateContainer)
		containers.GET("", h.ListContainers)
		containers.GET("/:id", h.GetContainer)
		containers.DELETE("/:id", h.RemoveContainer)

		// Lifecycle operations
		containers.POST("/:id/start", h.StartContainer)
		containers.POST("/:id/stop", h.StopContainer)
		containers.POST("/:id/restart", h.RestartContainer)

		// Monitoring operations
		containers.GET("/:id/logs", h.GetContainerLogs)
		containers.GET("/:id/stats", h.GetContainerStats)
	}
}

// CreateContainer creates a new container instance
// @Summary Create container
// @Description Create a new containerized IP Quorum instance
// @Tags containers
// @Accept json
// @Produce json
// @Param container body storage.CreateContainerInstanceRequest true "Container configuration"
// @Success 201 {object} storage.ContainerInstance
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers [post]
func (h *ContainerHandlers) CreateContainer(c *gin.Context) {
	var req storage.CreateContainerInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	container, err := h.containerService.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create container",
			zap.String("instance_id", req.InstanceID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create container",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Container created",
		zap.String("container_id", container.ID),
		zap.String("instance_id", container.InstanceID),
	)

	c.JSON(http.StatusCreated, container)
}

// ListContainers lists all container instances
// @Summary List containers
// @Description Get a list of all container instances
// @Tags containers
// @Produce json
// @Success 200 {array} storage.ContainerInstance
// @Failure 500 {object} ErrorResponse
// @Router /api/containers [get]
func (h *ContainerHandlers) ListContainers(c *gin.Context) {
	containers, err := h.containerService.List(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list containers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list containers",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// GetContainer gets a container by ID
// @Summary Get container
// @Description Get container details by ID
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} storage.ContainerInstance
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id} [get]
func (h *ContainerHandlers) GetContainer(c *gin.Context) {
	id := c.Param("id")

	container, err := h.containerService.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get container",
			zap.String("container_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Container not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, container)
}

// RemoveContainer removes a container
// @Summary Remove container
// @Description Remove a container instance
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Param force query bool false "Force removal"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id} [delete]
func (h *ContainerHandlers) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.Query("force") == "true"

	err := h.containerService.Remove(c.Request.Context(), id, force)
	if err != nil {
		h.logger.Error("Failed to remove container",
			zap.String("container_id", id),
			zap.Bool("force", force),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to remove container",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Container removed",
		zap.String("container_id", id),
		zap.Bool("force", force),
	)

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Container removed successfully",
	})
}

// StartContainer starts a container
// @Summary Start container
// @Description Start a stopped container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id}/start [post]
func (h *ContainerHandlers) StartContainer(c *gin.Context) {
	id := c.Param("id")

	err := h.containerService.Start(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to start container",
			zap.String("container_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to start container",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Container started", zap.String("container_id", id))

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Container started successfully",
	})
}

// StopContainer stops a container
// @Summary Stop container
// @Description Stop a running container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Param timeout query int false "Stop timeout in seconds" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id}/stop [post]
func (h *ContainerHandlers) StopContainer(c *gin.Context) {
	id := c.Param("id")

	timeout := 10 // default
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = t
		}
	}

	err := h.containerService.Stop(c.Request.Context(), id, timeout)
	if err != nil {
		h.logger.Error("Failed to stop container",
			zap.String("container_id", id),
			zap.Int("timeout", timeout),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to stop container",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Container stopped",
		zap.String("container_id", id),
		zap.Int("timeout", timeout),
	)

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Container stopped successfully",
	})
}

// RestartContainer restarts a container
// @Summary Restart container
// @Description Restart a container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Param timeout query int false "Stop timeout in seconds" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id}/restart [post]
func (h *ContainerHandlers) RestartContainer(c *gin.Context) {
	id := c.Param("id")

	timeout := 10 // default
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = t
		}
	}

	err := h.containerService.Restart(c.Request.Context(), id, timeout)
	if err != nil {
		h.logger.Error("Failed to restart container",
			zap.String("container_id", id),
			zap.Int("timeout", timeout),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to restart container",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Container restarted",
		zap.String("container_id", id),
		zap.Int("timeout", timeout),
	)

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Container restarted successfully",
	})
}

// GetContainerLogs gets container logs
// @Summary Get container logs
// @Description Retrieve logs from a container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Param tail query int false "Number of lines to tail" default(100)
// @Success 200 {object} storage.ContainerLogs
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id}/logs [get]
func (h *ContainerHandlers) GetContainerLogs(c *gin.Context) {
	id := c.Param("id")

	tail := 100 // default
	if tailStr := c.Query("tail"); tailStr != "" {
		if t, err := strconv.Atoi(tailStr); err == nil && t > 0 {
			tail = t
		}
	}

	logs, err := h.containerService.GetLogs(c.Request.Context(), id, tail)
	if err != nil {
		h.logger.Error("Failed to get container logs",
			zap.String("container_id", id),
			zap.Int("tail", tail),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get container logs",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetContainerStats gets container statistics
// @Summary Get container stats
// @Description Retrieve real-time statistics from a container
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} storage.ContainerStats
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/containers/{id}/stats [get]
func (h *ContainerHandlers) GetContainerStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.containerService.GetStats(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get container stats",
			zap.String("container_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get container stats",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SuccessResponse represents a successful operation response
type SuccessResponse struct {
	Message string `json:"message"`
}

// Made with Bob
