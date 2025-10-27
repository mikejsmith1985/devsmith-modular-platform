package services

import (
	"context"
	"testing"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzePreview_ReturnsExpectedFields(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	service := NewPreviewService(mockLogger)
	result, err := service.AnalyzePreview(context.Background(), "testdata/sample_project")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.FileTree)
	assert.NotEmpty(t, result.BoundedContexts)
	assert.NotEmpty(t, result.TechStack)
	assert.NotEmpty(t, result.ArchitecturePattern)
	assert.NotEmpty(t, result.EntryPoints)
	assert.NotEmpty(t, result.ExternalDependencies)
	assert.NotEmpty(t, result.Summary)
	assert.Empty(t, result.FunctionImplementations)
}
