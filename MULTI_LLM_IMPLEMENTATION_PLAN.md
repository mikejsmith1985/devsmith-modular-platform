# DevSmith Multi-LLM Platform & Prompt Customization - Implementation Plan

**Document Version:** 1.1  
**Created:** 2025-11-08  
**Last Updated:** 2025-11-08  
**Status:** Implementation Phase - 15% Complete (3/20 days)

---

## 📊 Progress Tracker

| Phase | Tasks | Status | Completion |
|-------|-------|--------|------------|
| **Phase 1: Database Schema & Migrations** | 3/3 | ✅ Complete | 100% |
| **Phase 2: Backend Services - Prompt Management** | 2/3 | 🔄 In Progress | 67% |
| **Phase 3: Multi-LLM Infrastructure** | 0/4 | ⏳ Pending | 0% |
| **Phase 4: Frontend Implementation** | 0/3 | ⏳ Pending | 0% |
| **Phase 5: Integration & Testing** | 0/2 | ⏳ Pending | 0% |
| **TOTAL** | 5/15 | 🔄 In Progress | 33% |

**Current Task:** Task 2.3 - Prompt API Endpoints (RED phase)

---

## 📋 Overview

This document outlines the complete implementation of two major features for the DevSmith Modular Platform:

1. **Prompt Transparency & Customization** - Users can view, edit, save, and reset AI prompts
2. **Multi-LLM Platform Architecture** - Support for multiple AI providers (Anthropic, OpenAI, DeepSeek, Mistral, Ollama) with per-app model selection

---

## 🎯 Project Goals

### Primary Objectives
- ✅ Enable users to view and customize AI prompts for all review modes
- ✅ Support multiple LLM providers (API-based and local)
- ✅ Secure API key management with encryption
- ✅ Per-app LLM preferences (Review uses Claude, Logs uses DeepSeek, etc.)
- ✅ Usage tracking and cost monitoring
- ✅ Factory reset capability for prompts
- ✅ Portal UI for managing AI configurations without touching DB/config files

### Testing Requirements
- ✅ **TDD Approach:** RED → GREEN → REFACTOR for all features
- ✅ **Unit Tests:** 70% minimum coverage, 90% for critical paths
- ✅ **Integration Tests:** All cross-service flows
- ✅ **E2E Tests:** Playwright + Percy for visual + functional validation
- ✅ **Manual Testing:** Claude API integration (user will manually enter API key)
- ✅ **Mock Testing:** All other API providers tested with mock data
- ⚠️ **NO hardcoded values, stubs, or mocks that could cause production issues**

---

## 🏗️ Architecture Overview

### Data Models

#### 1. Prompt Templates
```go
type PromptTemplate struct {
    ID          string    `json:"id" db:"id"`
    UserID      *int      `json:"user_id,omitempty" db:"user_id"` // NULL = system default
    Mode        string    `json:"mode" db:"mode"` // "preview", "skim", "scan", "detailed", "critical"
    UserLevel   string    `json:"user_level" db:"user_level"` // "beginner", "intermediate", "expert"
    OutputMode  string    `json:"output_mode" db:"output_mode"` // "quick", "detailed", "comprehensive"
    PromptText  string    `json:"prompt_text" db:"prompt_text"` // The actual prompt template
    Variables   []string  `json:"variables" db:"variables"` // ["{{code}}", "{{query}}", etc.]
    IsDefault   bool      `json:"is_default" db:"is_default"` // Factory default flag
    Version     int       `json:"version" db:"version"` // For versioning
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PromptExecution struct {
    ID             string    `json:"id" db:"id"`
    TemplateID     string    `json:"template_id" db:"template_id"`
    UserID         int       `json:"user_id" db:"user_id"`
    RenderedPrompt string    `json:"rendered_prompt" db:"rendered_prompt"`
    Response       string    `json:"response" db:"response"`
    ModelUsed      string    `json:"model_used" db:"model_used"`
    LatencyMs      int       `json:"latency_ms" db:"latency_ms"`
    TokensUsed     int       `json:"tokens_used" db:"tokens_used"`
    UserRating     *int      `json:"user_rating,omitempty" db:"user_rating"` // 1-5 stars
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
```

#### 2. LLM Configurations
```go
type LLMProvider string

const (
    ProviderOpenAI    LLMProvider = "openai"
    ProviderAnthropic LLMProvider = "anthropic"
    ProviderOllama    LLMProvider = "ollama"
    ProviderDeepSeek  LLMProvider = "deepseek"
    ProviderMistral   LLMProvider = "mistral"
    ProviderGoogle    LLMProvider = "google"
)

type LLMConfig struct {
    ID          string      `json:"id" db:"id"`
    UserID      int         `json:"user_id" db:"user_id"`
    Provider    LLMProvider `json:"provider" db:"provider"`
    ModelName   string      `json:"model_name" db:"model_name"`
    APIKey      string      `json:"-" db:"api_key_encrypted"` // NEVER return in JSON
    APIEndpoint string      `json:"api_endpoint" db:"api_endpoint"`
    IsDefault   bool        `json:"is_default" db:"is_default"`
    MaxTokens   int         `json:"max_tokens" db:"max_tokens"`
    Temperature float64     `json:"temperature" db:"temperature"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type AppLLMPreference struct {
    ID          string    `json:"id" db:"id"`
    UserID      int       `json:"user_id" db:"user_id"`
    AppName     string    `json:"app_name" db:"app_name"` // "review", "logs", "analytics"
    LLMConfigID string    `json:"llm_config_id" db:"llm_config_id"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type LLMUsageLog struct {
    ID         string      `json:"id" db:"id"`
    UserID     int         `json:"user_id" db:"user_id"`
    AppName    string      `json:"app_name" db:"app_name"`
    Provider   LLMProvider `json:"provider" db:"provider"`
    ModelName  string      `json:"model_name" db:"model_name"`
    TokensUsed int         `json:"tokens_used" db:"tokens_used"`
    LatencyMs  int64       `json:"latency_ms" db:"latency_ms"`
    CostUSD    float64     `json:"cost_usd" db:"cost_usd"`
    Success    bool        `json:"success" db:"success"`
    CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}
```

#### 3. Unified AI Client Interface
```go
type AIClient interface {
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    StreamGenerate(ctx context.Context, req *GenerateRequest) (<-chan *StreamChunk, error)
    GetModelInfo() *ModelInfo
}

type GenerateRequest struct {
    Prompt      string            `json:"prompt"`
    MaxTokens   int               `json:"max_tokens"`
    Temperature float64           `json:"temperature"`
    StopTokens  []string          `json:"stop_tokens,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type GenerateResponse struct {
    Text         string      `json:"text"`
    TokensUsed   int         `json:"tokens_used"`
    FinishReason string      `json:"finish_reason"`
    LatencyMs    int64       `json:"latency_ms"`
    Model        string      `json:"model"`
    Provider     LLMProvider `json:"provider"`
}
```

---

## 📅 Implementation Phases

### Phase 1: Database Schema & Migrations (Days 1-2)

#### Task 1.1: Create Prompt Templates Schema
- **File:** `db/migrations/20251108_001_prompt_templates.sql`
- **TDD Steps:**
  1. RED: Write migration test that expects tables to exist
  2. GREEN: Create migration with tables
  3. REFACTOR: Add indexes, constraints, optimize

**Schema:**
```sql
CREATE SCHEMA IF NOT EXISTS review;

CREATE TABLE review.prompt_templates (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT REFERENCES portal.users(id) ON DELETE CASCADE,
    mode VARCHAR(20) NOT NULL CHECK (mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    user_level VARCHAR(20) NOT NULL CHECK (user_level IN ('beginner', 'intermediate', 'expert')),
    output_mode VARCHAR(20) NOT NULL CHECK (output_mode IN ('quick', 'detailed', 'comprehensive')),
    prompt_text TEXT NOT NULL,
    variables JSONB DEFAULT '[]'::jsonb,
    is_default BOOLEAN DEFAULT false,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, mode, user_level, output_mode)
);

CREATE INDEX idx_prompt_templates_user ON review.prompt_templates(user_id);
CREATE INDEX idx_prompt_templates_mode ON review.prompt_templates(mode, user_level, output_mode);

CREATE TABLE review.prompt_executions (
    id SERIAL PRIMARY KEY,
    template_id VARCHAR(64) REFERENCES review.prompt_templates(id) ON DELETE SET NULL,
    user_id INT NOT NULL,
    rendered_prompt TEXT NOT NULL,
    response TEXT,
    model_used VARCHAR(100) NOT NULL,
    latency_ms INT,
    tokens_used INT,
    user_rating INT CHECK (user_rating BETWEEN 1 AND 5),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_prompt_executions_user ON review.prompt_executions(user_id, created_at DESC);
CREATE INDEX idx_prompt_executions_template ON review.prompt_executions(template_id, created_at DESC);
```

**Tests:**
- ✅ Migration applies successfully
- ✅ Migration rolls back cleanly
- ✅ All constraints enforced (mode, user_level, output_mode enums)
- ✅ Foreign keys work correctly
- ✅ Indexes created successfully

---

#### Task 1.2: Create LLM Configuration Schema
- **File:** `db/migrations/20251108_002_llm_configs.sql`

**Schema:**
```sql
CREATE TABLE portal.llm_configs (
    id VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('openai', 'anthropic', 'ollama', 'deepseek', 'mistral', 'google')),
    model_name VARCHAR(100) NOT NULL,
    api_key_encrypted TEXT,
    api_endpoint VARCHAR(255),
    is_default BOOLEAN DEFAULT false,
    max_tokens INT DEFAULT 4096,
    temperature DECIMAL(3,2) DEFAULT 0.7,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, provider, model_name)
);

CREATE INDEX idx_llm_configs_user ON portal.llm_configs(user_id);
CREATE INDEX idx_llm_configs_default ON portal.llm_configs(user_id, is_default) WHERE is_default = true;

CREATE TABLE portal.app_llm_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES portal.users(id) ON DELETE CASCADE,
    app_name VARCHAR(50) NOT NULL CHECK (app_name IN ('review', 'logs', 'analytics', 'build')),
    llm_config_id VARCHAR(64) REFERENCES portal.llm_configs(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, app_name)
);

CREATE INDEX idx_app_llm_prefs_user ON portal.app_llm_preferences(user_id, app_name);

CREATE TABLE portal.llm_usage_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    app_name VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    tokens_used INT NOT NULL,
    latency_ms INT NOT NULL,
    cost_usd DECIMAL(10,4) DEFAULT 0.0000,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_llm_usage_user_date ON portal.llm_usage_logs(user_id, created_at DESC);
CREATE INDEX idx_llm_usage_app ON portal.llm_usage_logs(app_name, created_at DESC);
```

**Tests:**
- ✅ All tables created successfully
- ✅ Provider enum validation works
- ✅ App name enum validation works
- ✅ Foreign key constraints enforced
- ✅ Unique constraints prevent duplicate configs

---

#### Task 1.3: Seed Default Prompts
- **File:** `db/seeds/20251108_001_default_prompts.sql`

**TDD Steps:**
1. RED: Test expects 15 default prompts (5 modes × 3 user levels)
2. GREEN: Insert all default prompts
3. REFACTOR: Optimize prompt text, ensure variables present

**Tests:**
- ✅ 15 default prompts inserted (5 modes × 3 user levels, using "quick" output mode as default)
- ✅ Each prompt contains required variables
- ✅ No user_id (system defaults)
- ✅ is_default flag set to true

---

### Phase 2: Backend Services - Prompt Management (Days 3-5)

#### Task 2.1: Prompt Template Repository ✅ **COMPLETE**
- **File:** `internal/review/repositories/prompt_template_repository.go`

**TDD Steps:**
1. ✅ RED: Write tests for FindByUserAndMode, FindDefaultByMode, Upsert, Delete, SaveExecution, GetExecutionHistory (8 tests)
2. ✅ GREEN: Implement repository methods (all 8 tests passing in 0.174s)
3. ✅ REFACTOR: Optimized queries, extracted common patterns, improved maintainability

**Completed Implementation:**
- ✅ FindByUserAndMode() - retrieves user custom prompts for specific mode/level/output
- ✅ FindDefaultByMode() - retrieves system defaults when no custom prompt exists
- ✅ Upsert() - creates or updates prompts using ON CONFLICT (user_id, mode, user_level, output_mode)
- ✅ DeleteUserCustom() - removes only user customizations, protects system defaults
- ✅ SaveExecution() - logs prompt execution with latency, tokens, model used
- ✅ GetExecutionHistory() - retrieves execution log ordered by created_at DESC

**Refactoring Improvements:**
- ✅ Extracted SQL query constants for better maintainability
- ✅ Created `scanPromptTemplate()` helper to reduce code duplication
- ✅ Simplified all methods to use query constants
- ✅ Improved error messages for better debugging
- ✅ Code reduced from 294 lines to ~220 lines with better organization

**Test Results:**
```
PASS: TestPromptTemplateRepository_FindByUserAndMode_UserCustom (0.01s)
PASS: TestPromptTemplateRepository_FindByUserAndMode_NoCustom (0.01s)
PASS: TestPromptTemplateRepository_FindDefaultByMode (0.01s)
PASS: TestPromptTemplateRepository_Upsert_Create (0.01s)
PASS: TestPromptTemplateRepository_Upsert_Update (0.02s)
PASS: TestPromptTemplateRepository_DeleteUserCustom (0.01s)
PASS: TestPromptTemplateRepository_SaveExecution (0.01s)
PASS: TestPromptTemplateRepository_GetExecutionHistory (0.09s)
```

**Methods:**
```go
type PromptTemplateRepository interface {
    FindByUserAndMode(ctx context.Context, userID int, mode, userLevel, outputMode string) (*models.PromptTemplate, error)
    FindDefaultByMode(ctx context.Context, mode, userLevel, outputMode string) (*models.PromptTemplate, error)
    Upsert(ctx context.Context, template *models.PromptTemplate) (*models.PromptTemplate, error)
    DeleteUserCustom(ctx context.Context, userID int, mode, userLevel, outputMode string) error
    SaveExecution(ctx context.Context, execution *models.PromptExecution) error
    GetExecutionHistory(ctx context.Context, userID int, limit int) ([]*models.PromptExecution, error)
}
```

**Tests:**
- ✅ FindByUserAndMode returns user custom if exists, nil if not
- ✅ FindDefaultByMode returns system default
- ✅ Upsert creates new template if doesn't exist
- ✅ Upsert updates existing template
- ✅ DeleteUserCustom removes only user's custom, not system default
- ✅ SaveExecution logs prompt usage
- ✅ GetExecutionHistory returns latest executions

**Status:** ✅ **TASK 2.1 COMPLETE** (RED → GREEN → REFACTOR cycle complete, all tests passing)

---

#### Task 2.2: Prompt Template Service ✅ **COMPLETE**
- **File:** `internal/review/services/prompt_template_service.go`
- **Test File:** `internal/review/services/prompt_template_service_test.go`
- **Commit:** ca92fb7

**TDD Completion:**
- ✅ **RED Phase:** 14 comprehensive test cases written
- ✅ **GREEN Phase:** All 6 methods implemented, 14/14 tests passing
- ✅ **REFACTOR Phase:** Constants extracted, godoc added, helper method created

**Implemented Methods:**
1. `GetEffectivePrompt` - Returns user custom or falls back to system default
2. `SaveCustomPrompt` - Validates variables and creates/updates custom prompts
3. `FactoryReset` - Deletes user customizations
4. `RenderPrompt` - Substitutes variables in templates
5. `LogExecution` - Records prompt usage with validation
6. `ExtractVariables` - Regex-based variable extraction (deduplicated)

**Test Coverage:**
- 93.3% coverage for SaveCustomPrompt (critical path)
- 100% coverage for GetEffectivePrompt, FactoryReset, RenderPrompt, ExtractVariables
- 71.4% coverage for LogExecution
- Used ElementsMatch for non-deterministic map ordering
- MockPromptTemplateRepository for isolation testing

**Tests:**
- ✅ GetEffectivePrompt returns user custom over system default
- ✅ GetEffectivePrompt falls back to system default if no custom
- ✅ GetEffectivePrompt errors if no default exists
- ✅ SaveCustomPrompt validates required variables ({{code}} for all, {{query}} for scan)
- ✅ SaveCustomPrompt creates unique ID per user/mode combo
- ✅ SaveCustomPrompt success creates template
- ✅ FactoryReset deletes user custom, leaves system default intact
- ✅ FactoryReset handles delete errors
- ✅ RenderPrompt substitutes all variables correctly
- ✅ RenderPrompt errors if variable missing
- ✅ RenderPrompt handles empty variables
- ✅ LogExecution records prompt usage
- ✅ LogExecution validates required fields (template_id, user_id, model_used)
- ✅ ExtractVariables finds single, multiple, duplicate, and no variables

**Status:** ✅ **TASK 2.2 COMPLETE** (RED → GREEN → REFACTOR cycle complete, 14/14 tests passing)

---

#### Task 2.3: Prompt API Endpoints 🔄 **NEXT**
- **File:** `internal/review/handlers/prompt_handler.go`

**Endpoints:**
```
GET    /api/review/prompts?mode={mode}&user_level={level}&output_mode={output}
PUT    /api/review/prompts
DELETE /api/review/prompts?mode={mode}&user_level={level}&output_mode={output}
GET    /api/review/prompts/history?limit=50
POST   /api/review/prompts/{execution_id}/rate
```

**Tests:**
- ✅ GET returns effective prompt (user custom or default)
- ✅ GET includes metadata (is_custom, can_reset)
- ✅ PUT validates prompt text contains required variables
- ✅ PUT creates/updates user custom prompt
- ✅ DELETE resets to factory default
- ✅ DELETE returns error if already using default
- ✅ GET history returns user's recent prompt executions
- ✅ POST rate updates execution rating (1-5 stars)
- ✅ All endpoints require authentication
- ✅ All endpoints return proper error codes

---

### Phase 3: Backend Services - Multi-LLM Infrastructure (Days 6-10)

#### Task 3.1: Encryption Service
- **File:** `internal/portal/services/encryption_service.go`

**TDD Steps:**
1. RED: Test encrypt/decrypt round-trip with user-specific keys
2. GREEN: Implement AES-256-GCM encryption with Argon2 key derivation
3. REFACTOR: Add key rotation support

**Methods:**
```go
type EncryptionService interface {
    EncryptAPIKey(apiKey string, userID int) (string, error)
    DecryptAPIKey(encrypted string, userID int) (string, error)
    ValidateMasterKey() error
}
```

**Tests:**
- ✅ EncryptAPIKey produces different ciphertext for same key (nonce randomness)
- ✅ DecryptAPIKey successfully decrypts encrypted key
- ✅ Decrypt fails with wrong user ID (different derived key)
- ✅ Decrypt fails with corrupted ciphertext
- ✅ ValidateMasterKey checks ENCRYPTION_MASTER_KEY env var present
- ✅ User-specific salt ensures different encryption per user

---

#### Task 3.2: AI Client Interface & Implementations
- **Files:**
  - `internal/ai/client.go` (interface)
  - `internal/ai/anthropic_client.go`
  - `internal/ai/openai_client.go`
  - `internal/ai/ollama_client.go`
  - `internal/ai/deepseek_client.go`
  - `internal/ai/mistral_client.go`

**TDD Steps:**
1. RED: Write tests for each provider (mock HTTP responses)
2. GREEN: Implement clients
3. REFACTOR: Extract common logic, add retries

**Anthropic Client (Real API - Manual Testing):**
- ⚠️ Will be manually tested with user's Claude API key
- ✅ Unit tests use mock HTTP server
- ✅ Integration test marked as manual

**All Other Clients (Mock Testing):**
- ✅ OpenAI client tested with mock responses
- ✅ Ollama client tested with mock local server
- ✅ DeepSeek client tested with mock API
- ✅ Mistral client tested with mock API

**Tests per Client:**
- ✅ Generate returns response with text
- ✅ Generate includes token count
- ✅ Generate includes latency measurement
- ✅ Generate handles API errors gracefully
- ✅ Generate respects max_tokens limit
- ✅ Generate respects temperature setting
- ✅ GetModelInfo returns correct metadata
- ✅ HTTP timeout prevents hanging
- ✅ Retry logic for transient failures (3 attempts)

---

#### Task 3.3: AI Client Factory
- **File:** `internal/ai/factory.go`

**TDD Steps:**
1. RED: Test GetClientForApp returns correct provider
2. GREEN: Implement factory with fallback chain
3. REFACTOR: Add health checks, caching

**Methods:**
```go
type AIClientFactory interface {
    GetClientForApp(ctx context.Context, userID int, appName string) (AIClient, error)
    GetClientWithFallback(ctx context.Context, userID int, appName string) AIClient
    RefreshConfig(ctx context.Context, userID int) error
}
```

**Tests:**
- ✅ GetClientForApp returns user's app-specific preference
- ✅ GetClientForApp falls back to user default if no app preference
- ✅ GetClientForApp falls back to system default (Ollama) if no config
- ✅ GetClientWithFallback tries primary, then default, then Ollama
- ✅ Factory caches clients per user (avoid recreating on every request)
- ✅ RefreshConfig clears cache after config update
- ✅ Decrypts API keys correctly
- ✅ Health check pings provider before returning

---

#### Task 3.4: LLM Configuration Repository
- **File:** `internal/portal/repositories/llm_config_repository.go`

**Methods:**
```go
type LLMConfigRepository interface {
    Create(ctx context.Context, config *models.LLMConfig) error
    Update(ctx context.Context, config *models.LLMConfig) error
    Delete(ctx context.Context, configID string) error
    FindByID(ctx context.Context, configID string) (*models.LLMConfig, error)
    FindByUser(ctx context.Context, userID int) ([]*models.LLMConfig, error)
    FindDefaultByUser(ctx context.Context, userID int) (*models.LLMConfig, error)
    SetDefault(ctx context.Context, userID int, configID string) error
    SaveAppPreference(ctx context.Context, pref *models.AppLLMPreference) error
    GetAppPreference(ctx context.Context, userID int, appName string) (*models.AppLLMPreference, error)
    GetAllAppPreferences(ctx context.Context, userID int) (map[string]*models.AppLLMPreference, error)
    LogUsage(ctx context.Context, log *models.LLMUsageLog) error
    GetUsageSummary(ctx context.Context, userID int, period string) (*UsageSummary, error)
}
```

**Tests:**
- ✅ Create inserts new LLM config
- ✅ Create enforces unique constraint (user, provider, model)
- ✅ Update modifies existing config
- ✅ Delete removes config
- ✅ FindByUser returns all user's configs
- ✅ FindDefaultByUser returns default config
- ✅ SetDefault clears old default, sets new one
- ✅ SaveAppPreference creates/updates preference
- ✅ GetAppPreference returns correct config for app
- ✅ GetAllAppPreferences returns map of all apps
- ✅ LogUsage records token usage
- ✅ GetUsageSummary calculates totals by app/provider

---

#### Task 3.5: LLM Configuration Service
- **File:** `internal/portal/services/llm_config_service.go`

**Methods:**
```go
type LLMConfigService interface {
    CreateConfig(ctx context.Context, userID int, req *CreateLLMConfigRequest) (*models.LLMConfig, error)
    UpdateConfig(ctx context.Context, userID int, configID string, req *UpdateLLMConfigRequest) error
    DeleteConfig(ctx context.Context, userID int, configID string) error
    GetUserConfigs(ctx context.Context, userID int) ([]*models.LLMConfig, error)
    SetAppPreference(ctx context.Context, userID int, appName, configID string) error
    GetAppPreferences(ctx context.Context, userID int) (map[string]*models.LLMConfig, error)
    GetUsageSummary(ctx context.Context, userID int, period string) (*UsageSummary, error)
}
```

**Tests:**
- ✅ CreateConfig encrypts API key before storage
- ✅ CreateConfig validates provider and model
- ✅ CreateConfig sets is_default if first config
- ✅ UpdateConfig re-encrypts API key if changed
- ✅ DeleteConfig prevents deleting if in use by app
- ✅ GetUserConfigs never returns decrypted API keys
- ✅ SetAppPreference validates app name
- ✅ GetAppPreferences returns effective config per app
- ✅ GetUsageSummary calculates costs correctly

---

#### Task 3.6: LLM Configuration API Endpoints
- **File:** `internal/portal/handlers/llm_config_handler.go`

**Endpoints:**
```
GET    /api/portal/llm-configs
POST   /api/portal/llm-configs
PUT    /api/portal/llm-configs/:id
DELETE /api/portal/llm-configs/:id
GET    /api/portal/llm-configs/providers (returns available providers/models)
POST   /api/portal/llm-configs/:id/test (health check)
GET    /api/portal/app-llm-preferences
PUT    /api/portal/app-llm-preferences/:app
GET    /api/portal/llm-usage/summary?period=7d
```

**Tests:**
- ✅ GET returns user's configs (API keys masked)
- ✅ POST creates config with encrypted API key
- ✅ POST validates provider exists
- ✅ POST validates model name format
- ✅ PUT updates config fields
- ✅ DELETE removes config
- ✅ DELETE fails if config in use
- ✅ GET providers returns static list
- ✅ POST test pings provider and returns status
- ✅ GET preferences returns app → config mapping
- ✅ PUT preference updates app preference
- ✅ GET usage summary aggregates by period
- ✅ All endpoints require authentication
- ✅ Users can only access their own configs

---

### Phase 4: Frontend - Prompt Editor (Days 11-13)

#### Task 4.1: Prompt Editor Modal Component
- **File:** `frontend/src/components/PromptEditorModal.jsx`

**TDD Steps:**
1. RED: Write Playwright test that opens modal, edits prompt, saves
2. GREEN: Build modal component
3. REFACTOR: Add syntax highlighting, variable validation

**Features:**
- Display current prompt (user custom or system default)
- Syntax highlighting for variables ({{code}}, {{query}})
- Variable reference panel
- Character count
- Save button (creates/updates user custom)
- Factory Reset button (only shown if custom exists)
- Cancel button

**Tests (Playwright):**
- ✅ Modal opens when clicking "Details" button
- ✅ Modal displays current prompt text
- ✅ Modal shows "Custom" badge if user has custom prompt
- ✅ Variable reference panel lists all available variables
- ✅ Editing prompt updates character count
- ✅ Save button creates user custom prompt
- ✅ Factory Reset button appears after saving custom
- ✅ Factory Reset removes custom, reloads default
- ✅ Cancel button closes modal without saving
- ✅ Modal persists prompt on page refresh (after save)

**Visual Tests (Percy):**
- ✅ Modal appearance (default state)
- ✅ Modal with custom prompt (badge visible)
- ✅ Variable reference panel expanded
- ✅ Long prompt text (scroll behavior)

---

#### Task 4.2: Add "Details" Buttons to Mode Cards
- **File:** `frontend/src/components/ReviewPage.jsx`

**Changes:**
- Add "Details" button to each mode card
- Track which mode's prompt is being edited
- Pass mode/userLevel/outputMode to modal

**Tests (Playwright):**
- ✅ Details button exists on Preview card
- ✅ Details button exists on Skim card
- ✅ Details button exists on Scan card
- ✅ Details button exists on Detailed card
- ✅ Details button exists on Critical card
- ✅ Clicking Details opens modal with correct mode
- ✅ Each mode loads its specific prompt

---

#### Task 4.3: Fix Clear/Reset Buttons
- **File:** `frontend/src/components/ReviewPage.jsx`

**Bug Fix:** Buttons currently use old `code`/`setCode` state, need to update to `files` array

**Updated Functions:**
```javascript
const clearCode = () => {
  setFiles(prevFiles => prevFiles.map(file => 
    file.id === activeFileId 
      ? { ...file, content: '', hasUnsavedChanges: false }
      : file
  ));
  setAnalysisResult(null);
  setError(null);
};

const resetToDefault = () => {
  const newFileId = `file_${Date.now()}`;
  setFiles([{
    id: newFileId,
    name: 'info.txt',
    language: 'plaintext',
    content: defaultCode,
    hasUnsavedChanges: false,
    path: null
  }]);
  setActiveFileId(newFileId);
  setAnalysisResult(null);
  setError(null);
  setTreeData(null);
  setShowTree(false);
};
```

**Tests (Playwright):**
- ✅ Clear button clears active file content
- ✅ Clear button clears analysis results
- ✅ Clear button does not affect other tabs
- ✅ Reset button replaces all files with default example
- ✅ Reset button clears file tree
- ✅ Reset button clears analysis results

---

### Phase 5: Frontend - LLM Configuration UI (Days 14-16)

#### Task 5.1: LLM Config Card on Portal Dashboard
- **File:** `frontend/src/components/PortalDashboard.jsx`

**Add Card:**
```jsx
<div className="card shadow-sm">
  <div className="card-body">
    <h5 className="card-title">
      <i className="bi bi-robot me-2"></i>
      AI Model Management
    </h5>
    <p className="card-text">
      Configure AI models and API keys for each app
    </p>
    <Link to="/llm-config" className="btn btn-primary">
      Manage Models
    </Link>
  </div>
</div>
```

**Tests (Playwright):**
- ✅ Card appears on portal dashboard
- ✅ Card has correct icon and text
- ✅ "Manage Models" button navigates to /llm-config

---

#### Task 5.2: LLM Configuration Page
- **File:** `frontend/src/pages/LLMConfigPage.jsx`

**Sections:**
1. Your AI Models (table of configs)
2. App-Specific Preferences (dropdowns)
3. Usage Summary (charts)

**Tests (Playwright):**
- ✅ Page loads at /llm-config
- ✅ "Your AI Models" table displays user's configs
- ✅ API keys shown as "Configured" badge, not plain text
- ✅ Default config has checkmark
- ✅ "Add Model" button opens modal
- ✅ Edit button opens edit modal
- ✅ Delete button removes config (after confirmation)
- ✅ App preference dropdowns show all user's configs
- ✅ Selecting preference updates immediately
- ✅ Usage summary displays total tokens/cost

**Visual Tests (Percy):**
- ✅ LLM Config page with no configs
- ✅ LLM Config page with multiple configs
- ✅ App preferences section
- ✅ Usage summary section

---

#### Task 5.3: Add LLM Config Modal
- **File:** `frontend/src/components/AddLLMConfigModal.jsx`

**Features:**
- Provider selection (Anthropic, OpenAI, Ollama, DeepSeek, Mistral)
- Model dropdown (filtered by provider)
- API key input (password field)
- Custom endpoint input (optional)
- Test connection button
- Save button

**Tests (Playwright):**
- ✅ Modal opens when clicking "Add Model"
- ✅ Provider dropdown lists all providers
- ✅ Model dropdown updates based on provider
- ✅ API key field is password type
- ✅ API key field hidden for Ollama (local)
- ✅ Custom endpoint field optional
- ✅ Test connection button pings provider
- ✅ Test connection shows success/failure
- ✅ Save button disabled until valid config
- ✅ Save button creates config and closes modal
- ✅ Newly created config appears in table

**Visual Tests (Percy):**
- ✅ Add modal initial state
- ✅ Add modal with Anthropic selected
- ✅ Add modal with Ollama selected (no API key field)
- ✅ Add modal test connection success
- ✅ Add modal test connection failure

---

#### Task 5.4: Manual Claude API Integration Test
- **User Action Required**
- **File:** `docs/MANUAL_TEST_CLAUDE.md` (instructions for Mike)

**Manual Test Steps:**
1. Navigate to /llm-config
2. Click "Add Model"
3. Select "Anthropic" provider
4. Select "claude-3-5-sonnet-20241022" model
5. Enter API key: `sk-ant-...` (Mike's actual key)
6. Click "Test Connection"
7. Verify success message
8. Click "Save"
9. Navigate to Review app
10. Set app preference to Claude config
11. Run code analysis
12. Verify Claude response received
13. Check usage logs show tokens/cost

**Expected Results:**
- ✅ Connection test succeeds
- ✅ Config saves successfully
- ✅ Review app uses Claude for analysis
- ✅ Response includes Claude-specific formatting
- ✅ Usage logs record tokens and cost
- ✅ Cost calculation accurate ($3/1M input, $15/1M output)

---

### Phase 6: Integration & E2E Testing (Days 17-19)

#### Task 6.1: Prompt Customization E2E Flow
- **File:** `tests/e2e/review/prompt-customization.spec.ts`

**Test Flow:**
1. Login as test user
2. Navigate to Review app
3. Click "Details" on Preview card
4. View system default prompt
5. Edit prompt text (add custom instruction)
6. Save custom prompt
7. Verify "Custom" badge appears
8. Run analysis with custom prompt
9. Verify AI follows custom instruction
10. Factory reset prompt
11. Verify default prompt restored
12. Run analysis again
13. Verify AI uses default behavior

**Percy Snapshots:**
- Prompt editor with default
- Prompt editor with custom (badge visible)
- Review results with custom prompt
- Review results after reset

---

#### Task 6.2: Multi-LLM Configuration E2E Flow
- **File:** `tests/e2e/portal/llm-config.spec.ts`

**Test Flow:**
1. Login as test user
2. Navigate to LLM Config page
3. Add Ollama config (local, no API key)
4. Verify config appears in table
5. Set Review app preference to Ollama
6. Navigate to Review app
7. Run analysis (should use Ollama)
8. Verify response from Ollama
9. Return to LLM Config page
10. Add mock OpenAI config (test API key)
11. Set Logs app preference to OpenAI
12. Navigate to Logs app
13. Trigger AI analysis (should use OpenAI mock)
14. Verify OpenAI response
15. Check usage summary shows both providers

**Percy Snapshots:**
- LLM Config page empty state
- LLM Config page with 2 configs
- App preferences set
- Usage summary with data

---

#### Task 6.3: Cross-App LLM Preference Test
- **File:** `tests/e2e/integration/cross-app-llm.spec.ts`

**Test Flow:**
1. Configure Review app: Claude (manual mock)
2. Configure Logs app: DeepSeek (mock)
3. Configure Analytics app: Ollama (local)
4. Run analysis in Review → verify Claude used
5. Run analysis in Logs → verify DeepSeek used
6. Run analysis in Analytics → verify Ollama used
7. Check usage logs show correct provider per app

---

#### Task 6.4: API Key Encryption Security Test
- **File:** `tests/integration/portal/encryption_test.go`

**Tests:**
- ✅ API key encrypted before DB insert
- ✅ Encrypted key different from plain key
- ✅ Same key encrypts differently each time (nonce)
- ✅ Decrypt returns original key
- ✅ Decrypt fails with wrong user ID
- ✅ Decrypt fails with corrupted data
- ✅ Master key rotation works (re-encrypt all keys)

---

### Phase 7: Documentation & Deployment (Day 20)

#### Task 7.1: User Documentation
- **File:** `docs/USER_GUIDE_PROMPTS.md`

**Contents:**
- What are prompts and why customize them
- How to access prompt editor
- How to use variables ({{code}}, {{query}})
- Best practices for prompt engineering
- Factory reset instructions
- Troubleshooting

---

#### Task 7.2: User Documentation - LLM Config
- **File:** `docs/USER_GUIDE_LLM_CONFIG.md`

**Contents:**
- Supported providers and models
- How to get API keys (Anthropic, OpenAI, etc.)
- How to add configurations
- How to set app preferences
- Understanding usage and costs
- Security notes (encryption, never exposed)
- Local vs API models comparison

---

#### Task 7.3: Developer Documentation
- **File:** `docs/DEV_GUIDE_MULTI_LLM.md`

**Contents:**
- Architecture overview
- How to add new provider
- AIClient interface specification
- Factory pattern usage
- Encryption service usage
- Testing strategy
- Error handling patterns

---

#### Task 7.4: Environment Variables
- **File:** `.env.example`

**Add:**
```bash
# Encryption for API keys
ENCRYPTION_MASTER_KEY=your-32-byte-base64-key-here

# Default LLM (fallback)
DEFAULT_LLM_PROVIDER=ollama
DEFAULT_LLM_MODEL=deepseek-coder:6.7b
DEFAULT_LLM_ENDPOINT=http://localhost:11434
```

---

## 🧪 Testing Strategy Summary

### Unit Tests
- **Target Coverage:** 70% minimum, 90% for critical paths
- **Frameworks:** Go testing, testify/assert, testify/mock
- **Mock External APIs:** All providers except Claude (manual)
- **Run Command:** `go test ./... -v -cover`

### Integration Tests
- **Database:** Use test database with transactions
- **External Services:** Mock HTTP servers
- **Cross-Service:** Test AI factory with real services
- **Run Command:** `go test ./... -tags=integration -v`

### E2E Tests (Playwright)
- **Browsers:** Chromium only (headless)
- **Scenarios:** Full user workflows
- **Visual Testing:** Percy snapshots
- **Run Command:** `npm run test:e2e`

### Manual Testing
- **Claude API:** User manually enters API key and tests
- **Instructions:** `docs/MANUAL_TEST_CLAUDE.md`

---

## 📊 Success Criteria

### Phase 1-2: Prompt Customization
- ✅ All 15 default prompts seeded
- ✅ Users can view, edit, save custom prompts
- ✅ Factory reset works correctly
- ✅ Prompt editor shows variables
- ✅ Custom prompts persist across sessions
- ✅ All unit tests pass (70%+ coverage)
- ✅ All integration tests pass
- ✅ E2E test passes
- ✅ Percy snapshots approved

### Phase 3-5: Multi-LLM Platform
- ✅ Encryption service encrypts/decrypts correctly
- ✅ All 5 provider clients implemented
- ✅ AI factory returns correct client per app
- ✅ Fallback chain works (primary → default → Ollama)
- ✅ LLM config UI functional
- ✅ App preferences save and apply
- ✅ Usage logs track tokens/cost
- ✅ Claude API manually tested (Mike)
- ✅ All unit tests pass (70%+ coverage)
- ✅ All integration tests pass
- ✅ All E2E tests pass
- ✅ Percy snapshots approved

### Overall Quality Gates
- ✅ No hardcoded values/stubs in production code
- ✅ No failing tests
- ✅ No linting errors
- ✅ API keys encrypted in DB (verified)
- ✅ Security audit passed (no plain-text keys)
- ✅ Documentation complete
- ✅ User can use platform without touching DB/config files

---

## 🚀 Deployment Checklist

### Pre-Deployment
- [ ] All tests passing
- [ ] Percy snapshots approved
- [ ] Database migrations tested on staging
- [ ] ENCRYPTION_MASTER_KEY generated and secured
- [ ] Environment variables documented
- [ ] User documentation complete
- [ ] Manual Claude test completed

### Deployment Steps
1. [ ] Generate and store ENCRYPTION_MASTER_KEY
2. [ ] Run migrations: `bash scripts/run-migrations.sh`
3. [ ] Verify seed data: `SELECT COUNT(*) FROM review.prompt_templates WHERE is_default = true;` (expect 15)
4. [ ] Rebuild services: `docker-compose up -d --build`
5. [ ] Run smoke tests: `bash scripts/regression-test.sh`
6. [ ] Manual smoke test: Create LLM config, set preference, run analysis
7. [ ] Monitor logs for errors
8. [ ] Verify usage tracking working

### Post-Deployment
- [ ] User notification: New features available
- [ ] Monitor error logs for 24h
- [ ] Check usage analytics
- [ ] Gather user feedback
- [ ] Document any issues in ERROR_LOG.md

---

## ❓ Open Questions

### Question 1: Master Key Storage
**Q:** Where should ENCRYPTION_MASTER_KEY be stored in production?  
**Options:**
- Environment variable (current approach)
- AWS Secrets Manager / Azure Key Vault
- HashiCorp Vault

**Recommendation:** Start with env var, migrate to secrets manager if scaling

---

### Question 2: API Key Rotation
**Q:** How should users rotate their API keys?  
**Options:**
- Edit config, enter new key (simple)
- "Rotate Key" button that re-encrypts (advanced)

**Recommendation:** Start with edit, add rotation later

---

### Question 3: Cost Limits
**Q:** Should there be default spending limits to prevent accidental $1000 bills?  
**Options:**
- No limits (user responsible)
- Soft limit ($50/month) with warning
- Hard limit ($100/month) with lockout

**Recommendation:** Soft limit with email alert

---

## 📋 Status Updates

### 2025-11-08 - Phase 1 Complete: Database Schema & Migrations
**Status:** ✅ Phase 1 Complete (Days 1-2)  
**Progress:** 2/20 days complete (10%)  
**Completed Tasks:**
- ✅ Created migration 20251108_001_prompt_templates.sql
  - prompt_templates table with mode/user_level/output_mode constraints
  - prompt_executions table for usage tracking
  - Proper indexes and triggers for updated_at
- ✅ Created migration 20251108_002_llm_configs.sql
  - llm_configs table with provider enum and encryption support
  - app_llm_preferences table for per-app LLM selection
  - llm_usage_logs table for token tracking and billing
  - Single-default trigger ensures only one default config per user
- ✅ Created seed data 20251108_001_default_prompts.sql
  - 15 default prompts (5 modes × 3 user levels)
  - All prompts use "quick" output_mode by default
  - Variables tracked in JSONB column
- ✅ Created comprehensive integration tests (tests/db/migrations_phase1_test.go)
  - Tests for constraint validation
  - Tests for foreign keys
  - Tests for triggers
  - Tests for seed data integrity
- ✅ Applied migrations to development database
  - All tables created successfully
  - All 15 default prompts seeded
  - Verification script confirms correct state

**Test Results:**
```
✓ Migration 20251108_001 applied successfully
✓ Migration 20251108_002 applied successfully
✓ Seed data applied successfully
✓ Found 15 default prompts (5 modes × 3 user levels)
✓ All tables and indexes created
✓ All constraints working correctly
```

**Next Steps:**
- Start Phase 2: Backend Services - Prompt Management
  - Task 2.1: Prompt Template Repository
  - Task 2.2: Prompt Template Service
  - Task 2.3: Prompt API Endpoints

**Notes:**
- Using standard PostgreSQL migrations (no ORM)
- All prompts include {{code}} variable
- Scan mode prompts include {{query}} variable
- ENCRYPTION_MASTER_KEY will be needed for Phase 3

### 2025-11-08 - Initial Planning Complete
**Status:** ✅ Planning Phase Complete  
**Progress:** 0/20 days complete (0%)  
**Next Steps:**
- User review and approval of plan
- Start Phase 1: Database migrations
- Generate ENCRYPTION_MASTER_KEY

**Questions for User:**
1. Do you have any questions about the implementation plan?
2. Should we proceed with Phase 1 (database migrations)?
3. Do you want to adjust any priorities or timelines?
4. Any additional requirements not covered?

---

### 2025-11-08 - Phase 1 Complete ✅
**Status:** Phase 1 Database Schema & Migrations COMPLETE  
**Progress:** 2/20 days complete (10%)  
**Duration:** ~2 hours

**Completed Tasks:**
1. ✅ Task 1.1: Prompt Templates Schema Created
   - Migration: `20251108_001_prompt_templates.sql`
   - Tables: `review.prompt_templates`, `review.prompt_executions`
   - Tests: 11 passing tests for table structure, constraints, indexes
   
2. ✅ Task 1.2: LLM Configuration Schema Created
   - Migration: `20251108_002_llm_configs.sql`
   - Tables: `portal.llm_configs`, `portal.app_llm_preferences`, `portal.llm_usage_logs`
   - Tests: 13 passing tests for constraints, foreign keys, uniqueness
   
3. ✅ Task 1.3: Default Prompts Seeded
   - Seed: `20251108_001_default_prompts.sql`
   - 15 default prompts inserted (5 modes × 3 user levels × 1 output mode)
   - Tests: 4 passing tests for seed data integrity

**Test Results:**
```
✓ TestMigration_PromptTemplates (0.10s)
✓ TestMigration_LLMConfigs (0.08s)
✓ TestSeeds_DefaultPrompts (0.04s)

PASS: All 3 test suites passing
Database: Tables created in main database with 15 default prompts
```

**Database Verification:**
```sql
-- Confirmed tables exist:
review.prompt_templates (11 columns, 3 indexes, triggers)
review.prompt_executions (9 columns, 3 indexes)
portal.llm_configs (11 columns, 3 indexes, triggers)
portal.app_llm_preferences (5 columns, 2 indexes)
portal.llm_usage_logs (10 columns, 4 indexes)

-- Confirmed seed data:
SELECT COUNT(*) FROM review.prompt_templates WHERE is_default = true;
-- Result: 15 (all 5 modes × 3 user levels)
```

**Next Steps:**
- Start Phase 2: Backend Services - Prompt Management
  - Task 2.1: Prompt Template Repository (TDD)
  - Task 2.2: Prompt Template Service (TDD)
  - Task 2.3: Prompt API Endpoints (TDD)

---

## 🔄 Instructions for New Chat Sessions

When starting a new chat session for this project:

1. **Reference This Document:** "Continue working on the Multi-LLM Platform implementation. See `MULTI_LLM_IMPLEMENTATION_PLAN.md` for full context."

2. **Check Latest Status:** Review the "Status Updates" section at the bottom of this document

3. **TDD Approach:** Always follow RED → GREEN → REFACTOR
   - RED: Write failing test first
   - GREEN: Implement minimal code to pass
   - REFACTOR: Improve code quality while keeping tests green

4. **No Shortcuts:** 
   - No hardcoded values
   - No stubs/mocks in production code
   - All features must be fully functional
   - Exception: Claude API (manual testing only)

5. **Update This Document:** After completing each task, append status update to this document with:
   - Date
   - Task completed
   - Test results
   - Any issues encountered
   - Next steps

6. **Commit Pattern:**
   ```bash
   git commit -m "feat(scope): description
   
   - What was implemented
   - Test results: XX/XX passing
   - Coverage: XX%
   
   Part of Multi-LLM Platform implementation"
   ```

7. **Ask Before Major Changes:** If you encounter issues requiring architectural changes, ask user before proceeding

8. **Documentation:** Keep user/dev docs updated as features are implemented

---

**END OF DOCUMENT**
