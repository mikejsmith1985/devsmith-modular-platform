// Package logging provides a client for sending logs to the DevSmith Logging service.
package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Client sends logs to the logging service via HTTP.
type Client struct {
	httpClient *http.Client
	endpoint   string
}

// NewClient creates a new logging client that posts to the provided endpoint.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Post sends a JSON payload to the logs service. payload will be marshaled to JSON.
func (c *Client) Post(ctx context.Context, data map[string]interface{}) error {
	if c == nil {
		return fmt.Errorf("logging client is nil")
	}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("warning: failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("logs service returned status %d", resp.StatusCode)
	}
	return nil
}
