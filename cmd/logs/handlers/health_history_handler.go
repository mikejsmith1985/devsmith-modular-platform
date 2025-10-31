package cmd_logs_handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logs_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
)

// parseLimit extracts and validates limit from query parameters
func parseLimit(c *gin.Context, defaultLimit, maxLimit int) int {
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= maxLimit {
			return l
		}
	}
	return defaultLimit
}

// sendJSONResponse writes a JSON response with standard format
func sendJSONResponse(c *gin.Context, data interface{}, count int) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"count":   count,
	})
}

// GetHealthHistory returns recent health checks
func GetHealthHistory(storage *logs_services.HealthStorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := parseLimit(c, 50, 1000)
		checks, err := storage.GetRecentChecks(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve health history",
			})
			return
		}
		sendJSONResponse(c, checks, len(checks))
	}
}

// GetHealthTrends returns trend data for a service
func GetHealthTrends(storage *logs_services.HealthStorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Service name required",
			})
			return
		}

		hours := 24
		if hoursStr := c.Query("hours"); hoursStr != "" {
			if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 720 {
				hours = h
			}
		}

		trend, err := storage.GetTrendData(c.Request.Context(), service, hours)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve trend data",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    trend,
		})
	}
}

// GetHealthPolicies returns all health policies
func GetHealthPolicies(policy *logs_services.HealthPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		policies, err := policy.GetAllPolicies(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve policies",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    policies,
			"count":   len(policies),
		})
	}
}

// GetHealthPolicy returns a specific service's policy
func GetHealthPolicy(policy *logs_services.HealthPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Service name required",
			})
			return
		}

		svcPolicy, err := policy.GetPolicy(c.Request.Context(), service)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Policy not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    svcPolicy,
		})
	}
}

// UpdateHealthPolicy updates a service's health policy
func UpdateHealthPolicy(policy *logs_services.HealthPolicyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Service name required",
			})
			return
		}

		var req struct {
			RepairStrategy    string `json:"repair_strategy"`
			MaxResponseTimeMs int    `json:"max_response_time_ms"`
			AutoRepairEnabled bool   `json:"auto_repair_enabled"`
			AlertOnWarn       bool   `json:"alert_on_warn"`
			AlertOnFail       bool   `json:"alert_on_fail"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		svcPolicy := &logs_services.HealthPolicy{
			ServiceName:       service,
			MaxResponseTimeMs: req.MaxResponseTimeMs,
			AutoRepairEnabled: req.AutoRepairEnabled,
			RepairStrategy:    req.RepairStrategy,
			AlertOnWarn:       req.AlertOnWarn,
			AlertOnFail:       req.AlertOnFail,
		}

		if err := policy.UpdatePolicy(c.Request.Context(), svcPolicy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update policy",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    svcPolicy,
			"message": "Policy updated successfully",
		})
	}
}

// GetRepairHistory returns recent auto-repair actions
func GetRepairHistory(repair *logs_services.AutoRepairService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := parseLimit(c, 50, 1000)
		repairs, err := repair.GetRepairHistory(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve repair history",
			})
			return
		}
		sendJSONResponse(c, repairs, len(repairs))
	}
}

// ManualRepair triggers a manual repair for a service
func ManualRepair(repair *logs_services.AutoRepairService, storage *logs_services.HealthStorageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.Param("service")
		if service == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Service name required",
			})
			return
		}

		var req struct {
			Strategy string `json:"strategy"` // restart or rebuild
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		if req.Strategy == "" {
			req.Strategy = "restart"
		}

		// Execute manual repair
		err := repair.ManualRepair(c.Request.Context(), service, req.Strategy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Repair failed",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "Repair initiated successfully",
			"service":  service,
			"strategy": req.Strategy,
		})
	}
}
