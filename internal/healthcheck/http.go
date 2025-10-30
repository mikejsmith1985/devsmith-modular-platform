package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPChecker validates HTTP endpoints are responding
type HTTPChecker struct {
	CheckName string
	URL       string
}

// Name returns the checker name
func (c *HTTPChecker) Name() string {
	return c.CheckName
}

// Check validates the HTTP endpoint
func (c *HTTPChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.URL, nil)
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to create request"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("HTTP request failed: %s", c.URL)
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log but don't fail - response already processed
		}
	}()

	result.Details["url"] = c.URL
	result.Details["status_code"] = resp.StatusCode
	result.Details["response_time_ms"] = time.Since(start).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("HTTP %d OK", resp.StatusCode)
	} else if resp.StatusCode >= 500 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("HTTP %d Server Error", resp.StatusCode)
	} else {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	result.Duration = time.Since(start)
	return result
}
