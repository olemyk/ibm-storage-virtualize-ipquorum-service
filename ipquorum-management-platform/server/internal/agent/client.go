package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an Agent HTTP client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	timeout    time.Duration
}

// Config holds agent client configuration
type Config struct {
	Host       string
	Port       int
	APIKey     string
	TLSEnabled bool
	TLSVerify  bool
	Timeout    time.Duration
}

// NewClient creates a new Agent client
func NewClient(cfg Config) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.TLSVerify,
		},
	}

	protocol := "http"
	if cfg.TLSEnabled {
		protocol = "https"
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: fmt.Sprintf("%s://%s:%d", protocol, cfg.Host, cfg.Port),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
		timeout: cfg.Timeout,
	}
}

// Health checks agent health
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// StartInstance starts an instance on the agent
func (c *Client) StartInstance(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/api/v1/instances/%s/start", c.baseURL, name)
	return c.doRequest(ctx, "POST", url, nil, nil)
}

// StopInstance stops an instance on the agent
func (c *Client) StopInstance(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/api/v1/instances/%s/stop", c.baseURL, name)
	return c.doRequest(ctx, "POST", url, nil, nil)
}

// RestartInstance restarts an instance on the agent
func (c *Client) RestartInstance(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/api/v1/instances/%s/restart", c.baseURL, name)
	return c.doRequest(ctx, "POST", url, nil, nil)
}

// GetStatus gets instance status from agent
func (c *Client) GetStatus(ctx context.Context, name string) (*InstanceStatus, error) {
	url := fmt.Sprintf("%s/api/v1/instances/%s/status", c.baseURL, name)

	var status InstanceStatus
	err := c.doRequest(ctx, "GET", url, nil, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// GetLogs gets instance logs from agent
func (c *Client) GetLogs(ctx context.Context, name string, lines int) (string, error) {
	url := fmt.Sprintf("%s/api/v1/instances/%s/logs?lines=%d", c.baseURL, name, lines)

	var response LogsResponse
	err := c.doRequest(ctx, "GET", url, nil, &response)
	if err != nil {
		return "", err
	}

	return response.Logs, nil
}

// CreateInstance creates an instance on the agent
func (c *Client) CreateInstance(ctx context.Context, req *CreateInstanceRequest) error {
	url := fmt.Sprintf("%s/api/v1/instances", c.baseURL)
	return c.doRequest(ctx, "POST", url, req, nil)
}

// DeleteInstance deletes an instance on the agent
func (c *Client) DeleteInstance(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/api/v1/instances/%s", c.baseURL, name)
	return c.doRequest(ctx, "DELETE", url, nil, nil)
}

// ListInstances lists all instances on the agent
func (c *Client) ListInstances(ctx context.Context) ([]InstanceInfo, error) {
	url := fmt.Sprintf("%s/api/v1/instances", c.baseURL)

	var instances []InstanceInfo
	err := c.doRequest(ctx, "GET", url, nil, &instances)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// doRequest performs HTTP request with API key authentication
func (c *Client) doRequest(ctx context.Context, method, url string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Response types

// InstanceStatus represents instance status from agent
type InstanceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Health    string `json:"health"`
	Message   string `json:"message"`
	Uptime    int64  `json:"uptime_seconds,omitempty"` // Uptime in seconds (0 if not running)
	StartedAt string `json:"started_at,omitempty"`     // ISO 8601 timestamp when started
}

// LogsResponse represents logs response from agent
type LogsResponse struct {
	Logs string `json:"logs"`
}

// InstanceInfo represents instance information from agent
type InstanceInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Health string `json:"health"`
}

// CreateInstanceRequest represents instance creation request
type CreateInstanceRequest struct {
	Name              string `json:"name"`
	APIEndpoint       string `json:"api_endpoint"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	EnableDownload    bool   `json:"enable_download"`
	EnableMkquorumapp bool   `json:"enable_mkquorumapp"`
	Partnersystem     string `json:"partnersystem,omitempty"`
	IPQuorumName      string `json:"ipquorum_name,omitempty"`
	IP6               bool   `json:"ip6"`
	PartnerIP6        bool   `json:"partnerip6"`
	NoMetadata        bool   `json:"nometadata"`
}
