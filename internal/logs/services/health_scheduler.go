package logs_services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

// HealthScheduler runs periodic health checks in the background
type HealthScheduler struct {
	interval          time.Duration
	storageService    *HealthStorageService
	autoRepairService *AutoRepairService
	running           bool
	mu                sync.Mutex
	stopChan          chan struct{}
}

// NewHealthScheduler creates a new health scheduler
func NewHealthScheduler(interval time.Duration, storage *HealthStorageService, autoRepair *AutoRepairService) *HealthScheduler {
	if interval == 0 {
		interval = 5 * time.Minute
	}
	return &HealthScheduler{
		interval:          interval,
		storageService:    storage,
		autoRepairService: autoRepair,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the background health check scheduler
func (s *HealthScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go s.run()
}

// Stop stops the background health check scheduler
func (s *HealthScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.running = false
		close(s.stopChan)
	}
}

// run is the main scheduler loop
func (s *HealthScheduler) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.executeHealthCheck()
		case <-s.stopChan:
			return
		}
	}
}

// executeHealthCheck runs a full health check and stores results
func (s *HealthScheduler) executeHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	report := s.RunHealthCheck(ctx)

	// Store the results
	healthCheckID, err := s.storageService.StoreHealthCheck(ctx, &report, "scheduled")
	if err != nil {
		fmt.Printf("Failed to store health check: %v\n", err)
		return
	}

	// If there are failures, trigger auto-repair
	if report.Status != healthcheck.StatusPass {
		issues := s.buildIssueMap(&report)
		if len(issues) > 0 {
			_, err := s.autoRepairService.AnalyzeAndRepair(ctx, healthCheckID, issues)
			if err != nil {
				fmt.Printf("Auto-repair failed: %v\n", err)
			}
		}
	}
}

// RunHealthCheck performs a complete health check
func (s *HealthScheduler) RunHealthCheck(ctx context.Context) healthcheck.HealthReport {
	runner := healthcheck.NewRunner()

	// Phase 1: Basic checks
	runner.AddChecker(&healthcheck.DockerChecker{})
	runner.AddChecker(&healthcheck.HTTPChecker{
		CheckName: "portal",
		URL:       "http://localhost:8080/health",
	})
	runner.AddChecker(&healthcheck.HTTPChecker{
		CheckName: "review",
		URL:       "http://localhost:8081/health",
	})
	runner.AddChecker(&healthcheck.HTTPChecker{
		CheckName: "logs",
		URL:       "http://localhost:8082/health",
	})
	runner.AddChecker(&healthcheck.HTTPChecker{
		CheckName: "analytics",
		URL:       "http://localhost:8083/health",
	})
	runner.AddChecker(&healthcheck.DatabaseChecker{
		CheckName:     "postgres",
		ConnectionURL: "postgres://devsmith:devsmith@localhost:5432/devsmith?sslmode=disable",
	})

	// Phase 2: Advanced checks
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

	// Phase 3: Security checks (if Trivy available)
	runner.AddChecker(&healthcheck.TrivyChecker{
		CheckName: "trivy_images",
		ScanType:  "image",
		Targets:   []string{"devsmith-portal", "devsmith-review", "devsmith-logs", "devsmith-analytics"},
	})

	return runner.Run()
}

// buildIssueMap creates a map of service -> issue type from health report
func (s *HealthScheduler) buildIssueMap(report *healthcheck.HealthReport) map[string]string {
	issues := make(map[string]string)

	for _, check := range report.Checks {
		if check.Status != healthcheck.StatusPass {
			// Extract service name from check name
			// e.g., "http_portal" -> "portal", "docker" -> map to services
			serviceName := s.extractServiceName(check.Name)
			if serviceName != "" {
				issueType := s.classifyIssue(&check)
				issues[serviceName] = issueType
			}
		}
	}

	return issues
}

// extractServiceName extracts the service name from a check name
func (s *HealthScheduler) extractServiceName(checkName string) string {
	switch checkName {
	case "http_portal", "portal":
		return "portal"
	case "http_review", "review":
		return "review"
	case "http_logs", "logs":
		return "logs"
	case "http_analytics", "analytics":
		return "analytics"
	case "database", "postgres":
		return "postgres"
	case "docker":
		return ""
	default:
		return ""
	}
}

// classifyIssue determines the issue type from a check result
func (s *HealthScheduler) classifyIssue(check *healthcheck.CheckResult) string {
	message := check.Message
	if check.Error != "" {
		message = check.Error
	}

	// Classify based on error message/type
	if check.Status == healthcheck.StatusFail {
		if message == "" {
			return "unknown"
		}
		if contains(message, "timeout", "deadline", "connection refused") {
			return "timeout"
		}
		if contains(message, "crash", "exit", "killed") {
			return "crash"
		}
		if contains(message, "dependency", "dependent") {
			return "dependency"
		}
		return "unknown"
	}

	return "warning"
}

// contains checks if string contains any of the given substrings
func contains(str string, substrs ...string) bool {
	for _, substr := range substrs {
		if str != "" && substr != "" {
			// Simple substring check
			for i := 0; i < len(str)-len(substr)+1; i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
