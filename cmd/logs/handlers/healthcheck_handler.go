package cmd_logs_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	// Add database check
	runner.AddChecker(&healthcheck.DatabaseChecker{
		CheckName:     "database",
		ConnectionURL: c.GetString("DATABASE_URL"),
	})

	// Phase 2: Advanced Diagnostics
	if advanced {
		// Gateway routing validation
		runner.AddChecker(&healthcheck.GatewayChecker{
			CheckName:  "gateway_routing",
			ConfigPath: "docker/nginx/nginx.conf",
			GatewayURL: "http://localhost:3000",
		})

		// Performance metrics
		runner.AddChecker(&healthcheck.MetricsChecker{
			CheckName: "performance_metrics",
			Endpoints: []healthcheck.MetricEndpoint{
				{Name: "portal", URL: "http://localhost:8080/health"},
				{Name: "review", URL: "http://localhost:8081/health"},
				{Name: "logs", URL: "http://localhost:8082/health"},
				{Name: "gateway", URL: "http://localhost:3000/"},
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
				"portal":    "http://localhost:8080/health",
				"review":    "http://localhost:8081/health",
				"logs":      "http://localhost:8082/health",
				"analytics": "http://localhost:8083/health",
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
