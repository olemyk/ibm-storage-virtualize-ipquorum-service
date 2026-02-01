package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/olemyk/ipquorum-go/internal/errors"
)

// DownloadPayload represents the download API payload
type DownloadPayload struct {
	Prefix   string `json:"prefix"`
	Filename string `json:"filename"`
}

// DownloadJAR downloads the ip_quorum.jar file
func (c *Client) DownloadJAR() error {
	if c.token == "" {
		return errors.NewAPIError("not authenticated. Call Authenticate() first", 0, "")
	}

	c.logger.Info("Proceeding to download %s...", c.config.OutputFile)

	url := fmt.Sprintf("%s/rest/v1/download", c.baseURL)

	// Build payload
	payload := DownloadPayload{
		Prefix:   "/dumps",
		Filename: "ip_quorum.jar",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to marshal payload: %v", err), 0, url)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to create request: %v", err), 0, url)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Auth-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	// Execute request with longer timeout for download
	client := &http.Client{
		Transport: c.httpClient.Transport,
		Timeout:   300 * time.Second, // 5 minutes for download
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("request failed: %v", err), 0, url)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	c.logger.Info("Download response status: %d", statusCode)

	// Log response headers in debug mode
	if c.logger.IsDebug() {
		c.logger.Debug("Download Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				c.logger.Debug("  %s: %s", key, value)
			}
		}
	}

	// Check status code
	if statusCode != 200 && statusCode != 201 {
		errorMsg := fmt.Sprintf("Download failed with status %d", statusCode)
		c.logger.Error(errorMsg)
		return errors.NewAPIError(errorMsg, statusCode, url)
	}

	// Create output file
	outFile, err := os.Create(c.config.OutputFile)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to create output file: %v", err), 0, url)
	}
	defer outFile.Close()

	// Copy response body to file
	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to write file: %v", err), 0, url)
	}

	// Verify file was created and has content
	if written > 0 {
		c.logger.Info("Downloaded %s (size: %d bytes)", c.config.OutputFile, written)
	} else {
		c.logger.Warning("Success status but empty file. Check server response.")
	}

	return nil
}

// Made with help from Bob
