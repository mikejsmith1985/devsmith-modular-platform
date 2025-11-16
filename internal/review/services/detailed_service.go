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

	reviewcontext "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/context"
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
// userMode: beginner, novice, intermediate, expert (adjusts explanation tone)
// outputMode: quick (concise), full (includes reasoning trace)
// Returns error if analysis fails.
func (s *DetailedService) AnalyzeDetailed(ctx context.Context, code, target, userMode, outputMode string) (*review_models.DetailedModeOutput, error) {
	// Start tracing span
	tracer := otel.Tracer("devsmith-review")
	ctx, span := tracer.Start(ctx, "DetailedService.AnalyzeDetailed",
		trace.WithAttributes(
			attribute.String("target", target),
			attribute.Int("code_length", len(code)),
			attribute.String("user_mode", userMode),
			attribute.String("output_mode", outputMode),
		),
	)
	defer span.End()

	correlationID := ctx.Value(logger.CorrelationIDKey)
	s.logger.Info("AnalyzeDetailed called", "correlation_id", correlationID, "target", target, "code_length", len(code), "user_mode", userMode, "output_mode", outputMode)

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

	// Build prompt using template with user/output modes
	prompt := BuildDetailedPrompt(code, target, userMode, outputMode)
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
		s.logger.Warn("DetailedService: failed to extract JSON - attempting repair", "correlation_id", correlationID, "error", extractErr)

		// Attempt automatic JSON-repair via a focused AI call.
		repaired, repairErr := s.attemptJSONRepair(ctx, resp)
		if repairErr == nil {
			// try to unmarshal repaired JSON
			var output review_models.DetailedModeOutput
			if uerr := json.Unmarshal([]byte(repaired), &output); uerr == nil {
				s.logger.Info("DetailedService: repaired AI response and parsed successfully", "correlation_id", correlationID)
				// persist repaired analysis for caching/inspection
				_ = s.maybePersistAnalysis(ctx, target, prompt, repaired, resp)
				span.SetAttributes(attribute.Bool("error", false))
				span.SetAttributes(attribute.Int("line_explanations_count", len(output.LineExplanations)))
				return &output, nil
			} else {
				// fall through to record repair failure
				s.logger.Error("DetailedService: repaired JSON still failed to unmarshal", "correlation_id", correlationID, "error", uerr)
			}
		} else {
			s.logger.Error("DetailedService: JSON repair attempt failed", "correlation_id", correlationID, "error", repairErr)
		}

		// include an excerpt of the raw AI response to help debugging / user guidance
		excerpt := resp
		if len(excerpt) > 800 {
			excerpt = excerpt[:800] + "..."
		}
		extractErrWrapped := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid response format and automatic repair failed. Raw response excerpt: " + excerpt,
			Cause:      extractErr,
			HTTPStatus: http.StatusBadGateway,
		}
		// persist the original raw response for short-term troubleshooting
		_ = s.maybePersistAnalysis(ctx, target, prompt, resp, resp)
		span.RecordError(extractErrWrapped)
		span.SetAttributes(attribute.Bool("error", true))
		return nil, extractErrWrapped
	}

	var output review_models.DetailedModeOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		s.logger.Warn("DetailedService: failed to unmarshal output - attempting repair", "correlation_id", correlationID, "error", err)

		// Try to repair the JSON using the AI
		repaired, repairErr := s.attemptJSONRepair(ctx, resp)
		if repairErr == nil {
			if uerr := json.Unmarshal([]byte(repaired), &output); uerr == nil {
				s.logger.Info("DetailedService: repaired AI output and parsed successfully", "correlation_id", correlationID)
				_ = s.maybePersistAnalysis(ctx, target, prompt, repaired, resp)
				span.SetAttributes(attribute.Bool("error", false))
				span.SetAttributes(attribute.Int("line_explanations_count", len(output.LineExplanations)))
				return &output, nil
			} else {
				s.logger.Error("DetailedService: repaired output still invalid", "correlation_id", correlationID, "error", uerr)
			}
		} else {
			s.logger.Error("DetailedService: JSON repair attempt failed", "correlation_id", correlationID, "error", repairErr)
		}

		excerpt := jsonStr
		if len(excerpt) > 800 {
			excerpt = excerpt[:800] + "..."
		}
		parseErr := &review_errors.InfrastructureError{
			Code:       "ERR_AI_RESPONSE_INVALID",
			Message:    "AI returned invalid JSON structure and automatic repair failed. Excerpt: " + excerpt,
			Cause:      err,
			HTTPStatus: http.StatusBadGateway,
		}
		// persist the problematic JSON for troubleshooting
		_ = s.maybePersistAnalysis(ctx, target, prompt, jsonStr, resp)
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

// attemptJSONRepair asks the AI to extract/repair JSON from a raw AI response.
// It returns the repaired JSON string (not further validated) or an error.
func (s *DetailedService) attemptJSONRepair(ctx context.Context, rawAI string) (string, error) {
	// Build a comprehensive repair prompt with explicit schema and fallback strategy
	repairPrompt := `The previous AI response was intended to be valid JSON for Detailed Mode code analysis.
Your task: Extract and return ONLY valid JSON that matches the following schema EXACTLY.

REQUIRED SCHEMA:
{
  "summary": "string - Brief overview of what the code does",
  "line_explanations": [
    {
      "line_number": number,
      "code": "string - The actual line of code",
      "explanation": "string - What this line does and why"
    }
  ],
  "algorithm_summary": "string - High-level explanation of the algorithm/logic",
  "complexity": "string - Time/space complexity (e.g., 'O(n)', 'O(1)')",
  "edge_cases": ["string - List of edge cases to consider"],
  "variable_tracking": [
    {
      "line_number": number,
      "variables": {
        "variable_name": "current_value_or_state"
      }
    }
  ],
  "control_flow": [
    {
      "type": "string - 'if', 'loop', 'function_call', etc.",
      "line_number": number,
      "description": "string - What this control flow does"
    }
  ]
}

RULES:
1. Return ONLY the JSON object - no explanatory text before or after
2. All fields are REQUIRED (use empty arrays [] or empty strings "" if no data)
3. Ensure all brackets, braces, and quotes are properly closed
4. Do not include any markdown code fences or language tags
5. If the original response is completely unusable, return minimal valid JSON with "summary": "Unable to parse code"

FALLBACK STRATEGY:
If you cannot extract meaningful analysis, return this minimal valid JSON:
{
  "summary": "Analysis could not be completed",
  "line_explanations": [],
  "algorithm_summary": "N/A",
  "complexity": "Unknown",
  "edge_cases": [],
  "variable_tracking": [],
  "control_flow": []
}

Now extract/repair JSON from this response:`
	fullPrompt := repairPrompt + "\n\nRaw output:\n" + rawAI + "\n\nReturn only JSON. If you cannot produce valid JSON, return an empty JSON object with the keys present and empty values."

	repairedResp, err := s.ollamaClient.Generate(ctx, fullPrompt)
	if err != nil {
		s.logger.Error("DetailedService: attemptJSONRepair failed to call AI", "error", err)
		return "", err
	}

	// Try to extract JSON from the repair response
	jsonStr, extractErr := ExtractJSON(repairedResp)
	if extractErr != nil {
		s.logger.Error("DetailedService: attemptJSONRepair could not extract JSON from AI repair response", "error", extractErr)
		return "", extractErr
	}
	return jsonStr, nil
}

// maybePersistAnalysis persists a minimal AnalysisResult for troubleshooting.
// This is intentionally best-effort: errors are logged but not returned to callers
// except as debug info (caller may ignore the returned error).
func (s *DetailedService) maybePersistAnalysis(ctx context.Context, filename string, prompt string, rawOutput string, rawAI string) error {
	res := &review_models.AnalysisResult{
		Mode:      review_models.DetailedMode,
		Prompt:    prompt,
		Summary:   "AUTO_CAPTURE: raw AI output captured for troubleshooting",
		Metadata:  "filename=" + filename + "; captured_at=" + time.Now().Format(time.RFC3339),
		ModelUsed: "",
		RawOutput: rawOutput,
		ReviewID:  0,
	}

	// Try to capture model from context if provided
	if ctx != nil {
		if m, ok := ctx.Value(reviewcontext.ModelContextKey).(string); ok && m != "" {
			res.ModelUsed = m
		}
	}

	if s.analysisRepo == nil {
		s.logger.Warn("DetailedService: analysisRepo is nil; skipping persistence of AI output")
		return nil
	}

	if err := s.analysisRepo.Create(ctx, res); err != nil {
		s.logger.Error("DetailedService: failed to persist analysis result", "error", err)
		return err
	}
	s.logger.Info("DetailedService: persisted analysis result for troubleshooting", "filename", filename)
	return nil
}
