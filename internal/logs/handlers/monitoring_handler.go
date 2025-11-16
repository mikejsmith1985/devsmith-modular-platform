// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/monitoring"
)

// MonitoringHandler handles monitoring dashboard API requests
type MonitoringHandler struct {
	collector monitoring.MetricsCollector
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(collector monitoring.MetricsCollector) *MonitoringHandler {
	return &MonitoringHandler{
		collector: collector,
	}
}

// MetricsResponse represents the time-series metrics response
type MetricsResponse struct {
	TimeRange     string            `json:"time_range"`
	ErrorRate     float64           `json:"error_rate"` // errors per minute
	ResponseTimes ResponseTimeStats `json:"response_times"`
	RequestCount  int64             `json:"request_count"`
	ErrorCount    int64             `json:"error_count"`
	DataPoints    []MetricDataPoint `json:"data_points"`
}

// ResponseTimeStats contains response time percentiles
type ResponseTimeStats struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
}

// MetricDataPoint represents a single time-series data point
type MetricDataPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorRate    float64   `json:"error_rate"`
	ResponseTime float64   `json:"response_time"`
	RequestCount int       `json:"request_count"`
}

// AlertResponse represents an active alert
type AlertResponse struct {
	ID          int64                  `json:"id"`
	AlertType   string                 `json:"alert_type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	ServiceName string                 `json:"service_name"`
	Triggered   time.Time              `json:"triggered"`
	Resolved    *time.Time             `json:"resolved,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StatsResponse provides summary statistics
type StatsResponse struct {
	ActiveAlerts    int               `json:"active_alerts"`
	ServicesUp      int               `json:"services_up"`
	ServicesDown    int               `json:"services_down"`
	ErrorRate       float64           `json:"error_rate"`        // last hour
	AvgResponseTime float64           `json:"avg_response_time"` // milliseconds
	ServiceHealth   map[string]string `json:"service_health"`    // service -> "healthy"/"degraded"/"down"
}

// GetMetrics returns time-series metrics data
// GET /api/logs/monitoring/metrics?window=1h&interval=1m
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	// Parse query parameters
	windowStr := c.DefaultQuery("window", "1h")
	// intervalStr := c.DefaultQuery("interval", "1m") // TODO: Use for time-series data points

	window, err := time.ParseDuration(windowStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid window parameter"})
		return
	}

	// TODO: Parse interval when implementing time-series data points
	// interval, err := time.ParseDuration(intervalStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interval parameter"})
	// 	return
	// }

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get error rate for the window
	errorRate, err := h.collector.GetErrorRate(ctx, window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get error rate"})
		return
	}

	// Get response times
	responseTimes, err := h.collector.GetResponseTimes(ctx, window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response times"})
		return
	}

	// Calculate percentiles
	stats := calculateResponseTimeStats(responseTimes)

	// TODO: Get time-series data points (requires additional query)
	// For now, return aggregated data
	dataPoints := []MetricDataPoint{}

	response := MetricsResponse{
		TimeRange:     windowStr,
		ErrorRate:     errorRate,
		ResponseTimes: stats,
		RequestCount:  int64(len(responseTimes)), // Approximate
		ErrorCount:    int64(errorRate * window.Minutes()),
		DataPoints:    dataPoints,
	}

	c.JSON(http.StatusOK, response)
}

// GetAlerts returns active alerts
// GET /api/logs/monitoring/alerts?active=true
func (h *MonitoringHandler) GetAlerts(c *gin.Context) {
	activeOnly := c.DefaultQuery("active", "true") == "true"

	// TODO: Implement alert retrieval from database
	// For now, return empty array
	alerts := []AlertResponse{}

	if activeOnly {
		// Filter to only unresolved alerts
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetStats returns summary statistics
// GET /api/logs/monitoring/stats
func (h *MonitoringHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get error rate for last hour
	errorRate, err := h.collector.GetErrorRate(ctx, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get error rate"})
		return
	}

	// Get response times for last hour
	responseTimes, err := h.collector.GetResponseTimes(ctx, time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response times"})
		return
	}

	// Calculate average response time
	avgResponseTime := 0.0
	if len(responseTimes) > 0 {
		sum := 0.0
		for _, rt := range responseTimes {
			sum += rt
		}
		avgResponseTime = sum / float64(len(responseTimes))
	}

	// TODO: Query service health from health check table
	serviceHealth := map[string]string{
		"portal":    "healthy",
		"review":    "healthy",
		"logs":      "healthy",
		"analytics": "healthy",
	}

	servicesUp := 0
	servicesDown := 0
	for _, health := range serviceHealth {
		if health == "healthy" {
			servicesUp++
		} else {
			servicesDown++
		}
	}

	response := StatsResponse{
		ActiveAlerts:    0, // TODO: Query from alerts table
		ServicesUp:      servicesUp,
		ServicesDown:    servicesDown,
		ErrorRate:       errorRate,
		AvgResponseTime: avgResponseTime,
		ServiceHealth:   serviceHealth,
	}

	c.JSON(http.StatusOK, response)
}

// calculateResponseTimeStats computes percentiles and statistics
func calculateResponseTimeStats(times []float64) ResponseTimeStats {
	if len(times) == 0 {
		return ResponseTimeStats{}
	}

	// Sort times for percentile calculation
	sorted := make([]float64, len(times))
	copy(sorted, times)

	// Simple insertion sort (good enough for ~1000 items)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	// Calculate percentiles
	p50 := sorted[len(sorted)*50/100]
	p95 := sorted[len(sorted)*95/100]
	p99 := sorted[len(sorted)*99/100]
	max := sorted[len(sorted)-1]

	// Calculate average
	sum := 0.0
	for _, t := range sorted {
		sum += t
	}
	avg := sum / float64(len(sorted))

	return ResponseTimeStats{
		P50: p50,
		P95: p95,
		P99: p99,
		Avg: avg,
		Max: max,
	}
}
