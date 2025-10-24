package handlers

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAnalysisID(t *testing.T) {
	id1 := generateAnalysisID()
	id2 := generateAnalysisID()

	// Should not be empty
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)

	// Should be unique
	assert.NotEqual(t, id1, id2)

	// Should be valid UUID format (contains hyphens)
	assert.Contains(t, id1, "-")
	assert.Contains(t, id2, "-")

	// Should have correct UUID structure (8-4-4-4-12)
	assert.Equal(t, 36, len(id1))
	assert.Equal(t, 36, len(id2))
}

func TestGenerateAnalysisID_MultipleCallsProduceDifferentIDs(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateAnalysisID()
		assert.False(t, ids[id], "Duplicate ID generated")
		ids[id] = true
	}
	assert.Equal(t, 100, len(ids))
}
