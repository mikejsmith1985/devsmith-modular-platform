// Package internal_logs_handlers provides HTTP handlers for logs operations.
package internal_logs_handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/cache"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// CachedDashboardHandler provides caching layer for dashboard endpoints.
type CachedDashboardHandler struct { //nolint:govet // struct alignment optimized for readability
	dashboardService logs_services.DashboardServiceInterface
	cache            *cache.DashboardCache
	logger           *logrus.Logger
}

// NewCachedDashboardHandler creates a new cached dashboard handler.
func NewCachedDashboardHandler(
	dashboardService logs_services.DashboardServiceInterface,
	c *cache.DashboardCache,
	logger *logrus.Logger,
) *CachedDashboardHandler {
	return &CachedDashboardHandler{
		dashboardService: dashboardService,
		cache:            c,
		logger:           logger,
	}
}

// GetDashboardStats returns dashboard statistics with caching.
func (h *CachedDashboardHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Try to get from cache first
	cached, err := h.cache.GetDashboardStats(ctx)
	if err == nil && cached != nil {
		h.logger.Debug("Dashboard stats served from cache")
		c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    cached,
		})
		return
	}

	// Cache miss - get fresh stats
	stats, err := h.dashboardService.GetDashboardStats(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard stats")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve dashboard statistics",
		})
		return
	}

	// Store in cache for future requests
	cacheErr := h.cache.SetDashboardStats(context.Background(), stats)
	if cacheErr != nil {
		h.logger.WithError(cacheErr).Warn("Failed to cache dashboard stats")
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// GetServiceStats returns service statistics with caching.
func (h *CachedDashboardHandler) GetServiceStats(c *gin.Context) {
	ctx := c.Request.Context()

	service := c.Query("service")
	if service == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "service parameter is required",
		})
		return
	}

	// Try to get from cache first
	cached, err := h.cache.GetServiceStats(ctx, service)
	if err == nil && cached != nil {
		h.logger.Debugf("Service stats for %s served from cache", service)
		c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    cached,
		})
		return
	}

	timeRangeStr := c.DefaultQuery("timeRange", "1h")
	timeRange, err := parseDuration(timeRangeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid timeRange format",
		})
		return
	}

	stats, err := h.dashboardService.GetServiceStats(ctx, service, timeRange)
	if err != nil {
		h.logger.WithError(err).Warnf("Failed to get stats for service %s", service)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to retrieve service statistics",
		})
		return
	}

	// Store in cache
	cacheErr := h.cache.SetServiceStats(context.Background(), service, stats)
	if cacheErr != nil {
		h.logger.WithError(cacheErr).Warn("Failed to cache service stats")
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stats,
	})
}

// InvalidateDashboardCache clears the dashboard cache.
func (h *CachedDashboardHandler) InvalidateDashboardCache(c *gin.Context) {
	err := h.cache.Clear(context.Background())
	if err != nil {
		h.logger.WithError(err).Error("Failed to clear cache")
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to invalidate cache",
		})
		return
	}

	h.logger.Info("Dashboard cache invalidated")
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Cache invalidated successfully",
	})
}
