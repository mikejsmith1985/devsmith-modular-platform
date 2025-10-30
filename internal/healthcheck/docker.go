package healthcheck

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DockerChecker validates Docker containers are running
type DockerChecker struct {
	ProjectName string
	Services    []string
}

// Name returns the checker name
func (c *DockerChecker) Name() string {
	return "docker_containers"
}

// Check validates all Docker containers are running
func (c *DockerChecker) Check() CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      c.Name(),
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if docker-compose is available
	if err := exec.CommandContext(ctx, "docker-compose", "version").Run(); err != nil {
		result.Status = StatusFail
		result.Message = "docker-compose not available"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// Get running containers
	cmd := exec.CommandContext(ctx, "docker-compose", "ps", "--services", "--filter", "status=running")
	output, err := cmd.Output()
	if err != nil {
		result.Status = StatusFail
		result.Message = "Failed to query docker-compose services"
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	runningServices := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(runningServices) == 1 && runningServices[0] == "" {
		runningServices = []string{}
	}

	// Check each expected service
	missingServices := []string{}
	for _, service := range c.Services {
		found := false
		for _, running := range runningServices {
			if running == service {
				found = true
				break
			}
		}
		if !found {
			missingServices = append(missingServices, service)
		}
	}

	result.Details["expected"] = len(c.Services)
	result.Details["running"] = len(runningServices)
	result.Details["missing"] = missingServices

	if len(missingServices) > 0 {
		result.Status = StatusFail
		result.Message = fmt.Sprintf("%d/%d services running", len(runningServices), len(c.Services))
		result.Error = fmt.Sprintf("Missing services: %s", strings.Join(missingServices, ", "))
	} else {
		result.Status = StatusPass
		result.Message = fmt.Sprintf("All %d services running", len(c.Services))
	}

	result.Duration = time.Since(start)
	return result
}
