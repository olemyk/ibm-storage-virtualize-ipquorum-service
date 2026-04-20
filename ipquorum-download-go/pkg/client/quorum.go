package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olemyk/ipquorum-go/internal/errors"
)

// QuorumAppPayload represents the mkquorumapp API payload
type QuorumAppPayload struct {
	IP6           bool   `json:"ip_6"`
	NoMetadata    bool   `json:"nometadata"`
	PartnerSystem string `json:"partnersystem"`
	PartnerIP6    bool   `json:"partnerip6"`
}

// CreateQuorumApp creates an IP Quorum application via mkquorumapp API
func (c *Client) CreateQuorumApp() error {
	if c.token == "" {
		return errors.NewAPIError("not authenticated. Call Authenticate() first", 0, "")
	}

	c.logger.Info("Creating IP-Quorum app via /rest/v1/mkquorumapp...")

	url := fmt.Sprintf("%s/rest/v1/mkquorumapp", c.baseURL)

	// Build payload
	payload := QuorumAppPayload{
		IP6:           c.config.IP6,
		NoMetadata:    c.config.NoMetadata,
		PartnerSystem: c.config.PartnerSystem,
		PartnerIP6:    c.config.PartnerIP6,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to marshal payload: %v", err), 0, url)
	}

	c.logger.Debug("Payload: %s", string(jsonPayload))

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to create request: %v", err), 0, url)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("request failed: %v", err), 0, url)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	c.logger.Info("mkquorumapp response status: %d", statusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Warning("Failed to read response body: %v", err)
	}

	// Log response in debug mode
	if c.logger.IsDebug() {
		// Try to pretty-print JSON
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonBody, "", "  ")
			c.logger.Debug("Response: %s", string(prettyJSON))
		} else {
			c.logger.Debug("Response text: %s", string(body))
		}
	}

	// Check status code
	if statusCode == 200 || statusCode == 201 {
		c.logger.Info("mkquorumapp call completed successfully")
		return nil
	}

	errorMsg := fmt.Sprintf("mkquorumapp failed with status %d", statusCode)
	c.logger.Error(errorMsg)
	return errors.NewAPIError(errorMsg, statusCode, url)
}

//
