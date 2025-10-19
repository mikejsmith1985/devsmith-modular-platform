# Issue #004: [COPILOT] Review Service - Foundation + Preview Mode

**Labels:** `copilot`, `review`, `reading-mode`, `ai-integration`
**Assignee:** Mike (with Copilot assistance)
**Created:** 2025-10-19
**Issue:** #4
**Estimated Complexity:** High
**Target Service:** review
**Estimated Time:** 90-120 minutes
**Depends On:** Issue #003 (Portal Authentication)

---

## Task Description

Build the Review Service foundation and implement Preview Mode (the first of five reading modes). Preview Mode provides rapid assessment of code structure and organization - teaching users to understand high-level codebase architecture quickly.

**Why This Task for Copilot:**
- Clear bounded context specification
- Well-defined AI integration patterns
- Complete code examples provided
- Standard Go patterns (Gin, pgx, Ollama)
- Builds on existing Portal Auth foundation

**Why Preview Mode First:**
- Simplest of the five reading modes
- No deep code analysis required
- Establishes AI integration patterns
- Foundation for other modes

---

## IMPORTANT: Before You Start

### Step 1: Create Feature Branch

**In your terminal, run these commands:**
```bash
# Make sure you're on development and it's up to date
git checkout development
git pull origin development

# Create and switch to feature branch for this issue
git checkout -b feature/004-copilot-review-service-preview-mode
```

### Step 2: Commit As You Go

**DO NOT wait until everything is done to commit!**

After completing each PHASE (see Implementation Checklist below):
1. Test that phase: `go test ./...`
2. If tests pass, commit that phase:
   ```bash
   git add <files-for-that-phase>
   git commit -m "feat(review): <brief description of phase>"
   ```

**Example commits:**
```bash
# After Phase 1 (models)
git add internal/review/models/
git commit -m "feat(review): add Review, Repository, and AnalysisResult models"

# After Phase 2 (database layer)
git add internal/review/db/
git commit -m "feat(review): implement review repository with CRUD operations"

# After Phase 3 (Ollama client)
git add internal/review/services/ollama_client.go
git commit -m "feat(review): implement Ollama API client for AI analysis"

# And so on...
```

### Step 3: Push Regularly

After every 2-3 commits:
```bash
git push -u origin feature/004-copilot-review-service-preview-mode
```

---

## Overview

### Feature Description

Implement the Review Service with Preview Mode - the first of five AI-powered reading modes. Preview Mode helps users rapidly understand code structure by providing high-level architectural analysis without diving into implementation details.

### User Story

As a developer learning to read code effectively, I want to quickly assess a GitHub repository's structure and architecture so that I can understand its organization before diving into implementation details.

### Success Criteria

- [ ] User can create a new review for a GitHub repository
- [ ] Preview Mode analyzes repository structure via Ollama API
- [ ] AI generates file structure tree with descriptions
- [ ] AI identifies bounded contexts and architectural patterns
- [ ] Results display in Templ-based UI with collapsible sections
- [ ] Review persists in database (review.reviews and review.analysis_results)
- [ ] User can switch between reading modes (UI only, other modes not implemented)
- [ ] All tests pass with 70%+ coverage
- [ ] Service health check endpoint works

---

## Context for Cognitive Load Management

### Bounded Context

**Service:** Review
**Domain:** Code Review and Analysis
**Related Entities:**
- `Review` (review context) - A user's analysis session for a repository
- `Repository` (GitHub repo being reviewed) - URL, owner, name, default branch
- `AnalysisResult` (AI output for a specific reading mode) - Preview, Skim, Scan, Detailed, Critical
- `ReadingMode` (enum) - Which of the five modes is active

**Context Boundaries:**
- ✅ **Within scope:** Code analysis, AI prompts, reading modes, review sessions, repository metadata
- ❌ **Out of scope:** User authentication (Portal), activity logging (Logs service), usage metrics (Analytics)

**Why This Separation:**
The Review service ONLY knows about "analyzing code repositories." It does NOT handle authentication (Portal does that), activity logging (Logs service), or usage tracking (Analytics service).

---

### Layering

**Primary Layer:** All three layers required (Controller → Orchestration → Data)

#### Controller Layer Files

```
cmd/review/handlers/
├── review_handler.go              # Create review, get review, switch modes
├── review_handler_test.go         # Handler tests
├── preview_handler.go             # Preview Mode specific endpoints
└── preview_handler_test.go        # Preview tests

cmd/review/templates/
├── layout.templ                   # Base HTML layout
├── new_review.templ               # Create new review form
├── review_dashboard.templ         # Main review interface
├── preview_mode.templ             # Preview Mode UI
└── components/
    ├── reading_mode_selector.templ  # Toggle between 5 modes
    ├── file_tree.templ              # Collapsible file tree
    └── ai_summary_panel.templ       # AI analysis output
```

#### Orchestration Layer Files

```
internal/review/services/
├── review_service.go              # Review business logic
├── review_service_test.go         # Service tests
├── preview_service.go             # Preview Mode analysis logic
├── preview_service_test.go        # Preview tests
├── ollama_client.go               # Ollama API integration
├── ollama_client_test.go          # Ollama client tests
├── github_repo_fetcher.go         # Fetch repo metadata from GitHub
└── github_repo_fetcher_test.go    # GitHub fetcher tests

internal/review/interfaces/
└── services_interface.go          # Abstract contracts for testing
```

#### Data Layer Files

```
internal/review/db/
├── review_repository.go           # Review CRUD operations
├── review_repository_test.go      # Repository tests
├── analysis_repository.go         # Analysis results CRUD
├── analysis_repository_test.go    # Analysis tests
└── migrations/
    ├── 20251019_001_create_reviews_table.sql
    ├── 20251019_002_create_repositories_table.sql
    └── 20251019_003_create_analysis_results_table.sql
```

**Cross-Layer Rules:**
- ✅ `review_handler.go` calls `review_service.go`
- ✅ `review_service.go` calls `review_repository.go`
- ✅ `preview_service.go` calls `ollama_client.go` and `analysis_repository.go`
- ❌ `review_handler.go` MUST NOT call `review_repository.go` directly
- ❌ `review_repository.go` MUST NOT import service or handler packages
- ❌ No circular dependencies between layers

---

## Implementation Specification

### Phase 1: Data Models

#### File: `internal/review/models/review.go`

```go
package models

import "time"

// ReadingMode represents the five modes of code reading
type ReadingMode string

const (
	PreviewMode  ReadingMode = "preview"
	SkimMode     ReadingMode = "skim"
	ScanMode     ReadingMode = "scan"
	DetailedMode ReadingMode = "detailed"
	CriticalMode ReadingMode = "critical"
)

// Review represents a user's code review session for a repository
type Review struct {
	ID           int64       `json:"id" db:"id"`
	UserID       int64       `json:"user_id" db:"user_id"`                 // References portal.users.id
	RepositoryID int64       `json:"repository_id" db:"repository_id"`     // References review.repositories.id
	Title        string      `json:"title" db:"title"`                     // User-provided review title
	CurrentMode  ReadingMode `json:"current_mode" db:"current_mode"`       // Active reading mode
	Status       string      `json:"status" db:"status"`                   // "in_progress", "completed"
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// Repository represents a GitHub repository being reviewed
type Repository struct {
	ID            int64     `json:"id" db:"id"`
	GithubURL     string    `json:"github_url" db:"github_url"`         // Full GitHub URL
	Owner         string    `json:"owner" db:"owner"`                   // Repo owner (user/org)
	Name          string    `json:"name" db:"name"`                     // Repo name
	DefaultBranch string    `json:"default_branch" db:"default_branch"` // Usually "main" or "master"
	LastFetchedAt time.Time `json:"last_fetched_at" db:"last_fetched_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// AnalysisResult stores AI analysis output for a specific reading mode
type AnalysisResult struct {
	ID         int64       `json:"id" db:"id"`
	ReviewID   int64       `json:"review_id" db:"review_id"`       // References review.reviews.id
	Mode       ReadingMode `json:"mode" db:"mode"`                 // Which reading mode generated this
	Prompt     string      `json:"prompt" db:"prompt"`             // Prompt sent to AI
	RawOutput  string      `json:"raw_output" db:"raw_output"`     // Full AI response
	Summary    string      `json:"summary" db:"summary"`           // Extracted summary for UI
	Metadata   string      `json:"metadata" db:"metadata"`         // JSON metadata (e.g., file structure)
	ModelUsed  string      `json:"model_used" db:"model_used"`     // e.g., "qwen2.5-coder:32b"
	TokensUsed int         `json:"tokens_used" db:"tokens_used"`   // For tracking AI costs
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

// PreviewModeOutput represents structured Preview Mode analysis
type PreviewModeOutput struct {
	FileStructure      []FileNode          `json:"file_structure"`
	BoundedContexts    []string            `json:"bounded_contexts"`
	TechnologyStack    []string            `json:"technology_stack"`
	ArchitecturalStyle string              `json:"architectural_style"` // e.g., "layered", "microservices"
	EntryPoints        []string            `json:"entry_points"`
	Dependencies       []DependencyInfo    `json:"dependencies"`
	Summary            string              `json:"summary"`
}

// FileNode represents a file or directory in the structure tree
type FileNode struct {
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	Type        string     `json:"type"` // "file" or "directory"
	Layer       string     `json:"layer"` // "controller", "service", "data", "config", "other"
	Description string     `json:"description"` // AI-generated description
	Children    []FileNode `json:"children,omitempty"`
}

// DependencyInfo represents an external dependency
type DependencyInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Purpose     string `json:"purpose"` // AI-generated purpose description
}
```

---

### Phase 2: Database Layer

#### File: `internal/review/db/migrations/20251019_001_create_reviews_table.sql`

```sql
-- Create review schema if not exists (isolated from portal schema)
CREATE SCHEMA IF NOT EXISTS review;

-- Reviews table
CREATE TABLE IF NOT EXISTS review.reviews (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,  -- References portal.users(id)
    repository_id BIGINT NOT NULL,  -- References review.repositories(id)
    title VARCHAR(255) NOT NULL,
    current_mode VARCHAR(20) NOT NULL DEFAULT 'preview',
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES portal.users(id) ON DELETE CASCADE,
    CONSTRAINT fk_repository FOREIGN KEY (repository_id) REFERENCES review.repositories(id) ON DELETE CASCADE,
    CONSTRAINT chk_mode CHECK (current_mode IN ('preview', 'skim', 'scan', 'detailed', 'critical')),
    CONSTRAINT chk_status CHECK (status IN ('in_progress', 'completed'))
);

-- Indexes for common queries
CREATE INDEX idx_reviews_user_id ON review.reviews(user_id);
CREATE INDEX idx_reviews_repository_id ON review.reviews(repository_id);
CREATE INDEX idx_reviews_status ON review.reviews(status);
CREATE INDEX idx_reviews_created_at ON review.reviews(created_at DESC);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION review.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON review.reviews
    FOR EACH ROW EXECUTE FUNCTION review.update_updated_at_column();
```

#### File: `internal/review/db/migrations/20251019_002_create_repositories_table.sql`

```sql
-- Repositories table
CREATE TABLE IF NOT EXISTS review.repositories (
    id BIGSERIAL PRIMARY KEY,
    github_url VARCHAR(500) NOT NULL UNIQUE,
    owner VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    default_branch VARCHAR(100) NOT NULL DEFAULT 'main',
    last_fetched_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_owner_name UNIQUE (owner, name)
);

-- Indexes
CREATE INDEX idx_repositories_owner_name ON review.repositories(owner, name);
CREATE INDEX idx_repositories_github_url ON review.repositories(github_url);
```

#### File: `internal/review/db/migrations/20251019_003_create_analysis_results_table.sql`

```sql
-- Analysis results table
CREATE TABLE IF NOT EXISTS review.analysis_results (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT NOT NULL,
    mode VARCHAR(20) NOT NULL,
    prompt TEXT NOT NULL,
    raw_output TEXT NOT NULL,
    summary TEXT,
    metadata JSONB,  -- Store structured data like file trees
    model_used VARCHAR(100) NOT NULL,
    tokens_used INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_review FOREIGN KEY (review_id) REFERENCES review.reviews(id) ON DELETE CASCADE,
    CONSTRAINT chk_analysis_mode CHECK (mode IN ('preview', 'skim', 'scan', 'detailed', 'critical'))
);

-- Indexes
CREATE INDEX idx_analysis_review_id ON review.analysis_results(review_id);
CREATE INDEX idx_analysis_mode ON review.analysis_results(mode);
CREATE INDEX idx_analysis_created_at ON review.analysis_results(created_at DESC);
CREATE INDEX idx_analysis_metadata ON review.analysis_results USING gin (metadata);  -- JSONB index
```

#### File: `internal/review/db/review_repository.go`

```go
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create creates a new review
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	query := `
		INSERT INTO review.reviews (user_id, repository_id, title, current_mode, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		review.UserID,
		review.RepositoryID,
		review.Title,
		review.CurrentMode,
		review.Status,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}
	return nil
}

// FindByID retrieves a review by ID
func (r *ReviewRepository) FindByID(ctx context.Context, id int64) (*models.Review, error) {
	review := &models.Review{}
	query := `
		SELECT id, user_id, repository_id, title, current_mode, status, created_at, updated_at
		FROM review.reviews
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&review.ID,
		&review.UserID,
		&review.RepositoryID,
		&review.Title,
		&review.CurrentMode,
		&review.Status,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find review: %w", err)
	}
	return review, nil
}

// FindByUserID retrieves all reviews for a user
func (r *ReviewRepository) FindByUserID(ctx context.Context, userID int64) ([]*models.Review, error) {
	query := `
		SELECT id, user_id, repository_id, title, current_mode, status, created_at, updated_at
		FROM review.reviews
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find reviews by user: %w", err)
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		review := &models.Review{}
		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.RepositoryID,
			&review.Title,
			&review.CurrentMode,
			&review.Status,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

// UpdateMode updates the current reading mode
func (r *ReviewRepository) UpdateMode(ctx context.Context, id int64, mode models.ReadingMode) error {
	query := `
		UPDATE review.reviews
		SET current_mode = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, mode, id)
	if err != nil {
		return fmt.Errorf("failed to update review mode: %w", err)
	}
	return nil
}

// UpdateStatus updates the review status
func (r *ReviewRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE review.reviews
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}
	return nil
}

// Delete deletes a review (cascade deletes analysis results)
func (r *ReviewRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM review.reviews WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}
	return nil
}
```

#### File: `internal/review/db/repository_repository.go`

```go
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type RepositoryRepository struct {
	db *pgxpool.Pool
}

func NewRepositoryRepository(db *pgxpool.Pool) *RepositoryRepository {
	return &RepositoryRepository{db: db}
}

// CreateOrGet creates a new repository or returns existing one
func (r *RepositoryRepository) CreateOrGet(ctx context.Context, repo *models.Repository) error {
	// Try to find existing
	query := `SELECT id, github_url, owner, name, default_branch, last_fetched_at, created_at
			  FROM review.repositories WHERE owner = $1 AND name = $2`
	err := r.db.QueryRow(ctx, query, repo.Owner, repo.Name).Scan(
		&repo.ID,
		&repo.GithubURL,
		&repo.Owner,
		&repo.Name,
		&repo.DefaultBranch,
		&repo.LastFetchedAt,
		&repo.CreatedAt,
	)
	if err == nil {
		// Found existing
		return nil
	}

	// Create new
	query = `
		INSERT INTO review.repositories (github_url, owner, name, default_branch)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err = r.db.QueryRow(ctx, query,
		repo.GithubURL,
		repo.Owner,
		repo.Name,
		repo.DefaultBranch,
	).Scan(&repo.ID, &repo.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	return nil
}

// FindByID retrieves a repository by ID
func (r *RepositoryRepository) FindByID(ctx context.Context, id int64) (*models.Repository, error) {
	repo := &models.Repository{}
	query := `
		SELECT id, github_url, owner, name, default_branch, last_fetched_at, created_at
		FROM review.repositories
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&repo.ID,
		&repo.GithubURL,
		&repo.Owner,
		&repo.Name,
		&repo.DefaultBranch,
		&repo.LastFetchedAt,
		&repo.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find repository: %w", err)
	}
	return repo, nil
}
```

#### File: `internal/review/db/analysis_repository.go`

```go
package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type AnalysisRepository struct {
	db *pgxpool.Pool
}

func NewAnalysisRepository(db *pgxpool.Pool) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

// Create stores a new analysis result
func (r *AnalysisRepository) Create(ctx context.Context, result *models.AnalysisResult) error {
	// Convert metadata to JSON if needed
	var metadataJSON []byte
	if result.Metadata != "" {
		metadataJSON = []byte(result.Metadata)
	}

	query := `
		INSERT INTO review.analysis_results (review_id, mode, prompt, raw_output, summary, metadata, model_used, tokens_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		result.ReviewID,
		result.Mode,
		result.Prompt,
		result.RawOutput,
		result.Summary,
		metadataJSON,
		result.ModelUsed,
		result.TokensUsed,
	).Scan(&result.ID, &result.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create analysis result: %w", err)
	}
	return nil
}

// FindByReviewAndMode retrieves analysis result for a review and mode
func (r *AnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode models.ReadingMode) (*models.AnalysisResult, error) {
	result := &models.AnalysisResult{}
	var metadataJSON []byte

	query := `
		SELECT id, review_id, mode, prompt, raw_output, summary, metadata, model_used, tokens_used, created_at
		FROM review.analysis_results
		WHERE review_id = $1 AND mode = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.db.QueryRow(ctx, query, reviewID, mode).Scan(
		&result.ID,
		&result.ReviewID,
		&result.Mode,
		&result.Prompt,
		&result.RawOutput,
		&result.Summary,
		&metadataJSON,
		&result.ModelUsed,
		&result.TokensUsed,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find analysis result: %w", err)
	}

	// Convert JSON metadata to string
	if metadataJSON != nil {
		result.Metadata = string(metadataJSON)
	}

	return result, nil
}

// FindAllByReview retrieves all analysis results for a review
func (r *AnalysisRepository) FindAllByReview(ctx context.Context, reviewID int64) ([]*models.AnalysisResult, error) {
	query := `
		SELECT id, review_id, mode, prompt, raw_output, summary, metadata, model_used, tokens_used, created_at
		FROM review.analysis_results
		WHERE review_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, reviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to find analysis results: %w", err)
	}
	defer rows.Close()

	var results []*models.AnalysisResult
	for rows.Next() {
		result := &models.AnalysisResult{}
		var metadataJSON []byte
		err := rows.Scan(
			&result.ID,
			&result.ReviewID,
			&result.Mode,
			&result.Prompt,
			&result.RawOutput,
			&result.Summary,
			&metadataJSON,
			&result.ModelUsed,
			&result.TokensUsed,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis result: %w", err)
		}
		if metadataJSON != nil {
			result.Metadata = string(metadataJSON)
		}
		results = append(results, result)
	}
	return results, nil
}
```

---

### Phase 3: Ollama Integration

#### File: `internal/review/services/ollama_client.go`

```go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

// OllamaRequest represents the request body for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

func NewOllamaClient() *OllamaClient {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama URL
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "qwen2.5-coder:32b" // Default model
	}

	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// Generate sends a prompt to Ollama and returns the response
func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaResponse
	err = json.Unmarshal(body, &ollamaResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Response, nil
}
```

---

### Phase 4: Preview Mode Service

#### File: `internal/review/services/preview_service.go`

```go
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type PreviewService struct {
	ollamaClient *OllamaClient
	analysisRepo AnalysisRepositoryInterface
}

type AnalysisRepositoryInterface interface {
	Create(ctx context.Context, result *models.AnalysisResult) error
	FindByReviewAndMode(ctx context.Context, reviewID int64, mode models.ReadingMode) (*models.AnalysisResult, error)
}

func NewPreviewService(ollamaClient *OllamaClient, analysisRepo AnalysisRepositoryInterface) *PreviewService {
	return &PreviewService{
		ollamaClient: ollamaClient,
		analysisRepo: analysisRepo,
	}
}

// AnalyzePreview generates Preview Mode analysis for a repository
func (s *PreviewService) AnalyzePreview(ctx context.Context, reviewID int64, repoOwner, repoName string) (*models.PreviewModeOutput, error) {
	// Check if we already have a cached analysis
	existing, err := s.analysisRepo.FindByReviewAndMode(ctx, reviewID, models.PreviewMode)
	if err == nil && existing != nil {
		// Parse cached result
		var output models.PreviewModeOutput
		if err := json.Unmarshal([]byte(existing.Metadata), &output); err == nil {
			return &output, nil
		}
	}

	// Generate prompt for Preview Mode
	prompt := s.buildPreviewPrompt(repoOwner, repoName)

	// Call Ollama API
	rawOutput, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate preview analysis: %w", err)
	}

	// Parse AI response into structured output
	output, err := s.parsePreviewOutput(rawOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse preview output: %w", err)
	}

	// Store result in database
	metadataJSON, _ := json.Marshal(output)
	analysisResult := &models.AnalysisResult{
		ReviewID:   reviewID,
		Mode:       models.PreviewMode,
		Prompt:     prompt,
		RawOutput:  rawOutput,
		Summary:    output.Summary,
		Metadata:   string(metadataJSON),
		ModelUsed:  "qwen2.5-coder:32b", // TODO: Get from config
		TokensUsed: 0, // TODO: Calculate tokens
	}

	err = s.analysisRepo.Create(ctx, analysisResult)
	if err != nil {
		return nil, fmt.Errorf("failed to store analysis result: %w", err)
	}

	return output, nil
}

// buildPreviewPrompt constructs the AI prompt for Preview Mode analysis
func (s *PreviewService) buildPreviewPrompt(owner, repo string) string {
	return fmt.Sprintf(`You are a code analysis assistant. Analyze the GitHub repository %s/%s in Preview Mode.

Preview Mode Goal: Provide rapid assessment of code structure and organization WITHOUT diving into implementation details.

Provide the following in JSON format:

1. file_structure: Array of file/directory nodes with:
   - name: filename
   - path: relative path
   - type: "file" or "directory"
   - layer: "controller", "service", "data", "config", or "other"
   - description: Brief description (1 sentence)
   - children: nested nodes for directories

2. bounded_contexts: Array of identified domain contexts (e.g., ["Authentication", "User Management"])

3. technology_stack: Array of technologies detected (e.g., ["Go 1.22", "PostgreSQL", "Gin", "HTMX"])

4. architectural_style: String describing the architecture (e.g., "Layered monolith with 3-tier structure")

5. entry_points: Array of main entry points (e.g., ["cmd/main.go", "cmd/server/main.go"])

6. dependencies: Array of key external dependencies with:
   - name: Package name
   - version: Version string
   - purpose: Why it's used (1 sentence)

7. summary: 2-3 sentence high-level summary of what this codebase does

Respond ONLY with valid JSON. No additional text.

Repository: https://github.com/%s/%s`, owner, repo, owner, repo)
}

// parsePreviewOutput parses the AI response into structured PreviewModeOutput
func (s *PreviewService) parsePreviewOutput(rawOutput string) (*models.PreviewModeOutput, error) {
	var output models.PreviewModeOutput
	err := json.Unmarshal([]byte(rawOutput), &output)
	if err != nil {
		// AI didn't return valid JSON, create a fallback response
		return &models.PreviewModeOutput{
			Summary: "Failed to parse AI response. The analysis may be incomplete.",
			FileStructure: []models.FileNode{},
			BoundedContexts: []string{},
			TechnologyStack: []string{},
			ArchitecturalStyle: "Unknown",
			EntryPoints: []string{},
			Dependencies: []models.DependencyInfo{},
		}, fmt.Errorf("invalid JSON from AI: %w", err)
	}
	return &output, nil
}
```

---

### Phase 5: Review Service (Orchestration)

#### File: `internal/review/services/review_service.go`

```go
package services

import (
	"context"
	"fmt"

	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

type ReviewService struct {
	reviewRepo     ReviewRepositoryInterface
	repositoryRepo RepositoryRepositoryInterface
	previewService *PreviewService
}

type ReviewRepositoryInterface interface {
	Create(ctx context.Context, review *models.Review) error
	FindByID(ctx context.Context, id int64) (*models.Review, error)
	FindByUserID(ctx context.Context, userID int64) ([]*models.Review, error)
	UpdateMode(ctx context.Context, id int64, mode models.ReadingMode) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
}

type RepositoryRepositoryInterface interface {
	CreateOrGet(ctx context.Context, repo *models.Repository) error
	FindByID(ctx context.Context, id int64) (*models.Repository, error)
}

func NewReviewService(
	reviewRepo ReviewRepositoryInterface,
	repositoryRepo RepositoryRepositoryInterface,
	previewService *PreviewService,
) *ReviewService {
	return &ReviewService{
		reviewRepo:     reviewRepo,
		repositoryRepo: repositoryRepo,
		previewService: previewService,
	}
}

// CreateReview creates a new code review session
func (s *ReviewService) CreateReview(ctx context.Context, userID int64, githubURL, title string) (*models.Review, error) {
	// Parse GitHub URL to extract owner and repo name
	owner, name, err := parseGitHubURL(githubURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub URL: %w", err)
	}

	// Create or get repository record
	repo := &models.Repository{
		GithubURL:     githubURL,
		Owner:         owner,
		Name:          name,
		DefaultBranch: "main", // TODO: Fetch actual default branch from GitHub API
	}
	err = s.repositoryRepo.CreateOrGet(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create/get repository: %w", err)
	}

	// Create review record
	review := &models.Review{
		UserID:       userID,
		RepositoryID: repo.ID,
		Title:        title,
		CurrentMode:  models.PreviewMode, // Start with Preview Mode
		Status:       "in_progress",
	}
	err = s.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return review, nil
}

// GetReview retrieves a review by ID
func (s *ReviewService) GetReview(ctx context.Context, id int64) (*models.Review, error) {
	return s.reviewRepo.FindByID(ctx, id)
}

// GetUserReviews retrieves all reviews for a user
func (s *ReviewService) GetUserReviews(ctx context.Context, userID int64) ([]*models.Review, error) {
	return s.reviewRepo.FindByUserID(ctx, userID)
}

// SwitchMode changes the current reading mode for a review
func (s *ReviewService) SwitchMode(ctx context.Context, reviewID int64, mode models.ReadingMode) error {
	return s.reviewRepo.UpdateMode(ctx, reviewID, mode)
}

// CompleteReview marks a review as completed
func (s *ReviewService) CompleteReview(ctx context.Context, reviewID int64) error {
	return s.reviewRepo.UpdateStatus(ctx, reviewID, "completed")
}

// parseGitHubURL extracts owner and repo name from a GitHub URL
func parseGitHubURL(url string) (owner, repo string, err error) {
	// Simple parsing (real implementation should be more robust)
	// Example: https://github.com/owner/repo or https://github.com/owner/repo.git
	// This is a stub - implement proper URL parsing
	return "owner", "repo", nil // TODO: Implement actual parsing
}
```

---

### Phase 6: HTTP Handlers (Controller Layer)

#### File: `cmd/review/handlers/review_handler.go`

```go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

type ReviewHandler struct {
	reviewService  *services.ReviewService
	previewService *services.PreviewService
}

func NewReviewHandler(reviewService *services.ReviewService, previewService *services.PreviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService:  reviewService,
		previewService: previewService,
	}
}

// CreateReview handles POST /api/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req struct {
		GitHubURL string `json:"github_url" binding:"required"`
		Title     string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from session (assumes Portal auth middleware sets this)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	review, err := h.reviewService.CreateReview(c.Request.Context(), userID.(int64), req.GitHubURL, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// GetReview handles GET /api/reviews/:id
func (h *ReviewHandler) GetReview(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetUserReviews handles GET /api/reviews
func (h *ReviewHandler) GetUserReviews(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	reviews, err := h.reviewService.GetUserReviews(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// SwitchMode handles POST /api/reviews/:id/mode
func (h *ReviewHandler) SwitchMode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req struct {
		Mode models.ReadingMode `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.reviewService.SwitchMode(c.Request.Context(), id, req.Mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mode switched"})
}

// GetPreviewAnalysis handles GET /api/reviews/:id/preview
func (h *ReviewHandler) GetPreviewAnalysis(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	// Get review to fetch repository info
	review, err := h.reviewService.GetReview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	// TODO: Get actual repo owner/name from repository record
	output, err := h.previewService.AnalyzePreview(c.Request.Context(), review.ID, "owner", "repo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
```

---

### Phase 7: Main Service Entry Point

#### File: `cmd/review/main.go`

Update the existing stub to include proper routes:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mikejsmith1985/devsmith-modular-platform/cmd/review/handlers"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/db"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/services"
)

func main() {
	// Get database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Initialize repositories
	reviewRepo := db.NewReviewRepository(dbPool)
	repositoryRepo := db.NewRepositoryRepository(dbPool)
	analysisRepo := db.NewAnalysisRepository(dbPool)

	// Initialize services
	ollamaClient := services.NewOllamaClient()
	previewService := services.NewPreviewService(ollamaClient, analysisRepo)
	reviewService := services.NewReviewService(reviewRepo, repositoryRepo, previewService)

	// Initialize handlers
	reviewHandler := handlers.NewReviewHandler(reviewService, previewService)

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "review",
			"status":  "healthy",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		api.POST("/reviews", reviewHandler.CreateReview)
		api.GET("/reviews", reviewHandler.GetUserReviews)
		api.GET("/reviews/:id", reviewHandler.GetReview)
		api.POST("/reviews/:id/mode", reviewHandler.SwitchMode)
		api.GET("/reviews/:id/preview", reviewHandler.GetPreviewAnalysis)
	}

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Printf("Review service starting on port %s...\\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

---

### Phase 8: Testing

#### File: `internal/review/services/preview_service_test.go`

```go
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/review/models"
)

// Mock implementations
type MockOllamaClient struct {
	mock.Mock
}

func (m *MockOllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

type MockAnalysisRepository struct {
	mock.Mock
}

func (m *MockAnalysisRepository) Create(ctx context.Context, result *models.AnalysisResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockAnalysisRepository) FindByReviewAndMode(ctx context.Context, reviewID int64, mode models.ReadingMode) (*models.AnalysisResult, error) {
	args := m.Called(ctx, reviewID, mode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalysisResult), args.Error(1)
}

// Test: Preview analysis generates and stores result
func TestPreviewService_AnalyzePreview_Success(t *testing.T) {
	// Arrange
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewPreviewService(mockOllama, mockRepo)

	ctx := context.Background()
	reviewID := int64(1)
	repoOwner := "testowner"
	repoName := "testrepo"

	// No cached result
	mockRepo.On("FindByReviewAndMode", ctx, reviewID, models.PreviewMode).
		Return(nil, fmt.Errorf("not found"))

	// Mock AI response (simplified JSON)
	aiResponse := `{
		"file_structure": [],
		"bounded_contexts": ["Authentication"],
		"technology_stack": ["Go"],
		"architectural_style": "Layered",
		"entry_points": ["main.go"],
		"dependencies": [],
		"summary": "A Go web service"
	}`
	mockOllama.On("Generate", ctx, mock.AnythingOfType("string")).
		Return(aiResponse, nil)

	// Mock repository save
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.AnalysisResult")).
		Return(nil)

	// Act
	output, err := service.AnalyzePreview(ctx, reviewID, repoOwner, repoName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "A Go web service", output.Summary)
	assert.Contains(t, output.BoundedContexts, "Authentication")
	mockOllama.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test: Cached analysis returned without calling Ollama
func TestPreviewService_AnalyzePreview_CacheHit(t *testing.T) {
	// Arrange
	mockOllama := new(MockOllamaClient)
	mockRepo := new(MockAnalysisRepository)
	service := NewPreviewService(mockOllama, mockRepo)

	ctx := context.Background()
	reviewID := int64(1)

	// Cached result exists
	cachedMetadata := `{"summary": "Cached summary", "file_structure": [], "bounded_contexts": [], "technology_stack": [], "architectural_style": "", "entry_points": [], "dependencies": []}`
	cachedResult := &models.AnalysisResult{
		ID:       1,
		ReviewID: reviewID,
		Mode:     models.PreviewMode,
		Metadata: cachedMetadata,
	}
	mockRepo.On("FindByReviewAndMode", ctx, reviewID, models.PreviewMode).
		Return(cachedResult, nil)

	// Act
	output, err := service.AnalyzePreview(ctx, reviewID, "owner", "repo")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Cached summary", output.Summary)
	// Ollama should NOT have been called
	mockOllama.AssertNotCalled(t, "Generate")
	mockRepo.AssertExpectations(t)
}
```

---

## Implementation Checklist

### Phase 1: Models ✅
- [ ] Create `internal/review/models/review.go` with all models
- [ ] Run: `go test ./internal/review/models/...`
- [ ] Commit: `git add internal/review/models/ && git commit -m "feat(review): add Review, Repository, and AnalysisResult models"`

### Phase 2: Database Layer ✅
- [ ] Create migration files in `internal/review/db/migrations/`
- [ ] Create `review_repository.go`
- [ ] Create `repository_repository.go`
- [ ] Create `analysis_repository.go`
- [ ] Run: `go test ./internal/review/db/...`
- [ ] Commit: `git add internal/review/db/ && git commit -m "feat(review): implement database layer with repositories and migrations"`

### Phase 3: Ollama Integration ✅
- [ ] Create `internal/review/services/ollama_client.go`
- [ ] Test manually: Start Ollama, send test request
- [ ] Commit: `git add internal/review/services/ollama_client.go && git commit -m "feat(review): implement Ollama API client for AI analysis"`

### Phase 4: Preview Service ✅
- [ ] Create `internal/review/services/preview_service.go`
- [ ] Create `internal/review/services/preview_service_test.go`
- [ ] Run: `go test ./internal/review/services/...`
- [ ] Commit: `git add internal/review/services/preview* && git commit -m "feat(review): implement Preview Mode analysis service"`

### Phase 5: Review Service ✅
- [ ] Create `internal/review/services/review_service.go`
- [ ] Run: `go test ./internal/review/services/...`
- [ ] Commit: `git add internal/review/services/review_service.go && git commit -m "feat(review): implement review orchestration service"`

### Phase 6: HTTP Handlers ✅
- [ ] Create `cmd/review/handlers/review_handler.go`
- [ ] Update `cmd/review/main.go` with routes
- [ ] Run: `go test ./cmd/review/handlers/...`
- [ ] Commit: `git add cmd/review/ && git commit -m "feat(review): add HTTP handlers and API routes"`

### Phase 7: Integration Testing ✅
- [ ] Start services: `make dev`
- [ ] Test health endpoint: `curl http://localhost:3000/review/health`
- [ ] Test create review API (use Postman/curl)
- [ ] Test preview analysis endpoint
- [ ] Verify database records created

### Phase 8: Final PR ✅
- [ ] Review all commits: `git log development..HEAD --oneline`
- [ ] Run full test suite: `make test`
- [ ] Push: `git push`
- [ ] Create PR on GitHub (Title: `[Issue #004] Review Service - Preview Mode`)
- [ ] Verify CI passes
- [ ] Tag @Claude for review

---

## Environment Variables

Add to `.env.example`:

```bash
# Ollama Configuration
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=qwen2.5-coder:32b

# Review Service
REVIEW_PORT=8081
```

---

## Testing Strategy

### Unit Tests (70%+ coverage required)

**Test Coverage Targets:**
- Models: 80%+ (mostly data structures, focus on validation logic)
- Repositories: 75%+ (test CRUD operations with test database)
- Services: 80%+ (mock dependencies, test business logic)
- Handlers: 70%+ (test HTTP request/response handling)

**Key Test Cases:**
1. ✅ Preview Service generates analysis successfully
2. ✅ Preview Service returns cached results when available
3. ✅ Review Repository CRUD operations work correctly
4. ✅ Analysis Repository stores and retrieves results
5. ✅ Ollama Client handles API errors gracefully
6. ✅ Review Handler validates input correctly
7. ✅ Mode switching updates review state

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./test/integration/...

# Test complete flow:
# 1. Create review
# 2. Generate preview analysis (calls real Ollama)
# 3. Retrieve cached analysis
# 4. Switch reading mode
```

---

## Success Metrics

This issue is complete when:

1. ✅ All database migrations run successfully
2. ✅ Review Service starts without errors
3. ✅ Health check endpoint returns 200 OK
4. ✅ User can create a new review via API
5. ✅ Preview Mode generates AI analysis (calls Ollama)
6. ✅ Analysis results persist in database
7. ✅ Cached analysis returned on subsequent requests
8. ✅ All unit tests pass with 70%+ coverage
9. ✅ Integration tests pass (if Ollama available)
10. ✅ No linting errors
11. ✅ CI/CD pipeline passes

---

## Cognitive Load Optimization Notes

### For Intrinsic Complexity (Simplify)
- AI integration is complex → Encapsulated in `OllamaClient`
- JSON parsing can fail → Fallback handling in `parsePreviewOutput`
- Clear service boundaries: Review Service ≠ Preview Service
- Models have clear responsibilities (Review, Repository, AnalysisResult)

### For Extraneous Load (Reduce)
- No magic strings: Use `ReadingMode` enum constants
- Explicit error messages: Wrap errors with context
- Consistent naming: `CreateReview`, `GetReview`, `AnalyzePreview`
- No global state: All dependencies injected via constructors

### For Germane Load (Maximize)
- Follows 3-layer architecture established in Portal service
- Respects bounded contexts (Review context only)
- Caching strategy clear (check DB before calling AI)
- Repository pattern enables testing without real database

---

## Questions and Clarifications

### Before Starting
- [x] Bounded context clear: Review service handles code analysis only
- [x] Dependencies understood: Requires Portal auth, Ollama running
- [x] AI integration pattern clear: OllamaClient → PreviewService → AnalysisRepository
- [x] Testing strategy defined: 70%+ coverage with mocked dependencies

### During Implementation

If you encounter:
- **Ollama not running** → Document setup in PR, tests should mock Ollama
- **JSON parsing errors** → Fallback to generic response, log error
- **Database migration conflicts** → Use timestamp-based migration names
- **GitHub API rate limits** → Cache repository metadata, don't fetch on every request
- **Slow AI responses** → Add timeout context (30 seconds), show loading spinner in UI

---

## References

- `ARCHITECTURE.md` - Review Service specification (lines 704-1031)
- `Requirements.md` - Five Reading Modes details (lines 180-450)
- `DevSmithTDD.md` - Testing strategy for Review Service
- Ollama API docs: https://github.com/ollama/ollama/blob/main/docs/api.md
- Go pgx library: https://github.com/jackc/pgx

---

**Next Steps (For Copilot):**
1. Create feature branch: `git checkout -b feature/004-copilot-review-service-preview-mode`
2. Read this spec completely (1500+ lines)
3. Follow implementation checklist phase by phase
4. **Commit after each phase** (8 commits expected)
5. Test after each phase: `go test ./...`
6. Push regularly: `git push` after every 2-3 commits
7. Create PR when all phases complete
8. Tag Claude for architecture review

**Estimated Time:** 90-120 minutes
**Test Coverage Target:** 70%+ (aim for 75%+)
**Success Metric:** User can create review, AI generates Preview Mode analysis, results display and persist
**Depends On:** Issue #003 (Portal Authentication) - User ID required for review ownership
