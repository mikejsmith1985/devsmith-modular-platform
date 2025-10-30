// Package handlers provides HTTP handlers for the logs UI service.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/apps/logs/templates"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/logs/services"
	"github.com/sirupsen/logrus"
)

// UIHandler handles UI requests for the logs service
type UIHandler struct {
	logger        *logrus.Logger
	policyService *services.HealthPolicyService
}

// NewUIHandler creates a new UI handler
func NewUIHandler(logger *logrus.Logger, policyService *services.HealthPolicyService) *UIHandler {
	return &UIHandler{
		logger:        logger,
		policyService: policyService,
	}
}

// DashboardHandler serves the main Logs dashboard.
func (h *UIHandler) DashboardHandler(c *gin.Context) {
	component := templates.Dashboard()
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render dashboard template")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render dashboard"})
	}
}

// HealthHandler serves health check in JSON format.
func (h *UIHandler) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "logs",
		"status":  "healthy",
	})
}

// HealthCheckDashboardHandler serves the system health check dashboard UI.
func (h *UIHandler) HealthCheckDashboardHandler(c *gin.Context) {
	// Get policies for the dashboard
	policies, err := h.policyService.GetAllPolicies(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Warn("Failed to fetch health policies, using empty list")
		policies = []services.HealthPolicy{}
	}

	// Build and run health check
	runner := healthcheck.NewRunner()

	// Add all Phase 1 checks
	runner.AddChecker(&healthcheck.DockerChecker{
		ProjectName: "devsmith-modular-platform",
		Services:    []string{"nginx", "portal", "review", "logs", "analytics", "postgres"},
	})

	services := map[string]string{
		"gateway": "http://localhost:3000/",
		"portal":  "http://localhost:8080/health",
		"review":  "http://localhost:8081/health",
		"logs":    "http://localhost:8082/health",
	}

	for name, url := range services {
		runner.AddChecker(&healthcheck.HTTPChecker{
			CheckName: "http_" + name,
			URL:       url,
		})
	}

	runner.AddChecker(&healthcheck.DatabaseChecker{
		CheckName:     "database",
		ConnectionURL: c.GetString("DATABASE_URL"),
	})

	// Add Phase 2 checks
	runner.AddChecker(&healthcheck.GatewayChecker{
		CheckName:  "gateway_routing",
		ConfigPath: "docker/nginx/nginx.conf",
		GatewayURL: "http://localhost:3000",
	})

	runner.AddChecker(&healthcheck.MetricsChecker{
		CheckName: "performance_metrics",
		Endpoints: []healthcheck.MetricEndpoint{
			{Name: "portal", URL: "http://localhost:8080/health"},
			{Name: "review", URL: "http://localhost:8081/health"},
			{Name: "logs", URL: "http://localhost:8082/health"},
			{Name: "gateway", URL: "http://localhost:3000/"},
		},
	})

	runner.AddChecker(&healthcheck.DependencyChecker{
		CheckName: "service_dependencies",
		Dependencies: map[string][]string{
			"portal":    {},
			"review":    {"portal", "logs"},
			"logs":      {},
			"analytics": {"logs"},
		},
		HealthChecks: map[string]string{
			"portal":    "http://localhost:8080/health",
			"review":    "http://localhost:8081/health",
			"logs":      "http://localhost:8082/health",
			"analytics": "http://localhost:8083/health",
		},
	})

	// Add Phase 3: Trivy security scanning
	runner.AddChecker(&healthcheck.TrivyChecker{
		CheckName: "security_scan",
		ScanType:  "image",
		Targets:   []string{"devsmith/portal:latest", "devsmith/review:latest", "devsmith/logs:latest"},
		TrivyPath: "scripts/trivy-scan.sh",
	})

	// Run all checks
	report := runner.Run()

	// Render dashboard with policies
	component := templates.HealthCheckDashboard(report, policies)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		h.logger.WithError(err).Error("Failed to render health check dashboard")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render health check dashboard"})
	}
}

// RegisterUIRoutes registers the UI routes for the logs handler.
func RegisterUIRoutes(router *gin.Engine, uiHandler *UIHandler) {
	// Dashboard UI route
	router.GET("/", uiHandler.DashboardHandler)
	router.GET("/dashboard", uiHandler.DashboardHandler)

	// Health check dashboard UI
	router.GET("/healthcheck", uiHandler.HealthCheckDashboardHandler)

	// Health check route (simple JSON)
	router.GET("/health", uiHandler.HealthHandler)
}
