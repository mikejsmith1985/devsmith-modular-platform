// Package handlers provides HTTP handlers for the Logs service API.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// GetCorrelatedLogs handles GET /api/logs/correlation/:correlationId - retrieve logs for correlation
func GetCorrelatedLogs(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param("correlationId")
		if correlationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "correlation_id is required"})
			return
		}

		// Parse pagination
		limit := 50
		if l := c.Query("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 1000 {
				limit = val
			}
		}

		offset := 0
		if o := c.Query("offset"); o != "" {
			if val, err := strconv.Atoi(o); err == nil && val >= 0 {
				offset = val
			}
		}

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

// GetCorrelationMetadata handles GET /api/logs/correlation/:correlationId/metadata
func GetCorrelationMetadata(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param("correlationId")
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

// GetTraceTimeline handles GET /api/logs/correlation/:correlationId/timeline
func GetTraceTimeline(svc *services.ContextService) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.Param("correlationId")
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
