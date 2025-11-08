package portal_repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SQL query constants
const (
	queryInsertConfig = `
		INSERT INTO portal.llm_configs (
			id, user_id, provider, model_name, api_key_encrypted, api_endpoint,
			is_default, max_tokens, temperature, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	querySelectConfigByID = `
		SELECT id, user_id, provider, model_name, api_key_encrypted, api_endpoint,
		       is_default, max_tokens, temperature, created_at, updated_at
		FROM portal.llm_configs
		WHERE id = $1
	`

	querySelectConfigsByUser = `
		SELECT id, user_id, provider, model_name, api_key_encrypted, api_endpoint,
		       is_default, max_tokens, temperature, created_at, updated_at
		FROM portal.llm_configs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	queryUpdateConfig = `
		UPDATE portal.llm_configs
		SET provider = $2, model_name = $3, api_key_encrypted = $4, api_endpoint = $5,
		    is_default = $6, max_tokens = $7, temperature = $8, updated_at = $9
		WHERE id = $1
	`

	queryDeleteConfig = `DELETE FROM portal.llm_configs WHERE id = $1`

	querySelectDefaultConfig = `
		SELECT id, user_id, provider, model_name, api_key_encrypted, api_endpoint,
		       is_default, max_tokens, temperature, created_at, updated_at
		FROM portal.llm_configs
		WHERE user_id = $1 AND is_default = true
		LIMIT 1
	`

	queryClearDefaults       = `UPDATE portal.llm_configs SET is_default = false WHERE user_id = $1`
	querySetDefault          = `UPDATE portal.llm_configs SET is_default = true WHERE id = $1`
	querySelectAppPreference = `
		SELECT id, user_id, app_name, llm_config_id, created_at, updated_at
		FROM portal.app_llm_preferences
		WHERE user_id = $1 AND app_name = $2
	`

	queryUpsertAppPreference = `
		INSERT INTO portal.app_llm_preferences (user_id, app_name, llm_config_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, app_name)
		DO UPDATE SET llm_config_id = $3, updated_at = $5
	`

	queryDeleteAppPreference = `DELETE FROM portal.app_llm_preferences WHERE user_id = $1 AND app_name = $2`

	querySelectAllAppPreferences = `
		SELECT id, user_id, app_name, llm_config_id, created_at, updated_at
		FROM portal.app_llm_preferences
		WHERE user_id = $1
		ORDER BY app_name
	`
)

// LLMConfig represents a user's LLM configuration
type LLMConfig struct {
	ID              string
	UserID          int
	Provider        string // 'openai', 'anthropic', 'ollama', 'deepseek', 'mistral', 'google'
	ModelName       string
	APIKeyEncrypted sql.NullString // NULL for Ollama (local, no encryption)
	APIEndpoint     sql.NullString // NULL uses provider default
	IsDefault       bool
	MaxTokens       int
	Temperature     float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AppLLMPreference represents an app-specific LLM preference
type AppLLMPreference struct {
	ID          int
	UserID      int
	AppName     string // 'review', 'logs', 'analytics', 'build'
	LLMConfigID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// LLMConfigRepository defines operations for LLM configuration persistence
type LLMConfigRepository interface {
	// CRUD Operations
	Create(ctx context.Context, config *LLMConfig) error
	FindByID(ctx context.Context, id string) (*LLMConfig, error)
	FindByUser(ctx context.Context, userID int) ([]*LLMConfig, error)
	Update(ctx context.Context, config *LLMConfig) error
	Delete(ctx context.Context, id string) error

	// Default Config Management
	FindDefaultByUser(ctx context.Context, userID int) (*LLMConfig, error)
	SetDefault(ctx context.Context, userID int, configID string) error

	// App Preferences
	GetAppPreference(ctx context.Context, userID int, appName string) (*AppLLMPreference, error)
	SetAppPreference(ctx context.Context, userID int, appName string, configID string) error
	ClearAppPreference(ctx context.Context, userID int, appName string) error
	GetAllAppPreferences(ctx context.Context, userID int) ([]*AppLLMPreference, error)
}

// PostgresLLMConfigRepository implements LLMConfigRepository with PostgreSQL
type PostgresLLMConfigRepository struct {
	db *sql.DB
}

// NewLLMConfigRepository creates a new PostgreSQL LLM config repository
func NewLLMConfigRepository(db *sql.DB) LLMConfigRepository {
	return &PostgresLLMConfigRepository{db: db}
}

// Create inserts a new LLM configuration
func (r *PostgresLLMConfigRepository) Create(ctx context.Context, config *LLMConfig) error {
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, queryInsertConfig,
		config.ID, config.UserID, config.Provider, config.ModelName,
		config.APIKeyEncrypted, config.APIEndpoint, config.IsDefault,
		config.MaxTokens, config.Temperature, config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create LLM config for user %d, provider %s: %w",
			config.UserID, config.Provider, err)
	}

	return nil
}

// FindByID retrieves a config by ID
func (r *PostgresLLMConfigRepository) FindByID(ctx context.Context, id string) (*LLMConfig, error) {
	config := &LLMConfig{}
	err := r.db.QueryRowContext(ctx, querySelectConfigByID, id).Scan(
		&config.ID, &config.UserID, &config.Provider, &config.ModelName,
		&config.APIKeyEncrypted, &config.APIEndpoint, &config.IsDefault,
		&config.MaxTokens, &config.Temperature, &config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found = nil, not error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find LLM config by ID %s: %w", id, err)
	}

	return config, nil
}

// FindByUser retrieves all configs for a user
func (r *PostgresLLMConfigRepository) FindByUser(ctx context.Context, userID int) ([]*LLMConfig, error) {
	rows, err := r.db.QueryContext(ctx, querySelectConfigsByUser, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM configs for user %d: %w", userID, err)
	}
	defer rows.Close()

	var configs []*LLMConfig
	for rows.Next() {
		config := &LLMConfig{}
		err := rows.Scan(
			&config.ID, &config.UserID, &config.Provider, &config.ModelName,
			&config.APIKeyEncrypted, &config.APIEndpoint, &config.IsDefault,
			&config.MaxTokens, &config.Temperature, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan LLM config for user %d: %w", userID, err)
		}
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating LLM configs for user %d: %w", userID, err)
	}

	return configs, nil
}

// Update modifies an existing config
func (r *PostgresLLMConfigRepository) Update(ctx context.Context, config *LLMConfig) error {
	config.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, queryUpdateConfig,
		config.ID, config.Provider, config.ModelName, config.APIKeyEncrypted,
		config.APIEndpoint, config.IsDefault, config.MaxTokens,
		config.Temperature, config.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update LLM config %s for user %d: %w",
			config.ID, config.UserID, err)
	}

	return nil
}

// Delete removes a config
func (r *PostgresLLMConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, queryDeleteConfig, id)
	if err != nil {
		return fmt.Errorf("failed to delete LLM config %s: %w", id, err)
	}
	return nil
}

// FindDefaultByUser retrieves the default config for a user
func (r *PostgresLLMConfigRepository) FindDefaultByUser(ctx context.Context, userID int) (*LLMConfig, error) {
	config := &LLMConfig{}
	err := r.db.QueryRowContext(ctx, querySelectDefaultConfig, userID).Scan(
		&config.ID, &config.UserID, &config.Provider, &config.ModelName,
		&config.APIKeyEncrypted, &config.APIEndpoint, &config.IsDefault,
		&config.MaxTokens, &config.Temperature, &config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No default = nil, not error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find default LLM config for user %d: %w", userID, err)
	}

	return config, nil
}

// SetDefault sets a config as the default, clearing other defaults
func (r *PostgresLLMConfigRepository) SetDefault(ctx context.Context, userID int, configID string) error {
	// Use transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for setting default config: %w", err)
	}
	defer tx.Rollback()

	// Clear all defaults for user
	_, err = tx.ExecContext(ctx, queryClearDefaults, userID)
	if err != nil {
		return fmt.Errorf("failed to clear default configs for user %d: %w", userID, err)
	}

	// Set new default
	_, err = tx.ExecContext(ctx, querySetDefault, configID)
	if err != nil {
		return fmt.Errorf("failed to set config %s as default for user %d: %w", configID, userID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit default config transaction: %w", err)
	}

	return nil
}

// GetAppPreference retrieves an app-specific LLM preference
func (r *PostgresLLMConfigRepository) GetAppPreference(ctx context.Context, userID int, appName string) (*AppLLMPreference, error) {
	pref := &AppLLMPreference{}
	err := r.db.QueryRowContext(ctx, querySelectAppPreference, userID, appName).Scan(
		&pref.ID, &pref.UserID, &pref.AppName, &pref.LLMConfigID,
		&pref.CreatedAt, &pref.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No preference = nil, not error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get app preference for user %d, app %s: %w", userID, appName, err)
	}

	return pref, nil
}

// SetAppPreference sets or updates an app preference
func (r *PostgresLLMConfigRepository) SetAppPreference(ctx context.Context, userID int, appName string, configID string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, queryUpsertAppPreference, userID, appName, configID, now, now)
	if err != nil {
		return fmt.Errorf("failed to set app preference for user %d, app %s, config %s: %w",
			userID, appName, configID, err)
	}
	return nil
}

// ClearAppPreference removes an app preference
func (r *PostgresLLMConfigRepository) ClearAppPreference(ctx context.Context, userID int, appName string) error {
	_, err := r.db.ExecContext(ctx, queryDeleteAppPreference, userID, appName)
	if err != nil {
		return fmt.Errorf("failed to clear app preference for user %d, app %s: %w", userID, appName, err)
	}
	return nil
}

// GetAllAppPreferences retrieves all app preferences for a user
// GetAllAppPreferences lists all app preferences for a user
func (r *PostgresLLMConfigRepository) GetAllAppPreferences(ctx context.Context, userID int) ([]*AppLLMPreference, error) {
	rows, err := r.db.QueryContext(ctx, querySelectAllAppPreferences, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query app preferences for user %d: %w", userID, err)
	}
	defer rows.Close()

	var prefs []*AppLLMPreference
	for rows.Next() {
		pref := &AppLLMPreference{}
		err := rows.Scan(&pref.ID, &pref.UserID, &pref.AppName, &pref.LLMConfigID,
			&pref.CreatedAt, &pref.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan app preference for user %d: %w", userID, err)
		}
		prefs = append(prefs, pref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating app preferences for user %d: %w", userID, err)
	}

	return prefs, nil
}
