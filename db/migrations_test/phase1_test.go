package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_PromptTemplates tests the prompt templates migration
func TestMigration_PromptTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()

	// Apply migration
	err := applyMigration(db, "20251108_001_prompt_templates.sql")
	require.NoError(t, err, "Migration should apply successfully")

	// Test 1: Review schema exists
	var schemaExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.schemata 
			WHERE schema_name = 'review'
		)
	`).Scan(&schemaExists)
	require.NoError(t, err)
	assert.True(t, schemaExists, "Review schema should exist")

	// Test 2: prompt_templates table exists with correct columns
	var tableExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'review' 
			AND table_name = 'prompt_templates'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "prompt_templates table should exist")

	// Test 3: Verify column constraints
	columns := []string{
		"id", "user_id", "mode", "user_level", "output_mode",
		"prompt_text", "variables", "is_default", "version",
		"created_at", "updated_at",
	}
	for _, col := range columns {
		var colExists bool
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'review'
				AND table_name = 'prompt_templates'
				AND column_name = $1
			)
		`, col).Scan(&colExists)
		require.NoError(t, err)
		assert.True(t, colExists, fmt.Sprintf("Column %s should exist", col))
	}

	// Test 4: Check mode constraint
	_, err = db.ExecContext(ctx, `
		INSERT INTO review.prompt_templates (id, mode, user_level, output_mode, prompt_text)
		VALUES ('test1', 'invalid_mode', 'beginner', 'quick', 'test')
	`)
	assert.Error(t, err, "Invalid mode should be rejected")

	// Test 5: Check user_level constraint
	_, err = db.ExecContext(ctx, `
		INSERT INTO review.prompt_templates (id, mode, user_level, output_mode, prompt_text)
		VALUES ('test2', 'preview', 'invalid_level', 'quick', 'test')
	`)
	assert.Error(t, err, "Invalid user_level should be rejected")

	// Test 6: Check output_mode constraint
	_, err = db.ExecContext(ctx, `
		INSERT INTO review.prompt_templates (id, mode, user_level, output_mode, prompt_text)
		VALUES ('test3', 'preview', 'beginner', 'invalid_output', 'test')
	`)
	assert.Error(t, err, "Invalid output_mode should be rejected")

	// Test 7: Valid insert should work
	_, err = db.ExecContext(ctx, `
		INSERT INTO review.prompt_templates (id, mode, user_level, output_mode, prompt_text)
		VALUES ('test4', 'preview', 'beginner', 'quick', 'test prompt')
	`)
	assert.NoError(t, err, "Valid insert should succeed")

	// Test 8: prompt_executions table exists
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'review' 
			AND table_name = 'prompt_executions'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "prompt_executions table should exist")

	// Test 9: Indexes created
	indexes := []string{
		"idx_prompt_templates_user",
		"idx_prompt_templates_mode",
		"idx_prompt_templates_default",
		"idx_prompt_executions_user",
		"idx_prompt_executions_template",
		"idx_prompt_executions_model",
	}
	for _, idx := range indexes {
		var idxExists bool
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes
				WHERE schemaname = 'review'
				AND indexname = $1
			)
		`, idx).Scan(&idxExists)
		require.NoError(t, err)
		assert.True(t, idxExists, fmt.Sprintf("Index %s should exist", idx))
	}

	// Test 10: Trigger exists
	var triggerExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger
			WHERE tgname = 'trigger_update_prompt_template_timestamp'
		)
	`).Scan(&triggerExists)
	require.NoError(t, err)
	assert.True(t, triggerExists, "Update timestamp trigger should exist")
}

// TestMigration_LLMConfigs tests the LLM configurations migration
func TestMigration_LLMConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()

	// Apply migration
	err := applyMigration(db, "20251108_002_llm_configs.sql")
	require.NoError(t, err, "Migration should apply successfully")

	// Test 1: llm_configs table exists
	var tableExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'portal' 
			AND table_name = 'llm_configs'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "llm_configs table should exist")

	// Test 2: Check provider constraint
	// First, create a test user
	var testUserID int
	err = db.QueryRowContext(ctx, `
		INSERT INTO portal.users (github_id, github_username, email)
		VALUES (999999, 'testuser', 'test@example.com')
		RETURNING id
	`).Scan(&testUserID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.llm_configs (id, user_id, provider, model_name)
		VALUES ('test1', $1, 'invalid_provider', 'model1')
	`, testUserID)
	assert.Error(t, err, "Invalid provider should be rejected")

	// Test 3: Valid providers accepted
	validProviders := []string{"openai", "anthropic", "ollama", "deepseek", "mistral", "google"}
	for _, provider := range validProviders {
		_, err = db.ExecContext(ctx, `
			INSERT INTO portal.llm_configs (id, user_id, provider, model_name)
			VALUES ($1, $2, $3, $4)
		`, fmt.Sprintf("test_%s", provider), testUserID, provider, "model1")
		assert.NoError(t, err, fmt.Sprintf("Provider %s should be accepted", provider))
	}

	// Test 4: app_llm_preferences table exists
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'portal' 
			AND table_name = 'app_llm_preferences'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "app_llm_preferences table should exist")

	// Test 5: Check app_name constraint
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.app_llm_preferences (user_id, app_name, llm_config_id)
		VALUES ($1, 'invalid_app', 'test_openai')
	`, testUserID)
	assert.Error(t, err, "Invalid app_name should be rejected")

	// Test 6: Valid app names accepted
	validApps := []string{"review", "logs", "analytics", "build"}
	for _, app := range validApps {
		_, err = db.ExecContext(ctx, `
			INSERT INTO portal.app_llm_preferences (user_id, app_name, llm_config_id)
			VALUES ($1, $2, $3)
		`, testUserID, app, "test_openai")
		assert.NoError(t, err, fmt.Sprintf("App %s should be accepted", app))
	}

	// Test 7: llm_usage_logs table exists
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'portal' 
			AND table_name = 'llm_usage_logs'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists, "llm_usage_logs table should exist")

	// Test 8: Usage log insert works
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.llm_usage_logs (user_id, app_name, provider, model_name, tokens_used, latency_ms, cost_usd)
		VALUES ($1, 'review', 'openai', 'gpt-4', 1000, 500, 0.030000)
	`, testUserID)
	assert.NoError(t, err, "Usage log insert should succeed")

	// Test 9: Single default trigger works
	// Insert first config as default
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.llm_configs (id, user_id, provider, model_name, is_default)
		VALUES ('default1', $1, 'openai', 'gpt-4', true)
	`, testUserID)
	require.NoError(t, err)

	// Insert second config as default - should unset first
	_, err = db.ExecContext(ctx, `
		INSERT INTO portal.llm_configs (id, user_id, provider, model_name, is_default)
		VALUES ('default2', $1, 'anthropic', 'claude-3', true)
	`, testUserID)
	require.NoError(t, err)

	// Verify only one default exists
	var defaultCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM portal.llm_configs
		WHERE user_id = $1 AND is_default = true
	`, testUserID).Scan(&defaultCount)
	require.NoError(t, err)
	assert.Equal(t, 1, defaultCount, "Only one default config should exist")
}

// TestSeeds_DefaultPrompts tests the default prompts seed data
func TestSeeds_DefaultPrompts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ctx := context.Background()

	// Apply migrations first
	err := applyMigration(db, "20251108_001_prompt_templates.sql")
	require.NoError(t, err)

	// Apply seed data
	err = applySeed(db, "20251108_001_default_prompts.sql")
	require.NoError(t, err, "Seed data should apply successfully")

	// Test 1: Verify 15 default prompts exist
	var promptCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates WHERE is_default = true
	`).Scan(&promptCount)
	require.NoError(t, err)
	assert.Equal(t, 15, promptCount, "Should have exactly 15 default prompts")

	// Test 2: Verify all modes covered
	modes := []string{"preview", "skim", "scan", "detailed", "critical"}
	for _, mode := range modes {
		var modeCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM review.prompt_templates 
			WHERE mode = $1 AND is_default = true
		`, mode).Scan(&modeCount)
		require.NoError(t, err)
		assert.Equal(t, 3, modeCount, fmt.Sprintf("Mode %s should have 3 prompts (3 user levels)", mode))
	}

	// Test 3: Verify all user levels covered
	userLevels := []string{"beginner", "intermediate", "expert"}
	for _, level := range userLevels {
		var levelCount int
		err = db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM review.prompt_templates 
			WHERE user_level = $1 AND is_default = true
		`, level).Scan(&levelCount)
		require.NoError(t, err)
		assert.Equal(t, 5, levelCount, fmt.Sprintf("User level %s should have 5 prompts (5 modes)", level))
	}

	// Test 4: Verify all prompts have required variables
	rows, err := db.QueryContext(ctx, `
		SELECT id, mode, prompt_text, variables 
		FROM review.prompt_templates 
		WHERE is_default = true
	`)
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var id, mode, promptText, variables string
		err := rows.Scan(&id, &mode, &promptText, &variables)
		require.NoError(t, err)

		assert.NotEmpty(t, promptText, fmt.Sprintf("Prompt %s should have text", id))
		assert.Contains(t, promptText, "{{code}}", fmt.Sprintf("Prompt %s should contain {{code}} variable", id))

		// Scan mode should have {{query}} variable
		if mode == "scan" {
			assert.Contains(t, promptText, "{{query}}", fmt.Sprintf("Scan mode prompt %s should contain {{query}} variable", id))
		}
	}

	// Test 5: Verify all prompts have NULL user_id (system defaults)
	var nonSystemCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review.prompt_templates 
		WHERE is_default = true AND user_id IS NOT NULL
	`).Scan(&nonSystemCount)
	require.NoError(t, err)
	assert.Equal(t, 0, nonSystemCount, "All default prompts should have NULL user_id")
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	// Get connection string from environment or use default
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://devsmith:devsmith@localhost:5432/devsmith_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "Should connect to test database")

	// Ensure clean state - drop and recreate schemas
	ctx := context.Background()
	_, err = db.ExecContext(ctx, "DROP SCHEMA IF EXISTS review CASCADE")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS portal")
	require.NoError(t, err)

	// Create minimal portal.users table for foreign key tests
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS portal.users (
			id SERIAL PRIMARY KEY,
			github_id INT UNIQUE NOT NULL,
			github_username VARCHAR(255) NOT NULL,
			email VARCHAR(255)
		)
	`)
	require.NoError(t, err)

	return db
}

func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

func applyMigration(db *sql.DB, filename string) error {
	content, err := os.ReadFile(fmt.Sprintf("../migrations/%s", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(content))
	return err
}

func applySeed(db *sql.DB, filename string) error {
	content, err := os.ReadFile(fmt.Sprintf("../seeds/%s", filename))
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	_, err = db.Exec(string(content))
	return err
}
