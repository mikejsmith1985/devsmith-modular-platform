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
	result, err := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NoError(t, err)
	assert.NotNil(t, result.FileTree, "Must return file/folder tree")
	assert.NotEmpty(t, result.FileTree, "FileTree should not be empty")
}

func TestPreviewMode_IdentifiesBoundedContexts(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NotEmpty(t, result.BoundedContexts, "Must identify bounded contexts")
}

func TestPreviewMode_DetectsTechStack(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NotEmpty(t, result.TechStack, "Should detect tech stack")
}

func TestPreviewMode_DetectsArchitecturePattern(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NotEmpty(t, result.ArchitecturePattern, "Should identify architecture pattern")
}

func TestPreviewMode_IdentifiesEntryPoints(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NotEmpty(t, result.EntryPoints, "Should identify entry points")
}

func TestPreviewMode_ListsExternalDependencies(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.NotEmpty(t, result.ExternalDependencies, "Should list external dependencies")
}

func TestPreviewMode_ManagesCognitiveLoad(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := review_services.NewPreviewService(mockLogger)
	result, _ := service.AnalyzePreview(context.Background(), "package main\nfunc main() {}")
	assert.Less(t, len(result.Summary), 500, "Summary should be brief (max 500 chars for cognitive load)")
	assert.NotNil(t, result, "Should return valid analysis result")
}
