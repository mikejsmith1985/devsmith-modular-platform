package review_services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/shared/logger"
)

func createTestLogger(t *testing.T) logger.Interface {
	cfg := &logger.Config{
		ServiceName: "test",
		LogLevel:    "info",
		LogToStdout: false,
	}
	log, err := logger.NewLogger(cfg)
	require.NoError(t, err)
	return log
}

func TestExportService_ExportJSON_Success(t *testing.T) {
	// GIVEN: Export service and analysis result
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID:  123,
		Mode:      "skim",
		Summary:   "Test summary",
		Metadata:  `{"key":"value"}`,
		RawOutput: "raw output",
		ModelUsed: "qwen2.5-coder:32b",
	}

	// WHEN: Exporting to JSON
	result, err := service.Export(context.Background(), analysis, FormatJSON)

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, FormatJSON, result.Format)
	assert.Equal(t, "application/json", result.MimeType)
	assert.Contains(t, result.Filename, "analysis_123_skim.json")

	// AND: JSON should be valid
	var data map[string]interface{}
	err = json.Unmarshal(result.Content, &data)
	assert.NoError(t, err)
	assert.Equal(t, "skim", data["mode"])
	assert.Equal(t, "Test summary", data["summary"])
}

func TestExportService_ExportMarkdown_Success(t *testing.T) {
	// GIVEN: Export service and analysis result
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID:  456,
		Mode:      "scan",
		Summary:   "Scan findings",
		Metadata:  `{"findings":["issue1","issue2"]}`,
		RawOutput: "detailed output",
		ModelUsed: "ollama",
	}

	// WHEN: Exporting to Markdown
	result, err := service.Export(context.Background(), analysis, FormatMarkdown)

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, FormatMarkdown, result.Format)
	assert.Equal(t, "text/markdown", result.MimeType)

	// AND: Markdown should contain expected content
	content := string(result.Content)
	assert.Contains(t, content, "# Code Review Analysis")
	assert.Contains(t, content, "scan Mode")
	assert.Contains(t, content, "Scan findings")
	assert.Contains(t, content, "## Summary")
}

func TestExportService_ExportPDF_Success(t *testing.T) {
	// GIVEN: Export service and analysis result
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID:  789,
		Mode:      "critical",
		Summary:   "Critical issues found",
		Metadata:  `{"severity":"high"}`,
		ModelUsed: "gpt4",
	}

	// WHEN: Exporting to PDF
	result, err := service.Export(context.Background(), analysis, FormatPDF)

	// THEN: Should succeed
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, FormatPDF, result.Format)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.Contains(t, result.Filename, ".pdf")
}

func TestExportService_Export_NilAnalysis(t *testing.T) {
	// GIVEN: Export service with nil analysis
	service := NewExportService(createTestLogger(t))

	// WHEN: Exporting nil analysis
	result, err := service.Export(context.Background(), nil, FormatJSON)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot export nil analysis")
}

func TestExportService_Export_ContextCancelled(t *testing.T) {
	// GIVEN: Cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{ReviewID: 1, Mode: "skim"}

	// WHEN: Exporting with cancelled context
	result, err := service.Export(ctx, analysis, FormatJSON)

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestExportService_Export_UnsupportedFormat(t *testing.T) {
	// GIVEN: Export service
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{ReviewID: 1, Mode: "skim"}

	// WHEN: Exporting with unsupported format
	result, err := service.Export(context.Background(), analysis, ExportFormat("xml"))

	// THEN: Should return error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestExportService_Filename_Format(t *testing.T) {
	// GIVEN: Export service
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID: 999,
		Mode:     "detailed",
		Summary:  "test",
		Metadata: `{"test":"value"}`,
	}

	// WHEN: Exporting each format
	jsonResult, err1 := service.Export(context.Background(), analysis, FormatJSON)
	mdResult, err2 := service.Export(context.Background(), analysis, FormatMarkdown)
	pdfResult, err3 := service.Export(context.Background(), analysis, FormatPDF)

	// THEN: All should succeed
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	// AND: Filenames should have correct extensions
	assert.NotNil(t, jsonResult)
	assert.NotNil(t, mdResult)
	assert.NotNil(t, pdfResult)
	assert.Contains(t, jsonResult.Filename, "analysis_999_detailed.json")
	assert.Contains(t, mdResult.Filename, "analysis_999_detailed.md")
	assert.Contains(t, pdfResult.Filename, "analysis_999_detailed.pdf")
}

func TestExportService_MimeTypes(t *testing.T) {
	// GIVEN: Export service
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID: 1,
		Mode:     "skim",
		Summary:  "test",
		Metadata: `{"key":"value"}`,
	}

	// WHEN: Exporting each format
	jsonResult, err1 := service.Export(context.Background(), analysis, FormatJSON)
	mdResult, err2 := service.Export(context.Background(), analysis, FormatMarkdown)
	pdfResult, err3 := service.Export(context.Background(), analysis, FormatPDF)

	// THEN: All should succeed
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	// AND: MIME types should be correct
	assert.NotNil(t, jsonResult)
	assert.NotNil(t, mdResult)
	assert.NotNil(t, pdfResult)
	assert.Equal(t, "application/json", jsonResult.MimeType)
	assert.Equal(t, "text/markdown", mdResult.MimeType)
	assert.Equal(t, "application/pdf", pdfResult.MimeType)
}

func TestExportService_ExportResult_Fields(t *testing.T) {
	// GIVEN: Export service
	service := NewExportService(createTestLogger(t))
	analysis := &review_models.AnalysisResult{
		ReviewID: 555,
		Mode:     "scan",
		Summary:  "Scan results",
		Metadata: `{"test":"data"}`,
	}

	// WHEN: Exporting
	result, err := service.Export(context.Background(), analysis, FormatJSON)

	// THEN: All fields should be populated
	require.NoError(t, err)
	assert.Equal(t, int64(555), result.SessionID)
	assert.Equal(t, FormatJSON, result.Format)
	assert.NotNil(t, result.Content)
	assert.NotEmpty(t, result.Filename)
	assert.NotZero(t, result.ExportedAt)
	assert.NotNil(t, result.AnalysisResult)
}
