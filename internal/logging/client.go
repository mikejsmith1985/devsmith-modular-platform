package logging

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// Client is a minimal HTTP client for sending logs to the Logs service.
type Client struct {
    endpoint string
    client   *http.Client
}

// NewClient creates a new logging client that posts to the provided endpoint.
func NewClient(endpoint string) *Client {
    return &Client{
        endpoint: endpoint,
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

// Post sends a JSON payload to the logs service. payload will be marshaled to JSON.
func (c *Client) Post(ctx context.Context, payload interface{}) error {
    if c == nil {
        return fmt.Errorf("logging client is nil")
    }
    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("post to logs service: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        return fmt.Errorf("logs service returned status %d", resp.StatusCode)
    }
    return nil
}
