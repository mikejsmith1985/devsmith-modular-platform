// Package review_services provides business logic services for the Review Service
package review_services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

// ExportFormat defines supported export formats
type ExportFormat string

// Export format constants
const (
	FormatPDF      ExportFormat = "pdf"
	FormatMarkdown ExportFormat = "md"
	FormatJSON     ExportFormat = "json"
)

// ExportRequest contains parameters for exporting analysis results
type ExportRequest struct {
	Format       ExportFormat `json:"format"`
	SessionID    int64        `json:"session_id"`
	IncludeCode  bool         `json:"include_code"`
	IncludeRawAI bool         `json:"include_raw_ai"`
}

// ExportResult contains exported analysis data
type ExportResult struct {
	AnalysisResult *review_models.AnalysisResult
	Content        []byte
	Filename       string
	MimeType       string
	ExportedAt     time.Time
	SessionID      int64
	Format         ExportFormat
}

// ExportService handles exporting analysis results in various formats
type ExportService struct {
	logger logger.Interface
}

// NewExportService creates a new export service
func NewExportService(logger logger.Interface) *ExportService {
	return &ExportService{
		logger: logger,
	}
}

// Export exports analysis results in the specified format
func (s *ExportService) Export(ctx context.Context, analysis *review_models.AnalysisResult, format ExportFormat) (*ExportResult, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	if analysis == nil {
		return nil, fmt.Errorf("export: cannot export nil analysis")
	}

	switch format {
	case FormatJSON:
		return s.exportJSON(ctx, analysis)
	case FormatMarkdown:
		return s.exportMarkdown(ctx, analysis)
	case FormatPDF:
		return s.exportPDF(ctx, analysis)
	default:
		return nil, fmt.Errorf("export: unsupported format: %s", format)
	}
}

// exportJSON exports analysis as JSON
func (s *ExportService) exportJSON(_ context.Context, analysis *review_models.AnalysisResult) (*ExportResult, error) {
	// Create JSON export structure
	exportData := map[string]interface{}{
		"export_date": time.Now().UTC(),
		"mode":        analysis.Mode,
		"summary":     analysis.Summary,
		"metadata":    json.RawMessage(analysis.Metadata),
		"model_used":  analysis.ModelUsed,
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		s.logger.Error("failed to marshal JSON", "error", err)
		return nil, fmt.Errorf("export: JSON marshaling failed: %w", err)
	}

	return &ExportResult{
		SessionID:      analysis.ReviewID,
		Format:         FormatJSON,
		Content:        jsonBytes,
		MimeType:       "application/json",
		Filename:       fmt.Sprintf("analysis_%d_%s.json", analysis.ReviewID, analysis.Mode),
		ExportedAt:     time.Now(),
		AnalysisResult: analysis,
	}, nil
}

// exportMarkdown exports analysis as Markdown
func (s *ExportService) exportMarkdown(_ context.Context, analysis *review_models.AnalysisResult) (*ExportResult, error) {
	var buf bytes.Buffer

	// Write markdown header
	buf.WriteString(fmt.Sprintf("# Code Review Analysis - %s Mode\n\n", analysis.Mode))
	buf.WriteString(fmt.Sprintf("**Exported:** %s\n\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("**Mode:** `%s`\n", analysis.Mode))
	buf.WriteString(fmt.Sprintf("**Model:** `%s`\n\n", analysis.ModelUsed))

	// Write summary
	buf.WriteString("## Summary\n\n")
	buf.WriteString(analysis.Summary)
	buf.WriteString("\n\n")

	// Write metadata
	if analysis.Metadata != "" {
		buf.WriteString("## Analysis Details\n\n")
		buf.WriteString("```json\n")
		buf.WriteString(analysis.Metadata)
		buf.WriteString("\n```\n\n")
	}

	// Write raw output (optional)
	if analysis.RawOutput != "" {
		buf.WriteString("## Raw AI Output\n\n")
		buf.WriteString("```\n")
		buf.WriteString(analysis.RawOutput)
		buf.WriteString("\n```\n")
	}

	return &ExportResult{
		SessionID:      analysis.ReviewID,
		Format:         FormatMarkdown,
		Content:        buf.Bytes(),
		MimeType:       "text/markdown",
		Filename:       fmt.Sprintf("analysis_%d_%s.md", analysis.ReviewID, analysis.Mode),
		ExportedAt:     time.Now(),
		AnalysisResult: analysis,
	}, nil
}

// exportPDF exports analysis as PDF (stub for now - requires PDF library)
func (s *ExportService) exportPDF(ctx context.Context, analysis *review_models.AnalysisResult) (*ExportResult, error) {
	// TODO: Implement PDF export using a library like gofpdf
	// For now, return markdown converted to PDF-like structure
	s.logger.Warn("PDF export not yet implemented, returning markdown")

	// Generate markdown first
	mdResult, err := s.exportMarkdown(ctx, analysis)
	if err != nil {
		return nil, err
	}

	// Update result for PDF
	mdResult.Format = FormatPDF
	mdResult.MimeType = "application/pdf"
	mdResult.Filename = fmt.Sprintf("analysis_%d_%s.pdf", analysis.ReviewID, analysis.Mode)

	return mdResult, nil
}
