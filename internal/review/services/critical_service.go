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

// CriticalService provides methods for analyzing repositories in Critical Mode.
// It identifies issues such as security vulnerabilities, bugs, performance problems, and code smells.
type CriticalService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewCriticalService creates a new instance of CriticalService with the provided dependencies.
func NewCriticalService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *CriticalService {
	return &CriticalService{ollamaClient: ollamaClient, analysisRepo: analysisRepo, logger: logger}
}

// AnalyzeCritical performs a detailed quality analysis of code in Critical Mode.
// Evaluates architecture, security, performance, and identifies improvements.
// Returns error if analysis fails.
func (s *CriticalService) AnalyzeCritical(ctx context.Context, code string) (*review_models.CriticalModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "CriticalService.AnalyzeCritical",
		trace.WithAttributes(
			attribute.Int("code_length", len(code)),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeCritical called", "correlation_id", correlationID, "code_length", len(code))

	// Build prompt using template
	prompt := BuildCriticalPrompt(code)
	span.SetAttributes(attribute.Int("prompt_length", len(prompt)))

	// Call Ollama for real analysis
	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("ollama_duration_ms", duration.Milliseconds()),
		attribute.Int("response_length", len(rawOutput)),
	)

	if err != nil {
		s.logger.Error("Critical analysis AI call failed", "correlation_id", correlationID, "error", err, "duration_ms", duration.Milliseconds())
		infraErr := &review_errors.InfrastructureError{
			Code:       "ERR_OLLAMA_UNAVAILABLE",
			Message:    "AI analysis service is unavailable",
			Cause:      err,
			HTTPStatus: http.StatusServiceUnavailable,
		}
		span.RecordError(infraErr)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, infraErr
	}
	s.logger.Info("Critical analysis AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds(), "output_length", len(rawOutput))

	// Extract JSON from response (handles cases where AI adds extra text)
	jsonStr, extractErr := ExtractJSON(rawOutput)
	if extractErr != nil {
		s.logger.Error("Failed to extract JSON from critical analysis output", "correlation_id", correlationID, "error", extractErr)
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

	// Parse JSON response
	var output review_models.CriticalModeOutput
	if unmarshalErr := json.Unmarshal([]byte(jsonStr), &output); unmarshalErr != nil {
		s.logger.Error("Failed to unmarshal critical analysis output", "correlation_id", correlationID, "error", unmarshalErr)
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

	// Validate output structure
	if output.Summary == "" {
		s.logger.Warn("Critical analysis returned empty summary", "correlation_id", correlationID)
		output.Summary = "Analysis completed but summary was empty"
	}

	span.SetAttributes(
		attribute.Bool("error", false),
		attribute.Bool("success", true),
		attribute.Int("issues_count", len(output.Issues)),
		attribute.String("overall_grade", output.OverallGrade),
	)

	s.logger.Info("Critical analysis completed", "correlation_id", correlationID, "issues_found", len(output.Issues), "grade", output.OverallGrade)
	return &output, nil
}
