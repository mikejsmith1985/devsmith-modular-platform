// Package handlers provides HTTP handlers for the Logs service API.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// HTTP handler constants for correlation endpoints
const (
	// ParamCorrelationID is the URL parameter name for correlation ID
	ParamCorrelationID = "correlationId"

	// QueryParamLimit is the query parameter name for pagination limit
	QueryParamLimit = "limit"

	// QueryParamOffset is the query parameter name for pagination offset
	QueryParamOffset = "offset"
)

// GetCorrelatedLogs handles GET /api/logs/correlation/:correlationId - retrieve logs for correlation.
//
// Returns all logs associated with a correlation ID with pagination support.
//
// Query Parameters:
// - limit: Pagination limit (default 100, max 1000)
// - offset: Number of results to skip (default 0)
//
// Response (200 OK):
//
//	{
//	  "correlation_id": "abc123def456...",
//	  "logs": [
//	    {
//	      "id": 1,
//	      "timestamp": "2025-01-01T12:00:00Z",
//	      "level": "info",
//	      "message": "Request started",
//	      "service": "portal",
//	      "context": { ... }
//	    },
//	    ...
//	  ],
//	  "count": 10,
//	  "limit": 100,
//	  "offset": 0
//	}
//
// Response (400 Bad Request): Missing correlation ID
// Response (500 Internal Server Error): Database query error
func GetCorrelatedLogs(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param(ParamCorrelationID)
		if correlationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "correlation_id is required"})
			return
		}

		// Parse pagination (uses shared pagination parser from rest_handler)
		limit, offset := parsePagination(c)

		logs, err := svc.GetCorrelatedLogs(c.Request.Context(), correlationID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"correlation_id": correlationID,
			"logs":           logs,
			"count":          len(logs),
			"limit":          limit,
			"offset":         offset,
		})
	}
}

// GetCorrelationMetadata handles GET /api/logs/correlation/:correlationId/metadata.
//
// Returns aggregated metadata about all logs in a correlation.
//
// Response (200 OK):
//
//	{
//	  "correlation_id": "abc123def456...",
//	  "total_logs": 42,
//	  "services": ["portal", "analytics", "review"],
//	  "trace_ids": ["trace-xyz123"]
//	}
//
// Response (400 Bad Request): Missing correlation ID
// Response (500 Internal Server Error): Database query error
func GetCorrelationMetadata(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param(ParamCorrelationID)
		if correlationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "correlation_id is required"})
			return
		}

		metadata, err := svc.GetCorrelationMetadata(c.Request.Context(), correlationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, metadata)
	}
}

// GetTraceTimeline handles GET /api/logs/correlation/:correlationId/timeline.
//
// Returns a chronological timeline of all events in a correlation,
// suitable for visualizing request flow through distributed services.
//
// Response (200 OK):
//
//	{
//	  "correlation_id": "abc123def456...",
//	  "timeline": [
//	    {
//	      "timestamp": "2025-01-01T12:00:00Z",
//	      "level": "info",
//	      "service": "portal",
//	      "message": "Request started",
//	      "trace_id": "trace-xyz",
//	      "span_id": "span-1"
//	    },
//	    {
//	      "timestamp": "2025-01-01T12:00:01Z",
//	      "level": "info",
//	      "service": "analytics",
//	      "message": "Processing event",
//	      "trace_id": "trace-xyz",
//	      "span_id": "span-2"
//	    },
//	    ...
//	  ],
//	  "count": 42
//	}
//
// Response (400 Bad Request): Missing correlation ID
// Response (500 Internal Server Error): Database query error
func GetTraceTimeline(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param(ParamCorrelationID)
		if correlationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "correlation_id is required"})
			return
		}

		timeline, err := svc.GetTraceTimeline(c.Request.Context(), correlationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"correlation_id": correlationID,
			"timeline":       timeline,
			"count":          len(timeline),
		})
	}
}
