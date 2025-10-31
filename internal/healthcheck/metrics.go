package healthcheck

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// MetricsChecker collects performance metrics from services
type MetricsChecker struct {
	CheckName string
	Endpoints []MetricEndpoint
}

// MetricEndpoint represents a service endpoint to measure
type MetricEndpoint struct {
	Name string
	URL  string
}

// PerformanceMetric represents timing data for an endpoint
type PerformanceMetric struct {
	Endpoint     string `json:"endpoint"`
	ResponseTime int64  `json:"response_time_ms"`
	StatusCode   int    `json:"status_code"`
	Status       string `json:"status"`
}

// Name returns the checker name
func (c *MetricsChecker) Name() string {
	return c.CheckName
}

// Check collects performance metrics from all endpoints
func (c *MetricsChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	metrics := []PerformanceMetric{}
	totalTime := int64(0)
	slowEndpoints := []string{}
	fastEndpoints := []string{}

	// Performance thresholds
	const (
		fastThreshold = 100  // < 100ms is fast
		slowThreshold = 1000 // > 1s is slow
	)

	for _, endpoint := range c.Endpoints {
		metric := c.measureEndpoint(endpoint)
		metrics = append(metrics, metric)
		totalTime += metric.ResponseTime

		if metric.ResponseTime > slowThreshold {
			slowEndpoints = append(slowEndpoints, fmt.Sprintf("%s (%dms)", endpoint.Name, metric.ResponseTime))
		} else if metric.ResponseTime < fastThreshold {
			fastEndpoints = append(fastEndpoints, fmt.Sprintf("%s (%dms)", endpoint.Name, metric.ResponseTime))
		}
	}

	avgTime := int64(0)
	if len(metrics) > 0 {
		avgTime = totalTime / int64(len(metrics))
	}

	result.Details["metrics"] = metrics
	result.Details["average_response_time_ms"] = avgTime
	result.Details["total_endpoints"] = len(c.Endpoints)
	result.Details["slow_endpoints"] = slowEndpoints
	result.Details["fast_endpoints"] = fastEndpoints

	// Determine status based on performance
	if len(slowEndpoints) > len(c.Endpoints)/2 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("Performance degraded: %d/%d endpoints slow (>1s)", len(slowEndpoints), len(c.Endpoints))
	} else if len(slowEndpoints) > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("Some slow endpoints detected: avg %dms", avgTime)
	} else {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("Good performance: avg %dms across %d endpoints", avgTime, len(c.Endpoints))
	}

	result.Duration = time.Since(start)
	return result
}

// measureEndpoint measures response time for a single endpoint
func (c *MetricsChecker) measureEndpoint(endpoint MetricEndpoint) PerformanceMetric {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	metric := PerformanceMetric{
		Endpoint: endpoint.Name,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.URL, http.NoBody)
	if err != nil {
		metric.ResponseTime = 0
		metric.StatusCode = 0
		metric.Status = "error"
		return metric
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		metric.ResponseTime = elapsed
		metric.StatusCode = 0
		metric.Status = "timeout"
		return metric
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("warning: failed to close response body: %v", err)
		}
	}()

	metric.ResponseTime = elapsed
	metric.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		metric.Status = "ok"
	} else if resp.StatusCode >= 500 {
		metric.Status = "error"
	} else {
		metric.Status = "warn"
	}

	return metric
}
