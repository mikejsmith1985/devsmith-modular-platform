package internal_analytics_handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/metrics"
)

// MetricsDashboardHandler handles metrics dashboard requests
type MetricsDashboardHandler struct {
	analyzer *metrics.Analyzer
}

// NewMetricsDashboardHandler creates a new metrics dashboard handler
func NewMetricsDashboardHandler() *MetricsDashboardHandler {
	return &MetricsDashboardHandler{
		analyzer: metrics.NewAnalyzer(),
	}
}

// GetDashboardData returns metrics summary for dashboard
func (h *MetricsDashboardHandler) GetDashboardData(c *gin.Context) {
	// Parse time range from query params (default: last 7 days)
	daysBack := 7
	if days := c.Query("days"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 {
			daysBack = parsed
		}
	}

	end := time.Now()
	start := end.AddDate(0, 0, -daysBack)

	summary, err := h.analyzer.Analyze(metrics.TimeRange{
		Start: start,
		End:   end,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to analyze metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
		"time_range": gin.H{
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
			"days":  daysBack,
		},
	})
}

// GetTrends returns trend data for specific metrics
func (h *MetricsDashboardHandler) GetTrends(c *gin.Context) {
	metricType := c.Query("metric_type")
	if metricType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing 'metric_type' parameter",
		})
		return
	}

	daysBack := 30
	if days := c.Query("days"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 {
			daysBack = parsed
		}
	}

	end := time.Now()
	start := end.AddDate(0, 0, -daysBack)

	summary, err := h.analyzer.Analyze(metrics.TimeRange{
		Start: start,
		End:   end,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to analyze trends",
			"details": err.Error(),
		})
		return
	}

	trend, exists := summary.Trends[metricType]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Trend not found for metric type",
			"type":    metricType,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    trend,
	})
}

// GetViolations returns rule violation history
func (h *MetricsDashboardHandler) GetViolations(c *gin.Context) {
	daysBack := 30
	if days := c.Query("days"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 {
			daysBack = parsed
		}
	}

	end := time.Now()
	start := end.AddDate(0, 0, -daysBack)

	summary, err := h.analyzer.Analyze(metrics.TimeRange{
		Start: start,
		End:   end,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to analyze violations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       summary.TopViolations,
		"total":      summary.RuleViolations,
		"violations": summary.TopViolations, // Compatibility with frontend expecting "violations" key
	})
}
