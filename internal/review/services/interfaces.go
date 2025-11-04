package review_services

import (
	"context"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// ====================================================================================
// CORE ANALYZER INTERFACES - Five Reading Modes (Platform Centerpiece)
// ====================================================================================

// PreviewAnalyzer defines the contract for Preview mode analysis.
// Preview mode provides rapid structural assessment of code.
type PreviewAnalyzer interface {
	// AnalyzePreview performs structural analysis on the provided code.
	// Returns PreviewModeOutput or an error if analysis fails.
	// MUST NOT return mock data - fail fast if AI unavailable.
	AnalyzePreview(ctx context.Context, code string) (*review_models.PreviewModeOutput, error)
}

// SkimAnalyzer defines the contract for Skim mode analysis.
// Skim mode focuses on abstractions and high-level understanding.
type SkimAnalyzer interface {
	// AnalyzeSkim performs abstraction-focused analysis on the provided code.
	// Returns SkimModeOutput or an error if analysis fails.
	AnalyzeSkim(ctx context.Context, code string) (*review_models.SkimModeOutput, error)
}

// ScanAnalyzer defines the contract for Scan mode analysis.
// Scan mode performs targeted search within code.
type ScanAnalyzer interface {
	// AnalyzeScan searches for specific patterns or information in code.
	// query: the search term or pattern
	// code: the code to search within
	// Returns ScanModeOutput or an error if analysis fails.
	AnalyzeScan(ctx context.Context, query string, code string) (*review_models.ScanModeOutput, error)
}

// DetailedAnalyzer defines the contract for Detailed mode analysis.
// Detailed mode provides line-by-line deep analysis.
type DetailedAnalyzer interface {
	// AnalyzeDetailed performs comprehensive line-by-line analysis.
	// code: the code to analyze
	// target: specific function/method/section to focus on (optional)
	// Returns DetailedModeOutput or an error if analysis fails.
	AnalyzeDetailed(ctx context.Context, code string, target string) (*review_models.DetailedModeOutput, error)
}

// CriticalAnalyzer defines the contract for Critical mode analysis.
// MOST IMPORTANT: Critical mode evaluates code quality and identifies issues.
// This is the centerpiece of the HITL (Human in the Loop) workflow.
type CriticalAnalyzer interface {
	// AnalyzeCritical performs quality evaluation and issue detection.
	// Returns CriticalModeOutput with categorized issues or an error.
	// Issues must include severity, file location, and suggested fixes.
	AnalyzeCritical(ctx context.Context, code string) (*review_models.CriticalModeOutput, error)
}

// ====================================================================================
// SERVICE REGISTRY - Aggregates all analyzers
// ====================================================================================

// ServiceRegistry aggregates all analyzer interfaces.
// This interface represents a fully-functional review service
// that can perform all five reading modes plus health checks.
type ServiceRegistry interface {
	PreviewAnalyzer
	SkimAnalyzer
	ScanAnalyzer
	DetailedAnalyzer
	CriticalAnalyzer
	HealthChecker
}

// ====================================================================================
// INFRASTRUCTURE INTERFACES - Health, Caching, Observability
// ====================================================================================

// HealthChecker defines the health check contract for services.
// All services MUST implement health checks to validate dependencies.
type HealthChecker interface {
	// HealthCheck validates that the service and its dependencies are operational.
	// Returns nil if healthy, error with details if unhealthy.
	HealthCheck(ctx context.Context) error
}

// ModelSelector defines the contract for model management.
// Services use this to determine which AI model to use for analysis.
type ModelSelector interface {
	// GetAvailableModels returns a list of AI models available for analysis.
	GetAvailableModels(ctx context.Context) ([]review_models.ModelInfo, error)

	// GetDefaultModel returns the default model to use if none specified.
	GetDefaultModel() string

	// ValidateModel checks if a model name is valid and available.
	ValidateModel(ctx context.Context, modelName string) error
}

// AnalysisCache defines the contract for caching analysis results.
// Caching reduces load on AI service and improves response times.
type AnalysisCache interface {
	// Get retrieves cached analysis result by key.
	Get(ctx context.Context, key string) (interface{}, bool, error)

	// Set stores analysis result in cache with expiration.
	Set(ctx context.Context, key string, value interface{}, ttl int) error

	// Invalidate removes cached result by key.
	Invalidate(ctx context.Context, key string) error
}

// PerformanceMonitor defines the contract for performance tracking.
// All services should report performance metrics for observability.
type PerformanceMonitor interface {
	// RecordLatency records the duration of an operation.
	RecordLatency(ctx context.Context, operation string, duration int64)

	// RecordError records an error occurrence.
	RecordError(ctx context.Context, operation string, err error)

	// RecordSuccess records a successful operation.
	RecordSuccess(ctx context.Context, operation string)
}

// CircuitBreaker defines the contract for circuit breaker pattern.
// Prevents cascading failures when AI service is degraded.
type CircuitBreaker interface {
	// Execute runs the function with circuit breaker protection.
	// Returns result or error if circuit is open or function fails.
	Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)

	// State returns current circuit state (closed, half-open, open).
	State() string
}

// ====================================================================================
// AI CLIENT INTERFACE - Ollama Integration
// ====================================================================================

// OllamaClientInterface defines the AI client contract.
// Accepts context and prompt, returns raw AI output or error.
// Enables swapping AI providers and mocking for tests.
type OllamaClientInterface interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// ====================================================================================
// DATA ACCESS INTERFACE - Repository Layer
// ====================================================================================

// AnalysisRepositoryInterface defines the analysis repo contract.
// Used for storing and retrieving analysis results.
// Enables mocking and swapping DB implementations.
type AnalysisRepositoryInterface interface {
	FindByReviewAndMode(ctx context.Context, reviewID int64, mode string) (*review_models.AnalysisResult, error)
	Create(ctx context.Context, result *review_models.AnalysisResult) error
	// DeleteOlderThan removes analysis results older than the provided cutoff time.
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}

// ====================================================================================
// TESTING UTILITIES - Stub Implementations (USE ONLY IN TESTS)
// ====================================================================================

// OllamaClientStub is a stub implementation of OllamaClientInterface for local dev/testing.
// WARNING: This should ONLY be used in test files, NEVER in production code.
type OllamaClientStub struct{}

// Generate returns a stubbed AI output for testing.
func (o *OllamaClientStub) Generate(ctx context.Context, prompt string) (string, error) {
	return `{"functions":[],"interfaces":[],"data_models":[],"workflows":[],"summary":"Stubbed AI output"}`, nil
}
