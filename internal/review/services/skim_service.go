package review_services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	review_errors "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/errors"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// SkimService provides Skim Mode analysis for code review sessions.
type SkimService struct {
	ollamaClient OllamaClientInterface
	analysisRepo AnalysisRepositoryInterface
	logger       logger.Interface
}

// NewSkimService creates a new SkimService with the given dependencies.
func NewSkimService(ollamaClient OllamaClientInterface, analysisRepo AnalysisRepositoryInterface, logger logger.Interface) *SkimService {
	return &SkimService{
		ollamaClient: ollamaClient,
		analysisRepo: analysisRepo,
		logger:       logger,
	}
}

// AnalyzeSkim performs Skim Mode analysis for the given code.
// Returns function signatures, interfaces, data models WITHOUT implementation details.
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
// Returns error if analysis fails.
func (s *SkimService) AnalyzeSkim(ctx context.Context, code, userMode, outputMode string) (*review_models.SkimModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "SkimService.AnalyzeSkim",
		trace.WithAttributes(
			attribute.Int("code_length", len(code)),
			attribute.String("user_mode", userMode),
			attribute.String("output_mode", outputMode),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeSkim called", "correlation_id", correlationID, "code_length", len(code), "user_mode", userMode, "output_mode", outputMode)

	// Build prompt using template with user/output modes
	prompt := BuildSkimPrompt(code, userMode, outputMode)
	span.SetAttributes(attribute.Int("prompt_length", len(prompt)))

	start := time.Now()
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("ollama_duration_ms", duration.Milliseconds()),
		attribute.Int("response_length", len(rawOutput)),
	)

	if err != nil {
		s.logger.Error("SkimService: AI call failed", "correlation_id", correlationID, "error", err, "duration_ms", duration.Milliseconds())
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
	s.logger.Info("SkimService: AI call succeeded", "correlation_id", correlationID, "duration_ms", duration.Milliseconds())

	output, parseErr := s.parseSkimOutput(rawOutput)
	if parseErr != nil {
		s.logger.Error("SkimService: failed to parse AI output", "correlation_id", correlationID, "error", parseErr)
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

	span.SetAttributes(
		attribute.Bool("error", false),
		attribute.Bool("success", true),
		attribute.Int("functions_count", len(output.Functions)),
		attribute.Int("interfaces_count", len(output.Interfaces)),
	)

	s.logger.Info("SkimService: analysis completed", "correlation_id", correlationID, "functions_count", len(output.Functions))
	return output, nil
}

// Fix parseSkimOutput to handle errors properly
func (s *SkimService) parseSkimOutput(raw string) (*review_models.SkimModeOutput, error) {
	// Extract JSON from response (handles cases where AI adds extra text)
	jsonStr, extractErr := ExtractJSON(raw)
	if extractErr != nil {
		return nil, fmt.Errorf("failed to extract JSON: %w", extractErr)
	}

	var output review_models.SkimModeOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		return nil, fmt.Errorf("failed to parse skim output: %w", err)
	}
	return &output, nil
}
