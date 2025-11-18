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

// PreviewService provides Preview Mode analysis for code review sessions.
type PreviewService struct {
	ollamaClient OllamaClientInterface
	logger       logger.Interface
}

// NewPreviewService creates a new PreviewService with the given dependencies.
func NewPreviewService(ollamaClient OllamaClientInterface, logger logger.Interface) *PreviewService {
	return &PreviewService{
		ollamaClient: ollamaClient,
		logger:       logger,
	}
}

// AnalyzePreview performs Preview Mode analysis for the given code.
// Returns rapid structural assessment.
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
// Returns error if analysis fails.
func (s *PreviewService) AnalyzePreview(ctx context.Context, code, userMode, outputMode string) (*review_models.PreviewModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "PreviewService.AnalyzePreview",
		trace.WithAttributes(
			attribute.Int("code_length", len(code)),
			attribute.String("user_mode", userMode),
			attribute.String("output_mode", outputMode),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzePreview called", "correlation_id", correlationID, "code_length", len(code), "user_mode", userMode, "output_mode", outputMode)

	// Build prompt using template with user/output modes
	prompt := BuildPreviewPrompt(code, userMode, outputMode)
	span.SetAttributes(attribute.Int("prompt_length", len(prompt)))

	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("ollama_duration_ms", duration.Milliseconds()),
		attribute.Int("response_length", len(rawOutput)),
	)

	if err != nil {
		s.logger.Error("PreviewService: AI call failed", "correlation_id", correlationID, "error", err, "duration_ms", duration.Milliseconds())
		aiErr := &review_errors.InfrastructureError{
			Code:       "ERR_OLLAMA_UNAVAILABLE",
			Message:    "AI analysis service is unavailable",
			Cause:      err,
			HTTPStatus: http.StatusServiceUnavailable,
		}
		span.RecordError(aiErr)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, aiErr
	}
	s.logger.Info("PreviewService: AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds())

	// DEBUG: Log raw AI output
	s.logger.Info("DEBUG PreviewService raw AI output", "correlation_id", correlationID, "output_length", len(rawOutput), "first_100_chars", func() string {
		if len(rawOutput) > 100 {
			return rawOutput[:100]
		}
		return rawOutput
	}())

	// Extract JSON from response (handles cases where AI adds extra text)
	jsonStr, extractErr := ExtractJSON(rawOutput)
	if extractErr != nil {
		s.logger.Error("PreviewService: failed to extract JSON", "correlation_id", correlationID, "error", extractErr)
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
	var output review_models.PreviewModeOutput
	if parseErr := json.Unmarshal([]byte(jsonStr), &output); parseErr != nil {
		s.logger.Error("PreviewService: failed to parse AI output", "correlation_id", correlationID, "error", parseErr)
		parseErrWrapped := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid response format",
			Cause:      parseErr,
			HTTPStatus: http.StatusBadGateway,
		}
		span.RecordError(parseErrWrapped)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, parseErrWrapped
	}

	// Validate output structure
	if output.Summary == "" {
		output.Summary = "No summary provided by AI"
	}

	span.SetAttributes(
		attribute.Bool("error", false),
		attribute.Bool("success", true),
		attribute.Int("bounded_contexts_count", len(output.BoundedContexts)),
		attribute.Int("tech_stack_count", len(output.TechStack)),
	)

	s.logger.Info("PreviewService: analysis completed successfully", "correlation_id", correlationID, "bounded_contexts_count", len(output.BoundedContexts))
	return &output, nil
}
