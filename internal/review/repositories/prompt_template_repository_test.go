package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	review_models "github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Should connect to test database")

	// Clean state
	ctx := context.Background()
	_, err = db.ExecContext(ctx, "DELETE FROM review.prompt_executions")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM review.prompt_templates WHERE user_id IS NOT NULL")
	require.NoError(t, err)

	// Create test users in portal.users for foreign key constraints
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.users (id, github_id, github_username, email)
		VALUES 
			(123, 123000, 'testuser123', 'test123@example.com'),
			(456, 456000, 'testuser456', 'test456@example.com'),
			(789, 789000, 'testuser789', 'test789@example.com'),
			(111, 111000, 'testuser111', 'test111@example.com'),
			(222, 222000, 'testuser222', 'test222@example.com'),
			(333, 333000, 'testuser333', 'test333@example.com')
		ON CONFLICT (id) DO NOTHING
	`)
	require.NoError(t, err)

	return db
}

func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// RED PHASE: Tests that should fail because repository doesn't exist yet

func TestPromptTemplateRepository_FindByUserAndMode_UserCustom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Insert a user custom prompt
	customPrompt := &review_models.PromptTemplate{
		ID:         "user_123_preview_beginner_quick",
		UserID:     intPtr(123),
		Mode:       "preview",
		UserLevel:  "beginner",
		OutputMode: "quick",
		PromptText: "Custom preview prompt for user 123",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
	}

	_, err := db.ExecContext(ctx, `
		INSERT INTO review.prompt_templates 
		(id, user_id, mode, user_level, output_mode, prompt_text, variables, is_default, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, customPrompt.ID, customPrompt.UserID, customPrompt.Mode, customPrompt.UserLevel,
		customPrompt.OutputMode, customPrompt.PromptText, `["{{code}}"]`, customPrompt.IsDefault, customPrompt.Version)
	require.NoError(t, err)

	// Test: Should find user's custom prompt
	result, err := repo.FindByUserAndMode(ctx, 123, "preview", "beginner", "quick")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user_123_preview_beginner_quick", result.ID)
	assert.Equal(t, intPtr(123), result.UserID)
	assert.Equal(t, "Custom preview prompt for user 123", result.PromptText)
}

func TestPromptTemplateRepository_FindByUserAndMode_NoCustom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Test: Should return nil when user has no custom prompt
	result, err := repo.FindByUserAndMode(ctx, 999, "preview", "beginner", "quick")
	assert.NoError(t, err)
	assert.Nil(t, result, "Should return nil when no user custom exists")
}

func TestPromptTemplateRepository_FindDefaultByMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Test: Should find system default prompt (seeded in Phase 1)
	result, err := repo.FindDefaultByMode(ctx, "preview", "beginner", "quick")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "default_preview_beginner_quick", result.ID)
	assert.Nil(t, result.UserID, "Default prompts have NULL user_id")
	assert.True(t, result.IsDefault)
}

func TestPromptTemplateRepository_Upsert_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Test: Create new custom prompt
	newPrompt := &review_models.PromptTemplate{
		ID:         "user_456_skim_intermediate_quick",
		UserID:     intPtr(456),
		Mode:       "skim",
		UserLevel:  "intermediate",
		OutputMode: "quick",
		PromptText: "My custom skim prompt",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
	}

	result, err := repo.Upsert(ctx, newPrompt)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user_456_skim_intermediate_quick", result.ID)

	// Verify it was actually inserted
	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates 
		WHERE id = $1
	`, result.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPromptTemplateRepository_Upsert_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Insert initial prompt
	initialPrompt := &review_models.PromptTemplate{
		ID:         "user_789_scan_expert_quick",
		UserID:     intPtr(789),
		Mode:       "scan",
		UserLevel:  "expert",
		OutputMode: "quick",
		PromptText: "Original scan prompt",
		Variables:  []string{"{{code}}", "{{query}}"},
		IsDefault:  false,
		Version:    1,
	}
	_, err := repo.Upsert(ctx, initialPrompt)
	require.NoError(t, err)

	// Test: Update existing prompt
	initialPrompt.PromptText = "Updated scan prompt"
	initialPrompt.Version = 2

	result, err := repo.Upsert(ctx, initialPrompt)
	require.NoError(t, err)
	assert.Equal(t, "Updated scan prompt", result.PromptText)
	assert.Equal(t, 2, result.Version)

	// Verify only one record exists
	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates 
		WHERE user_id = 789 AND mode = 'scan'
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should update, not create duplicate")
}

func TestPromptTemplateRepository_DeleteUserCustom(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Insert custom prompt
	customPrompt := &review_models.PromptTemplate{
		ID:         "user_111_detailed_beginner_quick",
		UserID:     intPtr(111),
		Mode:       "detailed",
		UserLevel:  "beginner",
		OutputMode: "quick",
		PromptText: "Custom detailed prompt",
		Variables:  []string{"{{code}}"},
		IsDefault:  false,
		Version:    1,
	}
	_, err := repo.Upsert(ctx, customPrompt)
	require.NoError(t, err)

	// Test: Delete user's custom prompt
	err = repo.DeleteUserCustom(ctx, 111, "detailed", "beginner", "quick")
	require.NoError(t, err)

	// Verify it's deleted
	var count int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates 
		WHERE user_id = 111 AND mode = 'detailed'
	`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify system defaults are not affected
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates 
		WHERE user_id IS NULL AND is_default = true
	`).Scan(&count)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "System defaults should remain")
}

func TestPromptTemplateRepository_SaveExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// Test: Save prompt execution
	execution := &review_models.PromptExecution{
		TemplateID:     "default_preview_beginner_quick",
		UserID:         222,
		RenderedPrompt: "Analyze this code: function test() {}",
		Response:       "This is a JavaScript function...",
		ModelUsed:      "claude-3-5-sonnet-20241022",
		LatencyMs:      1234,
		TokensUsed:     567,
		UserRating:     intPtr(5),
	}

	err := repo.SaveExecution(ctx, execution)
	require.NoError(t, err)
	assert.NotZero(t, execution.ID, "Should assign ID after insert")
	assert.NotZero(t, execution.CreatedAt, "Should set created_at timestamp")
}

func TestPromptTemplateRepository_GetExecutionHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := NewPromptTemplateRepository(db)
	ctx := context.Background()

	// First create prompt templates that can be referenced
	templateIDs := make([]string, 5)
	userID := 333
	for i := 0; i < 5; i++ {
		template := &review_models.PromptTemplate{
			UserID:     &userID,
			Mode:       "preview",
			UserLevel:  "beginner",
			OutputMode: "quick",
			PromptText: fmt.Sprintf("Test prompt %d: {{.code}}", i),
			Variables:  []string{"code"},
		}
		saved, err := repo.Upsert(ctx, template)
		require.NoError(t, err)
		templateIDs[i] = saved.ID
	}

	// Insert multiple executions for user 333
	for i := 0; i < 5; i++ {
		execution := &review_models.PromptExecution{
			TemplateID:     templateIDs[i],
			UserID:         333,
			RenderedPrompt: fmt.Sprintf("Prompt %d", i),
			Response:       fmt.Sprintf("Response %d", i),
			ModelUsed:      "claude-3-5-sonnet-20241022",
			LatencyMs:      1000 + i*100,
			TokensUsed:     500 + i*50,
		}
		err := repo.SaveExecution(ctx, execution)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test: Get execution history
	history, err := repo.GetExecutionHistory(ctx, 333, 3)
	require.NoError(t, err)
	assert.Len(t, history, 3, "Should return only 3 most recent")

	// Verify order (most recent first)
	assert.Equal(t, templateIDs[4], history[0].TemplateID)
	assert.Equal(t, templateIDs[3], history[1].TemplateID)
	assert.Equal(t, templateIDs[2], history[2].TemplateID)
}

// Helper function
func intPtr(i int) *int {
	return &i
}
