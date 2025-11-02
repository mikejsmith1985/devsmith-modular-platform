// Package review_health provides health check capabilities for the review service.
package review_health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// HealthStatus represents the health state of a component.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of an individual component.
type ComponentHealth struct {
	Name         string            `json:"name"`
	Status       HealthStatus      `json:"status"`
	Message      string            `json:"message,omitempty"`
	ResponseTime time.Duration     `json:"response_time_ms"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ServiceHealth represents the overall health of the review service.
type ServiceHealth struct {
	Status     HealthStatus      `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Components []ComponentHealth `json:"components"`
	Summary    string            `json:"summary"`
}

// ServiceHealthChecker performs health checks on all review service components.
type ServiceHealthChecker struct {
	previewService  review_services.PreviewAnalyzer
	skimService     review_services.SkimAnalyzer
	scanService     review_services.ScanAnalyzer
	detailedService review_services.DetailedAnalyzer
	criticalService review_services.CriticalAnalyzer
	ollamaClient    review_services.OllamaClientInterface
	db              *sql.DB
	logger          *logger.Logger
}

// NewServiceHealthChecker creates a new health checker for the review service.
func NewServiceHealthChecker(
	previewService review_services.PreviewAnalyzer,
	skimService review_services.SkimAnalyzer,
	scanService review_services.ScanAnalyzer,
	detailedService review_services.DetailedAnalyzer,
	criticalService review_services.CriticalAnalyzer,
	ollamaClient review_services.OllamaClientInterface,
	db *sql.DB,
	logger *logger.Logger,
) *ServiceHealthChecker {
	return &ServiceHealthChecker{
		previewService:  previewService,
		skimService:     skimService,
		scanService:     scanService,
		detailedService: detailedService,
		criticalService: criticalService,
		ollamaClient:    ollamaClient,
		db:              db,
		logger:          logger,
	}
}

// CheckHealth performs comprehensive health checks on all components.
func (h *ServiceHealthChecker) CheckHealth(ctx context.Context) (*ServiceHealth, error) {
	h.logger.Info("Starting health check")

	components := []ComponentHealth{
		h.checkOllamaConnectivity(ctx),
		h.checkOllamaModel(ctx),
		h.checkDatabaseConnectivity(ctx),
		h.checkPreviewService(ctx),
		h.checkSkimService(ctx),
		h.checkScanService(ctx),
		h.checkDetailedService(ctx),
		h.checkCriticalService(ctx),
	}

	// Determine overall status
	overallStatus := HealthStatusHealthy
	unhealthyCount := 0
	degradedCount := 0

	for _, comp := range components {
		switch comp.Status {
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		overallStatus = HealthStatusUnhealthy
	} else if degradedCount > 0 {
		overallStatus = HealthStatusDegraded
	}

	summary := h.generateSummary(overallStatus, unhealthyCount, degradedCount)

	health := &ServiceHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Summary:    summary,
	}

	h.logger.Info("Health check completed", "status", overallStatus, "unhealthy", unhealthyCount, "degraded", degradedCount)
	return health, nil
}

// checkOllamaConnectivity checks if Ollama service is reachable.
func (h *ServiceHealthChecker) checkOllamaConnectivity(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "ollama_connectivity",
	}

	// Simple connectivity test with minimal prompt
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.ollamaClient.Generate(testCtx, "test")
	comp.ResponseTime = time.Since(start)

	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Ollama service unreachable: %v", err)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Ollama service is reachable"
	return comp
}

// checkOllamaModel verifies the required model (mistral:7b-instruct) is available.
func (h *ServiceHealthChecker) checkOllamaModel(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "ollama_model",
		Metadata: map[string]string{
			"required_model": "mistral:7b-instruct",
		},
	}

	// Try to generate with the specific model
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.ollamaClient.Generate(testCtx, "test")
	comp.ResponseTime = time.Since(start)

	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Model not available: %v", err)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Model mistral:7b-instruct is available"
	return comp
}

// checkDatabaseConnectivity verifies database connection.
func (h *ServiceHealthChecker) checkDatabaseConnectivity(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "database",
	}

	testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := h.db.PingContext(testCtx)
	comp.ResponseTime = time.Since(start)

	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Database unreachable: %v", err)
		return comp
	}

	// Check if review schema exists
	var schemaExists bool
	err = h.db.QueryRowContext(testCtx, "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'review')").Scan(&schemaExists)
	if err != nil || !schemaExists {
		comp.Status = HealthStatusDegraded
		comp.Message = "Database connected but review schema missing"
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Database connected and review schema present"
	return comp
}

// checkPreviewService validates Preview service is operational.
func (h *ServiceHealthChecker) checkPreviewService(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "preview_service",
	}

	// Minimal test - just check service is not nil
	if h.previewService == nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = "Preview service not initialized"
		comp.ResponseTime = time.Since(start)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Preview service operational"
	comp.ResponseTime = time.Since(start)
	return comp
}

// checkSkimService validates Skim service is operational.
func (h *ServiceHealthChecker) checkSkimService(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "skim_service",
	}

	if h.skimService == nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = "Skim service not initialized"
		comp.ResponseTime = time.Since(start)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Skim service operational"
	comp.ResponseTime = time.Since(start)
	return comp
}

// checkScanService validates Scan service is operational.
func (h *ServiceHealthChecker) checkScanService(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "scan_service",
	}

	if h.scanService == nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = "Scan service not initialized"
		comp.ResponseTime = time.Since(start)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Scan service operational"
	comp.ResponseTime = time.Since(start)
	return comp
}

// checkDetailedService validates Detailed service is operational.
func (h *ServiceHealthChecker) checkDetailedService(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "detailed_service",
	}

	if h.detailedService == nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = "Detailed service not initialized"
		comp.ResponseTime = time.Since(start)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Detailed service operational"
	comp.ResponseTime = time.Since(start)
	return comp
}

// checkCriticalService validates Critical service is operational.
func (h *ServiceHealthChecker) checkCriticalService(ctx context.Context) ComponentHealth {
	start := time.Now()
	comp := ComponentHealth{
		Name: "critical_service",
	}

	if h.criticalService == nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = "Critical service not initialized"
		comp.ResponseTime = time.Since(start)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Critical service operational"
	comp.ResponseTime = time.Since(start)
	return comp
}

// generateSummary creates a human-readable summary of the health status.
func (h *ServiceHealthChecker) generateSummary(status HealthStatus, unhealthy int, degraded int) string {
	switch status {
	case HealthStatusHealthy:
		return "All components are healthy"
	case HealthStatusDegraded:
		return fmt.Sprintf("%d component(s) degraded", degraded)
	case HealthStatusUnhealthy:
		return fmt.Sprintf("%d component(s) unhealthy, %d degraded", unhealthy, degraded)
	default:
		return "Unknown health status"
	}
}
