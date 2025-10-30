package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/healthcheck"
)

// HealthScheduler runs periodic health checks and repairs
type HealthScheduler struct {
	interval      time.Duration
	storage       *HealthStorageService
	autoRepair    *AutoRepairService
	running       bool
	mu            sync.RWMutex
	stop          chan struct{}
	wg            sync.WaitGroup
	lastCheckTime time.Time
}

// NewHealthScheduler creates a new health scheduler
func NewHealthScheduler(interval time.Duration, storage *HealthStorageService, repair *AutoRepairService) *HealthScheduler {
	return &HealthScheduler{
		interval:   interval,
		storage:    storage,
		autoRepair: repair,
		stop:       make(chan struct{}),
	}
}

// Start begins the scheduled health checks
func (s *HealthScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()

	fmt.Printf("Health scheduler started (interval: %v)\n", s.interval)
}

// Stop stops the scheduled health checks
func (s *HealthScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stop)
	s.wg.Wait()
	fmt.Println("Health scheduler stopped")
}

// run performs the scheduled checks
func (s *HealthScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial check immediately
	s.executeHealthCheck()

	// Run periodic checks
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.executeHealthCheck()
		}
	}
}

// executeHealthCheck runs a full health check and handles repairs
func (s *HealthScheduler) executeHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	start := time.Now()

	// Run the health check using the same configuration as CLI
	report := s.buildAndRunHealthCheck()

	// Store the report
	_, err := s.storage.StoreHealthCheck(ctx, *report, "scheduled")
	if err != nil {
		fmt.Printf("Error storing health check: %v\n", err)
	}

	// Analyze and perform repairs if needed
	if _, err := s.autoRepair.AnalyzeAndRepair(ctx, report); err != nil {
		fmt.Printf("Error during auto-repair: %v\n", err)
	}

	s.mu.Lock()
	s.lastCheckTime = start
	s.mu.Unlock()

	fmt.Printf("Health check completed in %v\n", time.Since(start))
}

// buildAndRunHealthCheck constructs and executes a full health check
func (s *HealthScheduler) buildAndRunHealthCheck() *healthcheck.HealthReport {
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
		ConnectionURL: "postgres://devsmith:devsmith@postgres:5432/devsmith?sslmode=disable",
	})

	// Add Phase 2 checks (advanced diagnostics)
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

	// Add Trivy security scanning
	runner.AddChecker(&healthcheck.TrivyChecker{
		CheckName: "security_scan",
		ScanType:  "image",
		Targets:   []string{"devsmith/portal:latest", "devsmith/review:latest", "devsmith/logs:latest"},
		TrivyPath: "scripts/trivy-scan.sh",
	})

	// Run all checks
	// runner.Run() returns a value; take its address from a local variable
	// to avoid taking the address of a temporary (invalid in Go).
	result := runner.Run()
	return &result
}

// GetLastCheckTime returns when the last health check ran
func (s *HealthScheduler) GetLastCheckTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastCheckTime
}

// IsRunning returns whether the scheduler is currently running
func (s *HealthScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
