// Package handlers provides HTTP handlers for the Logs service API.
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

// parsePagination extracts and validates pagination parameters.
func parsePagination(c *gin.Context) (limit, offset int) {
	limit = 100
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
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
	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}
	if level := c.Query("level"); level != "" {
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

// PostLogs handles POST /api/logs - ingest log entries.
func PostLogs(svc LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry map[string]interface{}
		if err := c.ShouldBindJSON(&entry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Validate required fields
		if _, ok := entry["service"]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service is required"})
			return
		}
		if _, ok := entry["level"]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "level is required"})
			return
		}
		if _, ok := entry["message"]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
			return
		}

		id, err := svc.Insert(c.Request.Context(), entry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		entry, err := svc.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		count, err := svc.Delete(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": count, "timestamp": time.Now()})
	}
}
