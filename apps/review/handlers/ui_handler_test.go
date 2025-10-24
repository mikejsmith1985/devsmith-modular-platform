package handlers

import (
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateAnalysisID tests UUID generation
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

// TestGenerateAnalysisID_UUIDFormat validates UUID structure
func TestGenerateAnalysisID_UUIDFormat(t *testing.T) {
	id := generateAnalysisID()

	// UUID format: 8-4-4-4-12
	// Verify length is correct
	require.Equal(t, 36, len(id))

	// Verify it matches UUID pattern (basic check)
	for i := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			assert.Equal(t, byte('-'), id[i], fmt.Sprintf("Position %d should be hyphen", i))
		}
	}
}

// TestGenerateAnalysisID_Consistency tests ID generation consistency
func TestGenerateAnalysisID_Consistency(t *testing.T) {
	// Generate many IDs and ensure no duplicates
	idSet := make(map[string]bool)
	testCount := 1000

	for i := 0; i < testCount; i++ {
		id := generateAnalysisID()
		assert.False(t, idSet[id], "Duplicate ID found in generation test")
		idSet[id] = true
	}

	assert.Equal(t, testCount, len(idSet))
}

// TestGenerateAnalysisID_Idempotent tests that IDs are not predictable
func TestGenerateAnalysisID_NonDeterministic(t *testing.T) {
	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		ids[i] = generateAnalysisID()
	}

	// All IDs should be different
	idMap := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, idMap[id], "Non-deterministic ID generation failed")
		idMap[id] = true
	}

	assert.Equal(t, 10, len(idMap))
}
