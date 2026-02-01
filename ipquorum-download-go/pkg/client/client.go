package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/olemyk/ipquorum-go/internal/errors"
	"github.com/olemyk/ipquorum-go/pkg/config"
	"github.com/olemyk/ipquorum-go/pkg/logger"
)

// Client represents an IPQuorum API client
type Client struct {
	config     *config.Config
	httpClient *http.Client
	token      string
	logger     *logger.Logger
	baseURL    string
}

// NewClient creates a new IPQuorum client
func NewClient(cfg *config.Config, log *logger.Logger) *Client {
	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifySSL,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	baseURL := fmt.Sprintf("https://%s:%d", cfg.APIEndpoint, cfg.APIPort)

	log.Debug("Initialized client for %s", baseURL)
	log.Debug("SSL verification: %t", cfg.VerifySSL)

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		logger:     log,
		baseURL:    baseURL,
	}
}

// PreflightCheck validates endpoint reachability
func (c *Client) PreflightCheck() error {
	c.logger.Info("Pre-flight: validating inputs & endpoint reachability...")

	url := fmt.Sprintf("%s/rest/v1/auth", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.NewNetworkError("failed to create request", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "x509") || strings.Contains(err.Error(), "certificate") {
			c.logger.Error("Pre-flight: SSL/TLS error. Try --insecure or verify connectivity.")
			return errors.NewNetworkError("SSL/TLS error", err)
		}
		c.logger.Error("Pre-flight: Connection error. Verify endpoint and network connectivity.")
		return errors.NewNetworkError("connection error", err)
	}
	defer resp.Body.Close()

	c.logger.Info("Pre-flight: endpoint status: %d", resp.StatusCode)

	// Log response headers in debug mode
	if c.logger.IsDebug() {
		c.logger.Debug("Pre-flight: endpoint headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				c.logger.Debug("  %s: %s", key, value)
			}
		}
	}

	// Accept various status codes that indicate the endpoint is reachable
	switch resp.StatusCode {
	case 200, 401, 404, 405:
		c.logger.Info("Pre-flight: endpoint is reachable.")
		return nil
	default:
		c.logger.Warning("Pre-flight: unexpected status %d. Proceeding may fail.", resp.StatusCode)
		return nil
	}
}

// Authenticate performs authentication with retry logic
func (c *Client) Authenticate() error {
	c.logger.Info("Get Token, please wait...")

	url := fmt.Sprintf("%s/rest/v1/auth", c.baseURL)

	for attempt := 1; attempt <= c.config.MaxRetries; attempt++ {
		c.logger.Info("Auth attempt %d/%d...", attempt, c.config.MaxRetries)

		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			return errors.NewAuthenticationError("failed to create request", 0)
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Auth-Username", c.config.Username)
		req.Header.Set("X-Auth-Password", c.config.Password)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.Warning("Attempt %d failed with exception: %v", attempt, err)
			delay := c.config.BaseDelay + time.Duration(attempt)*time.Second + time.Duration(rand.Intn(2))*time.Second
			time.Sleep(delay)
			continue
		}

		statusCode := resp.StatusCode
		c.logger.Debug("Auth response status: %d", statusCode)

		// Fail fast on authentication errors (don't retry)
		if statusCode == 401 {
			resp.Body.Close()
			c.logger.Error("Authentication failed: Invalid username or password")
			return errors.NewAuthenticationError("Invalid username or password (HTTP 401)", 401)
		}

		if statusCode == 403 {
			resp.Body.Close()
			c.logger.Error("Authentication failed: Insufficient permissions or invalid credentials")
			return errors.NewAuthenticationError("Insufficient permissions or invalid credentials (HTTP 403)", 403)
		}

		// Handle rate limiting
		if statusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var delay time.Duration
			if retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					delay = time.Duration(seconds) * time.Second
				} else {
					delay = c.config.BaseDelay*time.Duration(attempt) + time.Duration(rand.Intn(3))*time.Second
				}
			} else {
				delay = c.config.BaseDelay*time.Duration(attempt) + time.Duration(rand.Intn(3))*time.Second
			}

			c.logger.Warning("Rate limited (429). Sleeping %v and retrying...", delay)
			resp.Body.Close()
			time.Sleep(delay)
			continue
		}

		// Try to extract token from headers
		token := resp.Header.Get("X-Auth-Token")
		if token == "" {
			token = resp.Header.Get("x-auth-token")
		}
		if token == "" {
			token = resp.Header.Get("Authorization")
		}
		if token == "" {
			token = resp.Header.Get("authorization")
		}

		// If not in headers, try JSON body
		if token == "" || token == "null" {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				var jsonBody map[string]interface{}
				if err := json.Unmarshal(body, &jsonBody); err == nil {
					// Try various token field names
					if t, ok := jsonBody["token"].(string); ok && t != "" {
						token = t
					} else if t, ok := jsonBody["access_token"].(string); ok && t != "" {
						token = t
					} else if t, ok := jsonBody["authToken"].(string); ok && t != "" {
						token = t
					} else if t, ok := jsonBody["session"].(string); ok && t != "" {
						token = t
					} else if data, ok := jsonBody["data"].(map[string]interface{}); ok {
						if t, ok := data["token"].(string); ok && t != "" {
							token = t
						}
					} else if result, ok := jsonBody["result"].(map[string]interface{}); ok {
						if t, ok := result["token"].(string); ok && t != "" {
							token = t
						}
					}
				}
			}
		} else {
			resp.Body.Close()
		}

		// Check if authentication was successful
		if (statusCode == 200 || statusCode == 201) && token != "" && token != "null" {
			c.logger.Info("Authentication successful")
			c.token = token
			if c.logger.IsDebug() {
				// Show only first 20 characters of token in debug mode
				tokenPreview := token
				if len(token) > 20 {
					tokenPreview = token[:20] + "..."
				}
				c.logger.Debug("Access token obtained: %s", tokenPreview)
			}
			return nil
		}

		// Log failure details for retryable errors
		c.logger.Warning("Attempt %d failed (http_code=%d, token=%s). Retrying...",
			attempt, statusCode, map[bool]string{true: "found", false: "not found"}[token != ""])

		// Wait before retry
		delay := c.config.BaseDelay + time.Duration(attempt)*time.Second + time.Duration(rand.Intn(2))*time.Second
		time.Sleep(delay)
	}

	return errors.NewAuthenticationError(
		fmt.Sprintf("Failed to obtain token after %d attempts", c.config.MaxRetries), 0)
}

// GetToken returns the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// Made with help from Bob
