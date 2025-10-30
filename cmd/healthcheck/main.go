// Package main is the healthcheck CLI application entry point.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

func main() {
	// Parse command-line flags
	format := flag.String("format", "human", "Output format: human or json")
	advanced := flag.Bool("advanced", true, "Include Phase 2 advanced diagnostics")
	flag.Parse()

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
			CheckName: fmt.Sprintf("http_%s", name),
			URL:       url,
		})
	}

	// Add database check
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable"
	}

	runner.AddChecker(&healthcheck.DatabaseChecker{
		CheckName:     "database",
		ConnectionURL: dbURL,
	})

	// Phase 2: Advanced Diagnostics (optional)
	if *advanced {
		// Gateway routing validation
		nginxConfig := os.Getenv("NGINX_CONFIG_PATH")
		if nginxConfig == "" {
			nginxConfig = "docker/nginx/nginx.conf"
		}

		runner.AddChecker(&healthcheck.GatewayChecker{
			CheckName:  "gateway_routing",
			ConfigPath: nginxConfig,
			GatewayURL: "http://localhost:3000",
		})

		// Performance metrics collection
		runner.AddChecker(&healthcheck.MetricsChecker{
			CheckName: "performance_metrics",
			Endpoints: []healthcheck.MetricEndpoint{
				{Name: "portal", URL: "http://localhost:8080/health"},
				{Name: "review", URL: "http://localhost:8081/health"},
				{Name: "logs", URL: "http://localhost:8082/health"},
				{Name: "gateway", URL: "http://localhost:3000/"},
			},
		})

		// Service dependency validation
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

	// Format output
	var output string
	var err error

	switch *format {
	case "json":
		output, err = healthcheck.FormatJSON(report)
	case "human":
		output = healthcheck.FormatHuman(report)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)

	// Exit with appropriate code
	if report.Status == healthcheck.StatusFail {
		os.Exit(1)
	}
}
