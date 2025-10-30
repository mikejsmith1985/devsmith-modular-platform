package review_handlers

import (
	"fmt"
	"regexp"
	"strings"
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

// TestGenerateAnalysisID_NonDeterministic tests that IDs are not predictable
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

func TestGenerateAnalysisID_RapidGeneration(t *testing.T) {
	ids := make([]string, 50)
	for i := 0; i < 50; i++ {
		ids[i] = generateAnalysisID()
	}

	// Verify uniqueness
	idMap := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, idMap[id], "Duplicate found in rapid generation")
		idMap[id] = true
	}
}

func TestGenerateAnalysisID_ConcurrentGeneration(t *testing.T) {
	// Test with goroutines
	done := make(chan string, 10)
	for i := 0; i < 10; i++ {
		go func() {
			done <- generateAnalysisID()
		}()
	}

	idMap := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := <-done
		assert.False(t, idMap[id], "Duplicate in concurrent generation")
		idMap[id] = true
	}

	assert.Equal(t, 10, len(idMap))
}

func TestGenerateAnalysisID_Validity(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := generateAnalysisID()
		assert.Equal(t, 36, len(id))
		assert.Contains(t, id, "-")
	}
}

func TestGenerateAnalysisID_HyphenPositions(t *testing.T) {
	id := generateAnalysisID()

	// Check hyphen positions (standard UUID format)
	assert.Equal(t, byte('-'), id[8])
	assert.Equal(t, byte('-'), id[13])
	assert.Equal(t, byte('-'), id[18])
	assert.Equal(t, byte('-'), id[23])
}

func TestGenerateAnalysisID_SequentialUniqueness(t *testing.T) {
	// Generate 5 sequential IDs and ensure all unique
	id1 := generateAnalysisID()
	id2 := generateAnalysisID()
	id3 := generateAnalysisID()
	id4 := generateAnalysisID()
	id5 := generateAnalysisID()

	ids := []string{id1, id2, id3, id4, id5}
	idSet := make(map[string]bool)

	for _, id := range ids {
		assert.False(t, idSet[id], fmt.Sprintf("Duplicate ID: %s", id))
		idSet[id] = true
	}

	assert.Equal(t, 5, len(idSet))
}

func TestGenerateAnalysisID_ReturnString(t *testing.T) {
	id := GenerateAnalysisID()
	assert.NotEmpty(t, id)
	assert.IsType(t, "", id)
}

func TestGenerateAnalysisID_Uniqueness_Batch(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateAnalysisID()
		assert.False(t, ids[id], "Duplicate ID generated: %s", id)
		ids[id] = true
	}
	assert.Equal(t, 100, len(ids))
}

func TestGenerateAnalysisID_ValidUUID(t *testing.T) {
	id := GenerateAnalysisID()
	assert.Len(t, id, 36) // UUID format: 8-4-4-4-12 = 36 chars
	assert.Contains(t, id, "-")
}

func TestGenerateAnalysisID_HexCharacters(t *testing.T) {
	id := GenerateAnalysisID()
	// Remove dashes and verify all chars are valid hex
	hexPart := strings.ReplaceAll(id, "-", "")
	for _, ch := range hexPart {
		assert.True(t, (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f'),
			"Invalid UUID character: %c", ch)
	}
}

func TestGenerateAnalysisID_ExportedFunction(t *testing.T) {
	id := GenerateAnalysisID()
	assert.NotEmpty(t, id)
	assert.Len(t, id, 36)
}

func TestGenerateAnalysisID_RegexValidation(t *testing.T) {
	id := GenerateAnalysisID()
	// Verify UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, err := regexp.MatchString(pattern, id)
	assert.NoError(t, err)
	assert.True(t, matched, "ID does not match UUID format: %s", id)
}
