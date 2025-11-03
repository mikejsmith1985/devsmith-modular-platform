// Package review_services contains business logic for review service reading modes, including Detailed Mode.
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

// DetailedService provides line-by-line code analysis for Detailed Mode.
// It identifies code complexity, side effects, and data flow between elements.
type DetailedService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewDetailedService creates a new DetailedService with the given Ollama client and analysis repository.
func NewDetailedService(ollama OllamaClientInterface, repo AnalysisRepositoryInterface, logger logger.Interface) *DetailedService {
	return &DetailedService{
		ollamaClient: ollama,
		analysisRepo: repo,
		logger:       logger,
	}
}

// AnalyzeDetailed performs a line-by-line analysis of code in Detailed Mode.
// Returns DetailedModeOutput with line explanations, algorithm analysis, and complexity assessment.
// Returns error if analysis fails.
func (s *DetailedService) AnalyzeDetailed(ctx context.Context, filename string, code string) (*review_models.DetailedModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "DetailedService.AnalyzeDetailed",
		trace.WithAttributes(
			attribute.String("filename", filename),
			attribute.Int("code_length", len(code)),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeDetailed called", "correlation_id", correlationID, "filename", filename, "code_length", len(code))

	if code == "" {
		s.logger.Error("DetailedService: code empty", "correlation_id", correlationID)
		err := &BusinessError{
			Code:       "ERR_INVALID_CODE",
			Message:    "Code cannot be empty",
			HTTPStatus: 400,
		}
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, err
	}

	// Build prompt using template
	prompt := BuildDetailedPrompt(code, filename)
	span.SetAttributes(attribute.Int("prompt_length", len(prompt)))

	start := time.Now()
	resp, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("ollama_duration_ms", duration.Milliseconds()),
		attribute.Int("response_length", len(resp)),
	)

	if err != nil {
		s.logger.Error("DetailedService: AI call failed", "correlation_id", correlationID, "error", err, "duration_ms", duration.Milliseconds())
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
	s.logger.Info("DetailedService: AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds())

	// Extract JSON from response (handles cases where AI adds extra text)
	jsonStr, extractErr := ExtractJSON(resp)
	if extractErr != nil {
		s.logger.Error("DetailedService: failed to extract JSON", "correlation_id", correlationID, "error", extractErr)
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

	var output review_models.DetailedModeOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		s.logger.Error("DetailedService: failed to unmarshal output", "correlation_id", correlationID, "error", err)
		parseErr := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid response format",
			Cause:      err,
			HTTPStatus: http.StatusBadGateway,
		}
		span.RecordError(parseErr)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, parseErr
	}

	span.SetAttributes(
		attribute.Bool("error", false),
		attribute.Bool("success", true),
		attribute.Int("line_explanations_count", len(output.LineExplanations)),
	)

	s.logger.Info("DetailedService: analysis completed", "correlation_id", correlationID, "line_explanations_count", len(output.LineExplanations))
	return &output, nil
}
