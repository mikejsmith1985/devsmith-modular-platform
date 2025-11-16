// Package review_health provides health check capabilities for the review service.
package review_health

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	// Fast connectivity test using /api/tags endpoint instead of generation
	testCtx, cancel := context.WithTimeout(ctx, 2*time.Second) // Reduced timeout
	defer cancel()

	// Get Ollama endpoint from environment or use default
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434"
	}

	// Make simple GET request to /api/tags (much faster than generation)
	req, err := http.NewRequestWithContext(testCtx, "GET", ollamaEndpoint+"/api/tags", nil)
	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Failed to create health check request: %v", err)
		comp.ResponseTime = time.Since(start)
		return comp
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	comp.ResponseTime = time.Since(start)

	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Ollama service unreachable: %v", err)
		return comp
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Ollama returned HTTP %d", resp.StatusCode)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Ollama service is reachable"
	return comp
}

// checkOllamaModel verifies the required model is available.
func (h *ServiceHealthChecker) checkOllamaModel(ctx context.Context) ComponentHealth {
	start := time.Now()

	// Get required model from environment or use default
	required := os.Getenv("OLLAMA_MODEL")
	if required == "" {
		required = "mistral:7b-instruct"
	}

	comp := ComponentHealth{
		Name: "ollama_model",
		Metadata: map[string]string{
			"required_model": required,
		},
	}

	// Fast model check using /api/tags instead of generation
	testCtx, cancel := context.WithTimeout(ctx, 2*time.Second) // Reduced timeout
	defer cancel()

	// Get Ollama endpoint from environment or use default
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434"
	}

	// Make request to /api/tags to list available models
	req, err := http.NewRequestWithContext(testCtx, "GET", ollamaEndpoint+"/api/tags", nil)
	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Failed to create model check request: %v", err)
		comp.ResponseTime = time.Since(start)
		return comp
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	comp.ResponseTime = time.Since(start)

	if err != nil {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Failed to check available models: %v", err)
		return comp
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		comp.Status = HealthStatusUnhealthy
		comp.Message = fmt.Sprintf("Model check returned HTTP %d", resp.StatusCode)
		return comp
	}

	// Parse the tags response to check if our model is available
	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		comp.Status = HealthStatusDegraded
		comp.Message = fmt.Sprintf("Failed to parse models list: %v", err)
		return comp
	}

	// Check if required model is in the list
	modelFound := false
	for _, model := range tagsResp.Models {
		if model.Name == required {
			modelFound = true
			break
		}
	}

	if !modelFound {
		comp.Status = HealthStatusDegraded
		comp.Message = fmt.Sprintf("Required model '%s' not found in available models", required)
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = fmt.Sprintf("Model %s is available", required)
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

	// Check if reviews schema exists (note: plural 'reviews' not singular 'review')
	var schemaExists bool
	err = h.db.QueryRowContext(testCtx, "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'reviews')").Scan(&schemaExists)
	if err != nil || !schemaExists {
		comp.Status = HealthStatusDegraded
		comp.Message = "Database connected but reviews schema missing"
		return comp
	}

	comp.Status = HealthStatusHealthy
	comp.Message = "Database connected and reviews schema present"
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
