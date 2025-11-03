package review_services

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	review_errors "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/errors"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// ScanService provides Scan Mode analysis for code review sessions.
// It integrates with Ollama for AI-powered code search and stores results in the analysis repository.
// All operations are logged with structured context for observability.
type ScanService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewScanService creates a new ScanService with the given dependencies and logger.
// ollamaClient: AI client for code search
// analysisRepo: Repository for persisting analysis results
// logger: Structured logger for observability
func NewScanService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *ScanService {
	return &ScanService{ollamaClient: ollamaClient, analysisRepo: analysisRepo, logger: logger}
}

// AnalyzeScan performs Scan Mode analysis for the given query and code.
// Returns a ScanModeOutput with matches and summary, or an error if analysis fails.
// Parameter order: query first (what to find), code second (where to search).
func (s *ScanService) AnalyzeScan(ctx context.Context, query string, code string) (*review_models.ScanModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "ScanService.AnalyzeScan",
		trace.WithAttributes(
			attribute.String("query", query),
			attribute.Int("code_length", len(code)),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeScan called", "correlation_id", correlationID, "query", query, "code_length", len(code))

	if query == "" {
		s.logger.Warn("AnalyzeScan: empty query", "correlation_id", correlationID)
		err := &BusinessError{
			Code:       "ERR_INVALID_QUERY",
			Message:    "Search query cannot be empty",
			HTTPStatus: 400,
		}
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, err
	}

	// Build prompt using template
	prompt := BuildScanPrompt(code, query)
	span.SetAttributes(attribute.Int("prompt_length", len(prompt)))

	start := time.Now()
	rawOutput, aiErr := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("ollama_duration_ms", duration.Milliseconds()),
		attribute.Int("response_length", len(rawOutput)),
	)

	if aiErr != nil {
		s.logger.Error("AI call failed", "correlation_id", correlationID, "duration_ms", duration.Milliseconds(), "error", aiErr)
		infraErr := &review_errors.InfrastructureError{
			Code:       "ERR_OLLAMA_UNAVAILABLE",
			Message:    "AI analysis service is unavailable",
			Cause:      aiErr,
			HTTPStatus: http.StatusServiceUnavailable,
		}
		span.RecordError(infraErr)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, infraErr
	}
	s.logger.Info("AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds())

	// Extract JSON from response (handles cases where AI adds extra text)
	jsonStr, extractErr := ExtractJSON(rawOutput)
	if extractErr != nil {
		s.logger.Error("Failed to extract JSON from scan analysis output", "correlation_id", correlationID, "error", extractErr)
		extractErrWrapped := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid response format",
			Cause:      extractErr,
			HTTPStatus: http.StatusBadGateway,
		}
		span.RecordError(extractErrWrapped)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, extractErrWrapped
	}

	var output review_models.ScanModeOutput
	unmarshalErr := json.Unmarshal([]byte(jsonStr), &output)
	if unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal scan analysis output", "correlation_id", correlationID, "error", unmarshalErr)
		parseErr := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid response format",
			Cause:      unmarshalErr,
			HTTPStatus: http.StatusBadGateway,
		}
		span.RecordError(parseErr)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, parseErr
	}

	span.SetAttributes(
		attribute.Bool("error", false),
		attribute.Bool("success", true),
		attribute.Int("matches_count", len(output.Matches)),
	)

	s.logger.Info("AnalyzeScan completed", "correlation_id", correlationID, "summary", output.Summary, "matches_count", len(output.Matches))
	return &output, nil
}

// BusinessError represents a business logic error (invalid input, quota exceeded, etc.)
type BusinessError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *BusinessError) Error() string {
	return e.Message
}
