package healthcheck

import (
	"fmt"
	"time"
)

// DependencyChecker validates service interdependencies
type DependencyChecker struct {
	CheckName    string
	Dependencies map[string][]string // service -> list of dependencies
	HealthChecks map[string]string   // service -> health check URL
}

// ServiceDependency represents a service and its status
type ServiceDependency struct {
	Service      string   `json:"service"`
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies"`
	HealthyDeps  int      `json:"healthy_deps"`
	TotalDeps    int      `json:"total_deps"`
}

// Name returns the checker name
func (c *DependencyChecker) Name() string {
	return c.CheckName
}

// Check validates service interdependencies
func (c *DependencyChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.CheckName,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Check health of all services first
	serviceHealth := make(map[string]bool)
	for service, healthURL := range c.HealthChecks {
		serviceHealth[service] = c.isServiceHealthy(healthURL)
	}

	// Analyze dependencies
	dependencyStatuses := []ServiceDependency{}
	unhealthyChains := []string{}
	healthyServices := 0
	totalServices := len(c.Dependencies)

	for service, deps := range c.Dependencies {
		healthyDeps := 0
		for _, dep := range deps {
			if serviceHealth[dep] {
				healthyDeps++
			}
		}

		status := "healthy"
		if !serviceHealth[service] {
			status = "unhealthy"
		} else if healthyDeps < len(deps) {
			status = "degraded"
			unhealthyChains = append(unhealthyChains, 
				fmt.Sprintf("%s (missing: %d/%d deps)", service, len(deps)-healthyDeps, len(deps)))
		} else {
			healthyServices++
		}

		dependencyStatuses = append(dependencyStatuses, ServiceDependency{
			Service:      service,
			Status:       status,
			Dependencies: deps,
			HealthyDeps:  healthyDeps,
			TotalDeps:    len(deps),
		})
	}

	result.Details["dependency_status"] = dependencyStatuses
	result.Details["healthy_services"] = healthyServices
	result.Details["total_services"] = totalServices
	result.Details["unhealthy_chains"] = unhealthyChains

	// Determine overall status
	if healthyServices == totalServices {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("All %d services and dependencies healthy", totalServices)
	} else if len(unhealthyChains) > 0 {
		result.Status = StatusWarn
		result.Message = fmt.Sprintf("%d services have unhealthy dependencies", len(unhealthyChains))
		result.Error = fmt.Sprintf("Affected: %v", unhealthyChains)
	} else {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("%d/%d services unhealthy", totalServices-healthyServices, totalServices)
	}

	result.Duration = time.Since(start)
	return result
}

// isServiceHealthy checks if a service is responding to health checks
func (c *DependencyChecker) isServiceHealthy(healthURL string) bool {
	checker := &HTTPChecker{
		CheckName: "temp",
		URL:       healthURL,
	}
	checkResult := checker.Check()
	return checkResult.Status == StatusPass
}

