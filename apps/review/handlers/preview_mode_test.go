// Package review_handlers contains HTTP handlers for the review app.
package review_handlers

import (
	"context"
	"testing"

	review_services "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPreviewMode_ReturnsFileStructure(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, err := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.NoError(t, err)
	assert.NotNil(t, result.FileTree, "Must return file/folder tree")
	assert.NotEmpty(t, result.FileTree[0].Description, "Each node has description")
}

func TestPreviewMode_IdentifiesBoundedContexts(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.NotEmpty(t, result.BoundedContexts, "Must identify bounded contexts")
}

func TestPreviewMode_DetectsTechStack(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.Contains(t, result.TechStack, "Go", "Should detect Go stack")
}

func TestPreviewMode_DetectsArchitecturePattern(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.Contains(t, result.ArchitecturePattern, "layered", "Should identify architecture pattern")
}

func TestPreviewMode_IdentifiesEntryPoints(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.NotEmpty(t, result.EntryPoints, "Should identify entry points")
}

func TestPreviewMode_ListsExternalDependencies(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.NotEmpty(t, result.ExternalDependencies, "Should list external dependencies")
}

func TestPreviewMode_ManagesCognitiveLoad(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.Less(t, len(result.Summary), 500, "Summary should be brief")
	assert.Empty(t, result.FunctionImplementations, "Preview mode skips implementation details")
}
