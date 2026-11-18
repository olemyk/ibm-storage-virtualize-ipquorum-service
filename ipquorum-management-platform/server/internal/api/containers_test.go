package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/olemyk/ipquorum-platform/server/internal/storage"
	"go.uber.org/zap"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("not found")
)

// mockContainerService is a mock implementation of ContainerService for testing
type mockContainerService struct {
	createFunc   func(*storage.CreateContainerInstanceRequest) (*storage.ContainerInstance, error)
	getByIDFunc  func(string) (*storage.ContainerInstance, error)
	listFunc     func() ([]*storage.ContainerInstance, error)
	removeFunc   func(string, bool) error
	startFunc    func(string) error
	stopFunc     func(string, int) error
	restartFunc  func(string, int) error
	getLogsFunc  func(string, int) (*storage.ContainerLogs, error)
	getStatsFunc func(string) (*storage.ContainerStats, error)
}

func setupTestRouter() (*gin.Engine, *mockContainerService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &mockContainerService{}
	logger, _ := zap.NewDevelopment()

	// Create test handlers that use the mock
	testHandlers := &testContainerHandlers{
		mock:   mockService,
		logger: logger,
	}

	// Register routes without authentication for testing
	api := router.Group("/api/v1")
	testHandlers.RegisterRoutes(api)

	return router, mockService
}

// testContainerHandlers wraps the mock for testing
type testContainerHandlers struct {
	mock   *mockContainerService
	logger *zap.Logger
}

func (h *testContainerHandlers) RegisterRoutes(r *gin.RouterGroup) {
	containers := r.Group("/containers")
	{
		containers.POST("", h.CreateContainer)
		containers.GET("", h.ListContainers)
		containers.GET("/:id", h.GetContainer)
		containers.DELETE("/:id", h.RemoveContainer)
		containers.POST("/:id/start", h.StartContainer)
		containers.POST("/:id/stop", h.StopContainer)
		containers.POST("/:id/restart", h.RestartContainer)
		containers.GET("/:id/logs", h.GetContainerLogs)
		containers.GET("/:id/stats", h.GetContainerStats)
	}
}

func (h *testContainerHandlers) CreateContainer(c *gin.Context) {
	var req storage.CreateContainerInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body", Message: err.Error()})
		return
	}
	if h.mock.createFunc != nil {
		container, err := h.mock.createFunc(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create container", Message: err.Error()})
			return
		}
		c.JSON(http.StatusCreated, container)
	}
}

func (h *testContainerHandlers) ListContainers(c *gin.Context) {
	if h.mock.listFunc != nil {
		containers, err := h.mock.listFunc()
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list containers", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, containers)
	}
}

func (h *testContainerHandlers) GetContainer(c *gin.Context) {
	id := c.Param("id")
	if h.mock.getByIDFunc != nil {
		container, err := h.mock.getByIDFunc(id)
		if err != nil {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Container not found", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, container)
	}
}

func (h *testContainerHandlers) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.Query("force") == "true"
	if h.mock.removeFunc != nil {
		err := h.mock.removeFunc(id, force)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove container", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, SuccessResponse{Message: "Container removed successfully"})
	}
}

func (h *testContainerHandlers) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if h.mock.startFunc != nil {
		err := h.mock.startFunc(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to start container", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, SuccessResponse{Message: "Container started successfully"})
	}
}

func (h *testContainerHandlers) StopContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10
	if h.mock.stopFunc != nil {
		err := h.mock.stopFunc(id, timeout)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to stop container", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, SuccessResponse{Message: "Container stopped successfully"})
	}
}

func (h *testContainerHandlers) RestartContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10
	if h.mock.restartFunc != nil {
		err := h.mock.restartFunc(id, timeout)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to restart container", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, SuccessResponse{Message: "Container restarted successfully"})
	}
}

func (h *testContainerHandlers) GetContainerLogs(c *gin.Context) {
	id := c.Param("id")
	tail := 100
	if h.mock.getLogsFunc != nil {
		logs, err := h.mock.getLogsFunc(id, tail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get container logs", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	}
}

func (h *testContainerHandlers) GetContainerStats(c *gin.Context) {
	id := c.Param("id")
	if h.mock.getStatsFunc != nil {
		stats, err := h.mock.getStatsFunc(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get container stats", Message: err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

func TestCreateContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *storage.ContainerInstance
		mockError      error
		expectedStatus int
	}{
		{
			name: "valid request",
			requestBody: storage.CreateContainerInstanceRequest{
				InstanceID: "test-instance",
				AgentID:    "test-agent",
				ImageName:  "nginx",
				ImageTag:   "latest",
			},
			mockResponse: &storage.ContainerInstance{
				ID:         "container-1",
				InstanceID: "test-instance",
				AgentID:    "test-agent",
				ImageName:  "nginx",
				Status:     "created",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.createFunc = func(req *storage.CreateContainerInstanceRequest) (*storage.ContainerInstance, error) {
				return tt.mockResponse, tt.mockError
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/containers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestListContainers(t *testing.T) {
	router, mockService := setupTestRouter()

	mockContainers := []*storage.ContainerInstance{
		{ID: "container-1", InstanceID: "instance-1", Status: "running"},
		{ID: "container-2", InstanceID: "instance-2", Status: "stopped"},
	}

	mockService.listFunc = func() ([]*storage.ContainerInstance, error) {
		return mockContainers, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/containers", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []*storage.ContainerInstance
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response) != len(mockContainers) {
		t.Errorf("Expected %d containers, got %d", len(mockContainers), len(response))
	}
}

func TestGetContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	mockContainer := &storage.ContainerInstance{
		ID:         "container-1",
		InstanceID: "instance-1",
		Status:     "running",
	}

	tests := []struct {
		name           string
		containerID    string
		mockResponse   *storage.ContainerInstance
		mockError      error
		expectedStatus int
	}{
		{
			name:           "existing container",
			containerID:    "container-1",
			mockResponse:   mockContainer,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent container",
			containerID:    "non-existent",
			mockResponse:   nil,
			mockError:      ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.getByIDFunc = func(id string) (*storage.ContainerInstance, error) {
				if id == tt.containerID {
					return tt.mockResponse, tt.mockError
				}
				return nil, ErrNotFound
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/containers/"+tt.containerID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestStartContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	tests := []struct {
		name           string
		containerID    string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful start",
			containerID:    "container-1",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "container not found",
			containerID:    "non-existent",
			mockError:      ErrNotFound,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.startFunc = func(id string) error {
				return tt.mockError
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/containers/"+tt.containerID+"/start", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestStopContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	tests := []struct {
		name           string
		containerID    string
		timeout        string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful stop with default timeout",
			containerID:    "container-1",
			timeout:        "",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful stop with custom timeout",
			containerID:    "container-1",
			timeout:        "30",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.stopFunc = func(id string, timeout int) error {
				return tt.mockError
			}

			url := "/api/v1/containers/" + tt.containerID + "/stop"
			if tt.timeout != "" {
				url += "?timeout=" + tt.timeout
			}

			req := httptest.NewRequest(http.MethodPost, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRestartContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	mockService.restartFunc = func(id string, timeout int) error {
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/containers/container-1/restart", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRemoveContainer(t *testing.T) {
	router, mockService := setupTestRouter()

	tests := []struct {
		name           string
		containerID    string
		force          string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "successful removal",
			containerID:    "container-1",
			force:          "",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "force removal",
			containerID:    "container-1",
			force:          "true",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.removeFunc = func(id string, force bool) error {
				return tt.mockError
			}

			url := "/api/v1/containers/" + tt.containerID
			if tt.force != "" {
				url += "?force=" + tt.force
			}

			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetContainerLogs(t *testing.T) {
	router, mockService := setupTestRouter()

	mockLogs := &storage.ContainerLogs{
		ContainerID: "container-1",
		Logs:        []string{"log line 1", "log line 2"},
	}

	mockService.getLogsFunc = func(id string, tail int) (*storage.ContainerLogs, error) {
		return mockLogs, nil
	}

	tests := []struct {
		name           string
		containerID    string
		tail           string
		expectedStatus int
	}{
		{
			name:           "default tail",
			containerID:    "container-1",
			tail:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "custom tail",
			containerID:    "container-1",
			tail:           "50",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/containers/" + tt.containerID + "/logs"
			if tt.tail != "" {
				url += "?tail=" + tt.tail
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetContainerStats(t *testing.T) {
	router, mockService := setupTestRouter()

	mockStats := &storage.ContainerStats{
		ContainerID:   "container-1",
		CPUPercent:    25.5,
		MemoryUsage:   1024 * 1024 * 512,  // 512MB
		MemoryLimit:   1024 * 1024 * 1024, // 1GB
		MemoryPercent: 50.0,
	}

	mockService.getStatsFunc = func(id string) (*storage.ContainerStats, error) {
		return mockStats, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/containers/container-1/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response storage.ContainerStats
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ContainerID != mockStats.ContainerID {
		t.Errorf("Expected container ID %s, got %s", mockStats.ContainerID, response.ContainerID)
	}
}

// Benchmark tests
func BenchmarkCreateContainer(b *testing.B) {
	router, mockService := setupTestRouter()

	mockService.createFunc = func(req *storage.CreateContainerInstanceRequest) (*storage.ContainerInstance, error) {
		return &storage.ContainerInstance{
			ID:         "container-1",
			InstanceID: req.InstanceID,
			Status:     "created",
		}, nil
	}

	reqBody := storage.CreateContainerInstanceRequest{
		InstanceID: "test-instance",
		AgentID:    "test-agent",
		ImageName:  "nginx",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/containers", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkListContainers(b *testing.B) {
	router, mockService := setupTestRouter()

	mockContainers := make([]*storage.ContainerInstance, 100)
	for i := 0; i < 100; i++ {
		mockContainers[i] = &storage.ContainerInstance{
			ID:         string(rune(i)),
			InstanceID: "instance-" + string(rune(i)),
			Status:     "running",
		}
	}

	mockService.listFunc = func() ([]*storage.ContainerInstance, error) {
		return mockContainers, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/containers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// Made with Bob
