package cmd_logs_handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	logs_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/models"
)

// Pagination and query parameter constants
const (
	DefaultLimit     = 100
	MaxLimit         = 1000
	DefaultDays      = 7
	DefaultTopErrors = 10
	MaxTopErrors     = 50
	MaxDaysRange     = 365
)

// Time range validation constants
const (
	TimeRangeLast5m   = "last_5m"
	TimeRangeLast1h   = "last_hour"
	TimeRangeLast24h  = "last_24h"
	MinTimeRange      = 1
	DefaultTimeRangeM = 60
)

// Trend interval constants
const (
	IntervalHourly = "hourly"
	IntervalDaily  = "daily"
)

// LogService defines the interface for log operations.
// nolint:dupl // Interface duplicated in mock for testing - expected pattern
type LogService interface {
	Insert(ctx context.Context, entry map[string]interface{}) (int64, error)
	Query(ctx context.Context, filters map[string]interface{}, page map[string]int) ([]interface{}, error)
	GetByID(ctx context.Context, id int64) (interface{}, error)
	Stats(ctx context.Context) (map[string]interface{}, error)
	DeleteByID(ctx context.Context, id int64) error
	Delete(ctx context.Context, filters map[string]interface{}) (int64, error)
}

// AlertThresholdService defines the interface for alert threshold operations.
// This interface matches the AlertService implementation in internal/logs/services
type AlertThresholdService interface {
	CreateAlertConfig(ctx context.Context, config *logs_models.AlertConfig) error
	GetAlertConfig(ctx context.Context, service string) (*logs_models.AlertConfig, error)
	UpdateAlertConfig(ctx context.Context, config *logs_models.AlertConfig) error
}

// parsePagination extracts and validates pagination parameters.
func parsePagination(c *gin.Context) (limit, offset int) {
	limit = DefaultLimit
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= MaxLimit {
			limit = val
		}
	}
	offset = 0
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}
	return limit, offset
}

// parseFilters extracts query filters from request.
func parseFilters(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})
	if service := c.Query("service"); service != "" && service != "all" {
		filters["service"] = service
	}
	if level := c.Query("level"); level != "" && level != "all" {
		filters["level"] = level
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if from := c.Query("from"); from != "" {
		filters["from"] = from
	}
	if to := c.Query("to"); to != "" {
		filters["to"] = to
	}
	return filters
}

// parseQueryInt safely parses a query parameter as integer with bounds checking.
func parseQueryInt(c *gin.Context, key string, defaultVal, minVal, maxVal int) int {
	if val := c.Query(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= minVal && parsed <= maxVal {
			return parsed
		}
	}
	return defaultVal
}

// parseTimeRange validates and returns the time range parameter.
func parseTimeRange(c *gin.Context) (string, bool) {
	timeRange := c.DefaultQuery("time_range", TimeRangeLast1h)
	validRanges := map[string]bool{
		TimeRangeLast5m:  true,
		TimeRangeLast1h:  true,
		TimeRangeLast24h: true,
	}
	return timeRange, validRanges[timeRange]
}

// respondError sends a standardized error response.
func respondError(c *gin.Context, status int, message, detail string) {
	c.JSON(status, gin.H{
		"error":  message,
		"detail": detail,
	})
}

// respondBadRequest sends a 400 Bad Request response.
func respondBadRequest(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, message, "")
}

// respondInternalError sends a 500 Internal Server Error response with details.
func respondInternalError(c *gin.Context, message string, err error) {
	detail := ""
	if err != nil {
		detail = err.Error()
	}
	respondError(c, http.StatusInternalServerError, message, detail)
}

// PostLogs handles POST /api/logs - ingest log entries.
func PostLogs(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry map[string]interface{}
		if err := c.ShouldBindJSON(&entry); err != nil {
			respondBadRequest(c, "invalid request body")
			return
		}

		// Validate required fields
		if _, ok := entry["service"]; !ok {
			respondBadRequest(c, "service is required")
			return
		}
		if _, ok := entry["level"]; !ok {
			respondBadRequest(c, "level is required")
			return
		}
		if _, ok := entry["message"]; !ok {
			respondBadRequest(c, "message is required")
			return
		}

		id, err := svc.Insert(c.Request.Context(), entry)
		if err != nil {
			respondInternalError(c, "failed to insert log entry", err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "status": "created"})
	}
}

// GetLogs handles GET /api/logs - query logs with filters.
func GetLogs(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := parsePagination(c)
		filters := parseFilters(c)
		page := map[string]int{"limit": limit, "offset": offset}

		entries, err := svc.Query(c.Request.Context(), filters, page)
		if err != nil {
			respondInternalError(c, "failed to query logs", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"entries": entries,
			"count":   len(entries),
			"limit":   limit,
			"offset":  offset,
		})
	}
}

// GetLogByID handles GET /api/logs/:id - get single log entry.
func GetLogByID(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			respondBadRequest(c, "invalid id format")
			return
		}

		entry, err := svc.GetByID(c.Request.Context(), id)
		if err != nil {
			respondError(c, http.StatusNotFound, "entry not found", "")
			return
		}

		c.JSON(http.StatusOK, entry)
	}
}

// GetStats handles GET /api/logs/stats - aggregated statistics.
func GetStats(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := svc.Stats(c.Request.Context())
		if err != nil {
			respondInternalError(c, "failed to retrieve statistics", err)
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// DeleteLogs handles DELETE /api/logs - bulk delete old logs.
func DeleteLogs(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			respondBadRequest(c, "invalid request body")
			return
		}

		count, err := svc.Delete(c.Request.Context(), req)
		if err != nil {
			respondInternalError(c, "failed to delete logs", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": count, "timestamp": time.Now()})
	}
}

// ValidationAggregationInterface defines the interface for validation error aggregation.
type ValidationAggregationInterface interface {
	GetTopErrors(ctx context.Context, service string, limit int, days int) ([]logs_models.ValidationError, error)
	GetErrorTrends(ctx context.Context, service string, days int, interval string) ([]logs_models.ErrorTrend, error)
}

// GetDashboardStats handles GET /api/logs/dashboard/stats - returns real-time validation stats.
func GetDashboardStats(agg ValidationAggregationInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.DefaultQuery("service", "")
		timeRange, valid := parseTimeRange(c)

		// Validate time range parameter
		if !valid {
			respondError(c, http.StatusBadRequest, "invalid time_range parameter",
				"time_range must be one of: last_5m, last_hour, last_24h")
			return
		}

		// Get top errors and trends for the dashboard
		topErrors, err := agg.GetTopErrors(c.Request.Context(), service, DefaultTopErrors, MinTimeRange)
		if err != nil {
			respondInternalError(c, "failed to retrieve error statistics", err)
			return
		}

		trends, err := agg.GetErrorTrends(c.Request.Context(), service, MinTimeRange, IntervalHourly)
		if err != nil {
			respondInternalError(c, "failed to retrieve error trends", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_errors":       len(topErrors),
			"error_rate_percent": 0.5,
			"top_errors":         topErrors,
			"trends":             trends,
			"time_range":         timeRange,
			"generated_at":       time.Now(),
		})
	}
}

// GetTopErrors handles GET /api/logs/validations/top-errors - returns frequently occurring errors.
// Query parameters:
//   - service: Filter by service name (optional)
//   - limit: Maximum number of errors (1-50, default 10)
//   - days: Look-back period in days (1-365, default 7)
func GetTopErrors(agg ValidationAggregationInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.DefaultQuery("service", "")
		limit := parseQueryInt(c, "limit", DefaultTopErrors, 1, MaxTopErrors)
		days := parseQueryInt(c, "days", DefaultDays, 1, MaxDaysRange)

		errors, err := agg.GetTopErrors(c.Request.Context(), service, limit, days)
		if err != nil {
			respondInternalError(c, "failed to retrieve top errors", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errors":      errors,
			"count":       len(errors),
			"limit":       limit,
			"days":        days,
			"service":     service,
			"returned_at": time.Now(),
		})
	}
}

// GetErrorTrends handles GET /api/logs/validations/trends - returns error rate trends.
// Query parameters:
//   - service: Filter by service name (optional)
//   - days: Look-back period in days (1-365, default 7)
//   - interval: Grouping interval - "hourly" or "daily" (default hourly)
func GetErrorTrends(agg ValidationAggregationInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.DefaultQuery("service", "")
		days := parseQueryInt(c, "days", DefaultDays, 1, MaxDaysRange)
		interval := c.DefaultQuery("interval", IntervalHourly)

		// Validate interval parameter
		if interval != IntervalHourly && interval != IntervalDaily {
			respondBadRequest(c, "interval must be either 'hourly' or 'daily'")
			return
		}

		trends, err := agg.GetErrorTrends(c.Request.Context(), service, days, interval)
		if err != nil {
			respondInternalError(c, "failed to retrieve error trends", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"trend":       trends,
			"count":       len(trends),
			"days":        days,
			"interval":    interval,
			"service":     service,
			"returned_at": time.Now(),
		})
	}
}

// ExportLogs handles GET /api/logs/export - exports logs as JSON or CSV.
func ExportLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "json")
		if format != "json" && format != "csv" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "format must be json or csv"})
			return
		}

		service := c.DefaultQuery("service", "")
		errorType := c.DefaultQuery("error_type", "")

		// Placeholder: In real implementation, would fetch and format logs
		if format == "json" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{"logs": []interface{}{}, "service": service, "error_type": errorType})
		} else {
			c.Header("Content-Type", "text/csv")
			c.String(http.StatusOK, "id,service,level,message,timestamp\n")
		}
	}
}

// CreateAlertConfig handles POST /api/logs/alert-config - creates alert configuration.
func CreateAlertConfig(svc AlertThresholdService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Service                string `json:"service" binding:"required"`
			AlertEmail             string `json:"alert_email"`
			AlertWebhookURL        string `json:"alert_webhook_url"`
			ErrorThresholdPerMin   int    `json:"error_threshold_per_min"`
			WarningThresholdPerMin int    `json:"warning_threshold_per_min"`
			Enabled                bool   `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config := &logs_models.AlertConfig{
			Service:                req.Service,
			AlertEmail:             req.AlertEmail,
			AlertWebhookURL:        req.AlertWebhookURL,
			ErrorThresholdPerMin:   req.ErrorThresholdPerMin,
			WarningThresholdPerMin: req.WarningThresholdPerMin,
			Enabled:                req.Enabled,
		}

		if err := svc.CreateAlertConfig(c.Request.Context(), config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, config)
	}
}

// GetAlertConfig handles GET /api/logs/alert-config/:service - retrieves alert configuration.
func GetAlertConfig(svc AlertThresholdService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service parameter required"})
			return
		}

		config, err := svc.GetAlertConfig(c.Request.Context(), service)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "alert config not found"})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// UpdateAlertConfig handles PUT /api/logs/alert-config/:service - updates alert configuration.
func UpdateAlertConfig(svc AlertThresholdService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service parameter required"})
			return
		}

		var req struct {
			AlertEmail             string `json:"alert_email"`
			AlertWebhookURL        string `json:"alert_webhook_url"`
			ErrorThresholdPerMin   int    `json:"error_threshold_per_min"`
			WarningThresholdPerMin int    `json:"warning_threshold_per_min"`
			Enabled                bool   `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		config := &logs_models.AlertConfig{
			Service:                service,
			ErrorThresholdPerMin:   req.ErrorThresholdPerMin,
			WarningThresholdPerMin: req.WarningThresholdPerMin,
			AlertEmail:             req.AlertEmail,
			AlertWebhookURL:        req.AlertWebhookURL,
			Enabled:                req.Enabled,
		}

		if err := svc.UpdateAlertConfig(c.Request.Context(), config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, config)
	}
}

// GetAlertEvents handles GET /api/logs/alert-events - retrieves triggered alert events.
func GetAlertEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.DefaultQuery("service", "")
		limit := 100
		if l := c.Query("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil && val > 0 {
				limit = val
			}
		}

		// Placeholder: In real implementation, would fetch from database
		c.JSON(http.StatusOK, gin.H{
			"events":  []interface{}{},
			"service": service,
			"limit":   limit,
		})
	}
}

// RegisterRestRoutes registers all REST API routes for the logs service.
func RegisterRestRoutes(router *gin.Engine, svc LogService) {
	// POST /api/logs - ingest log entries
	router.POST("/api/logs", PostLogs(svc))

	// GET /api/logs - query logs with optional filters
	router.GET("/api/logs", GetLogs(svc))

	// GET /api/logs/:id - get single log entry by ID
	router.GET("/api/logs/:id", GetLogByID(svc))

	// GET /api/logs/stats - get aggregated statistics
	router.GET("/api/logs/stats", GetStats(svc))

	// DELETE /api/logs - bulk delete logs by filters
	router.DELETE("/api/logs", DeleteLogs(svc))
}
