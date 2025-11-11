package cmd_logs_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/config"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

// GetHealthCheck runs the full system health check and returns the report
func GetHealthCheck(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	advanced := c.DefaultQuery("advanced", "true") == "true"

	// Create health check runner
	runner := healthcheck.NewRunner()

	// Add Docker container checks
	runner.AddChecker(&healthcheck.DockerChecker{
		ProjectName: "devsmith-modular-platform",
		Services:    []string{"nginx", "portal", "review", "logs", "analytics", "postgres"},
	})

	// Add service health endpoint checks
	services := map[string]string{
		"gateway": config.GetServiceHealthURL("gateway"),
		"portal":  config.GetServiceHealthURL("portal"),
		"review":  config.GetServiceHealthURL("review"),
		"logs":    config.GetServiceHealthURL("logs"),
	}

	for name, url := range services {
		runner.AddChecker(&healthcheck.HTTPChecker{
			CheckName: "http_" + name,
			URL:       url,
		})
	}

	// Add database check
	runner.AddChecker(&healthcheck.DatabaseChecker{
		CheckName:     "database",
		ConnectionURL: config.GetDatabaseURL(),
	})

	// Phase 2: Advanced Diagnostics
	if advanced {
		// Gateway routing validation
		runner.AddChecker(&healthcheck.GatewayChecker{
			CheckName:  "gateway_routing",
			ConfigPath: "docker/nginx/nginx.conf",
			GatewayURL: config.GetGatewayURL(),
		})

		// Performance metrics
		runner.AddChecker(&healthcheck.MetricsChecker{
			CheckName: "performance_metrics",
			Endpoints: []healthcheck.MetricEndpoint{
				{Name: "portal", URL: config.GetServiceHealthURL("portal")},
				{Name: "review", URL: config.GetServiceHealthURL("review")},
				{Name: "logs", URL: config.GetServiceHealthURL("logs")},
				{Name: "gateway", URL: config.GetServiceHealthURL("gateway")},
			},
		})

		// Service dependencies
		runner.AddChecker(&healthcheck.DependencyChecker{
			CheckName: "service_dependencies",
			Dependencies: map[string][]string{
				"portal":    {},
				"review":    {"portal", "logs"},
				"logs":      {},
				"analytics": {"logs"},
			},
			HealthChecks: map[string]string{
				"portal":    config.GetServiceHealthURL("portal"),
				"review":    config.GetServiceHealthURL("review"),
				"logs":      config.GetServiceHealthURL("logs"),
				"analytics": config.GetServiceHealthURL("analytics"),
			},
		})
	}

	// Run all checks
	report := runner.Run()

	// Return appropriate format
	if format == "human" {
		output := healthcheck.FormatHuman(&report)
		c.String(http.StatusOK, output)
		return
	}

	// Default to JSON
	if report.Status == healthcheck.StatusFail {
		c.JSON(http.StatusServiceUnavailable, report)
	} else {
		c.JSON(http.StatusOK, report)
	}
}
