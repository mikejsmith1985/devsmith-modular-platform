# OpenHands Implementation Spec Template

**Created:** [DATE]
**Issue:** #[ISSUE_NUMBER]
**Estimated Complexity:** [Low | Medium | High]
**Target Service:** [portal | review | logs | analytics | build]

---

## Overview

### Feature Description
[1-2 sentence description of what needs to be built]

### User Story
As a [type of user], I want [goal] so that [benefit/value].

### Success Criteria
- [ ] [Specific, measurable outcome 1]
- [ ] [Specific, measurable outcome 2]
- [ ] [Specific, measurable outcome 3]

---

## Context for Cognitive Load Management

### Bounded Context
**Service:** [Service name]
**Domain:** [Business domain within service]
**Related Entities:**
- `Entity1` - [Purpose in this context]
- `Entity2` - [Purpose in this context]

**Context Boundaries:**
- ✅ **Within scope:** [What concerns belong here]
- ❌ **Out of scope:** [What concerns should NOT be here]

**Example:**
```
Service: Portal
Domain: Authentication
Related Entities:
  - User (authentication identity)
  - Session (active login session)
  - GitHubToken (OAuth credential)

Within scope: Login, logout, session management
Out of scope: User profile editing (that's a different context)
```

---

### Layering

**Primary Layer:** [Controller | Orchestration | Data]
**Layer Responsibilities:**

#### Controller Layer Files (if applicable)
```
handlers/
├── [handler_name].go       # HTTP request/response handling
└── [handler_name]_test.go  # Handler tests

templates/
├── [template_name].templ   # UI rendering
└── components/
    └── [component].templ   # Reusable UI components
```

#### Orchestration Layer Files (if applicable)
```
services/
├── [service_name].go       # Business logic
└── [service_name]_test.go  # Service tests

interfaces/
└── [interface_name].go     # Abstract contracts
```

#### Data Layer Files (if applicable)
```
db/
├── [repository_name].go       # Database queries
├── [repository_name]_test.go  # Repository tests
└── migrations/
    └── [timestamp]_[description].sql
```

**Cross-Layer Rules:**
- ✅ Controllers may call Services
- ✅ Services may call Repositories
- ❌ Controllers MUST NOT call Repositories directly
- ❌ Data layer MUST NOT call Services or Controllers
- ❌ No circular dependencies between layers

---

### Abstractions to Implement

**New Interfaces (if any):**
```go
// interfaces/[interface_name].go
type [InterfaceName] interface {
    // [MethodName] - [What it does and why]
    [MethodName](ctx context.Context, [params]) ([return], error)
}
```

**Existing Interfaces to Use:**
- `[InterfaceName]` from `[package]` - [Why we're using this abstraction]

**Implementation Strategy:**
1. Define interface first (abstraction)
2. Create concrete implementation (concretion)
3. Test against interface, not concrete type
4. Future implementations can swap in without breaking consumers

---

### Scope Management

**Global/Package-Level State:**
```go
// AVOID if possible. Only use for:
// - Configuration loaded at startup
// - Read-only reference data
// - Singleton connections (DB pool, Redis client)

// If you must use package-level state, document it here:
// var [globalVar] *[Type]  // Purpose: [why this needs to be global]
```

**Struct-Level State:**
```go
// Preferred: Encapsulate dependencies in structs

type [ServiceName] struct {
    // Dependencies passed via constructor
    repo       [RepositoryInterface]
    aiClient   *ollama.Client
    logger     *log.Logger

    // Configuration
    config     *Config
}
```

**Function-Level Scope:**
```go
// Keep variables as local as possible
// Pass dependencies explicitly
// Minimize side effects
```

---

## Implementation Details

### 1. Database Changes

#### Schema: `[schema_name]`

**New Tables:**
```sql
CREATE TABLE [schema].[table_name] (
    id SERIAL PRIMARY KEY,
    [column_name] [TYPE] [CONSTRAINTS],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_[table]_[column] ON [schema].[table_name]([column]);

-- Comments
COMMENT ON TABLE [schema].[table_name] IS '[Purpose of this table]';
COMMENT ON COLUMN [schema].[table_name].[column] IS '[Purpose of this column]';
```

**Migrations:**
- Migration file: `[timestamp]_[description].sql`
- Rollback file: `[timestamp]_[description]_down.sql`

**Data Relationships:**
```
[Table1] 1:N [Table2]
  - [Table1].id → [Table2].[table1_id]
  - Relationship meaning: [Describe the relationship]

No cross-schema foreign keys!
If referencing another schema, store ID but no FK constraint.
```

---

### 2. Go Structs and Models

```go
// models/[model_name].go

// [ModelName] represents [business concept]
// Bounded Context: [context name]
// Layer: [which layer this model belongs to]
type [ModelName] struct {
    ID        int       `json:"id" db:"id"`
    [Field]   [Type]    `json:"[field]" db:"[field]" binding:"[validation]"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validation rules for Gin binding
// - required: Field cannot be empty
// - email: Must be valid email format
// - min=X: Minimum length
// - max=X: Maximum length
// - oneof=a b c: Must be one of these values
```

---

### 3. API Endpoints

```go
// handlers/[handler_name].go

// Handle[Action] [description of what this handler does]
// Method: [GET|POST|PUT|PATCH|DELETE]
// Path: /api/[service]/[resource]/[...params]
// Auth: [Required | Optional | Public]
func Handle[Action](c *gin.Context) {
    // 1. Parse and validate input
    var req [RequestType]
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request",
            "details": err.Error(),
        })
        return
    }

    // 2. Call service layer (business logic)
    result, err := [service].[Method](c.Request.Context(), req)
    if err != nil {
        // Log error with context
        log.Error().
            Err(err).
            Str("endpoint", "[endpoint_name]").
            Msg("[Action] failed")

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "[User-friendly error message]",
        })
        return
    }

    // 3. Return success response
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    result,
    })
}
```

**Endpoint Specification:**
```
POST /api/[service]/[resource]

Request Body:
{
    "[field]": "[type]",  // [description]
    "[field]": "[type]"   // [description]
}

Response (200 OK):
{
    "success": true,
    "data": {
        "[field]": "[type]",  // [description]
    }
}

Response (400/500 Error):
{
    "error": "User-friendly message",
    "details": "Developer-friendly details" // Only in dev mode
}
```

---

### 4. Service Layer Implementation

```go
// services/[service_name].go

type [ServiceName] struct {
    repo   [RepositoryInterface]
    // ... other dependencies
}

// New[ServiceName] creates a new [ServiceName]
// Dependencies are passed explicitly (no globals!)
func New[ServiceName](repo [RepositoryInterface]) *[ServiceName] {
    return &[ServiceName]{
        repo: repo,
    }
}

// [MethodName] [what this method does and why]
// This is where business logic lives
func (s *[ServiceName]) [MethodName](ctx context.Context, [params]) ([result], error) {
    // 1. Validate business rules
    if [validation_check] {
        return nil, errors.New("[business rule violation]")
    }

    // 2. Call repository (data layer)
    data, err := s.repo.[RepoMethod](ctx, [params])
    if err != nil {
        return nil, fmt.Errorf("[operation] failed: %w", err)
    }

    // 3. Transform data (if needed)
    result := transform(data)

    // 4. Call external services (if needed)
    // - AI services
    // - GitHub API
    // - Logging service

    // 5. Save results
    if err := s.repo.[SaveMethod](ctx, result); err != nil {
        return nil, fmt.Errorf("save failed: %w", err)
    }

    return result, nil
}
```

---

### 5. Data Layer Implementation

```go
// db/[repository_name].go

type [RepositoryName] struct {
    db *sql.DB  // Or *pgxpool.Pool for pgx
}

// New[RepositoryName] creates a new [RepositoryName]
func New[RepositoryName](db *sql.DB) *[RepositoryName] {
    return &[RepositoryName]{db: db}
}

// [MethodName] [what this method does]
// SQL queries ONLY in this layer
func (r *[RepositoryName]) [MethodName](ctx context.Context, [params]) ([result], error) {
    query := `
        SELECT id, [columns]
        FROM [schema].[table]
        WHERE [condition]
        ORDER BY [column] DESC
        LIMIT $1 OFFSET $2
    `

    rows, err := r.db.QueryContext(ctx, query, [params])
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()

    var results []*[ModelType]
    for rows.Next() {
        var item [ModelType]
        err := rows.Scan(&item.Field1, &item.Field2, ...)
        if err != nil {
            return nil, fmt.Errorf("scan failed: %w", err)
        }
        results = append(results, &item)
    }

    return results, nil
}
```

**SQL Query Guidelines:**
- ✅ Use parameterized queries (`$1`, `$2`) - prevents SQL injection
- ✅ Use context for cancellation/timeouts
- ✅ Close rows with `defer rows.Close()`
- ✅ Check `rows.Err()` after iteration
- ❌ Never concatenate user input into SQL
- ❌ No business logic in SQL (keep it in service layer)

---

### 6. Template Implementation (if UI)

```go
// templates/[template_name].templ

package templates

import (
    "github.com/mikejsmith1985/devsmith-platform/apps/[service]/models"
)

// [TemplateName] renders [description]
// Props are typed - compile-time safety!
templ [TemplateName]([param1] [Type], [param2] [Type]) {
    @Layout("[PageTitle]") {
        <div class="container mx-auto p-4">
            <h1 class="text-2xl font-bold mb-4">
                {[param1].Title}
            </h1>

            if len([param2]) == 0 {
                <p class="text-gray-500">No items found</p>
            } else {
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                    for _, item := range [param2] {
                        @[ComponentName](item)
                    }
                </div>
            }
        </div>
    }
}

// [ComponentName] renders [description]
// Reusable component
templ [ComponentName](item *models.[ModelType]) {
    <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
            <h2 class="card-title">{item.Title}</h2>
            <p>{item.Description}</p>

            <!-- HTMX for interactivity -->
            <button
                class="btn btn-primary"
                hx-post={"/api/[service]/[resource]/" + item.ID}
                hx-target="#result"
                hx-swap="innerHTML"
            >
                Action
            </button>
        </div>
    </div>
}
```

**HTMX Integration:**
- `hx-get/post/put/delete`: HTTP method and URL
- `hx-target`: Where to put response
- `hx-swap`: How to insert (innerHTML, outerHTML, beforeend)
- `hx-trigger`: What triggers request (click, load, every 2s)

---

### 7. Testing Requirements

#### Unit Tests (70%+ coverage)

```go
// services/[service_name]_test.go

func Test[ServiceName]_[MethodName](t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{}
    service := New[ServiceName](mockRepo)
    ctx := context.Background()

    // Mock expected calls
    mockRepo.On("[Method]", ctx, mock.Anything).
        Return([expectedResult], nil)

    // Act
    result, err := service.[MethodName](ctx, [params])

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, [expectedValue], result.[Field])
    mockRepo.AssertExpectations(t)
}

func Test[ServiceName]_[MethodName]_ErrorCase(t *testing.T) {
    // Test error handling
    mockRepo := &MockRepository{}
    service := New[ServiceName](mockRepo)
    ctx := context.Background()

    // Mock error
    mockRepo.On("[Method]", ctx, mock.Anything).
        Return(nil, errors.New("database error"))

    // Act
    result, err := service.[MethodName](ctx, [params])

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "database error")
}
```

#### Integration Tests

```go
// tests/integration/[feature]_test.go
// +build integration

func TestIntegration[Feature](t *testing.T) {
    // Requires real database
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    // Test full flow
    // ...
}
```

#### Test Coverage Commands
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Implementation Checklist

### Phase 1: Setup
- [ ] Create branch: `feature/[issue-number]-[description]`
- [ ] Create database migration files
- [ ] Run migrations locally: `go run db/migrate.go up`
- [ ] Define Go structs/models

### Phase 2: Data Layer
- [ ] Create repository interface (abstraction)
- [ ] Implement repository (concretion)
- [ ] Write repository tests
- [ ] Verify tests pass: `go test ./db/...`

### Phase 3: Service Layer
- [ ] Create service interface (if needed)
- [ ] Implement service with business logic
- [ ] Write service tests (mock repository)
- [ ] Verify tests pass: `go test ./services/...`

### Phase 4: Controller Layer
- [ ] Create HTTP handlers
- [ ] Create Templ templates (if UI)
- [ ] Write handler tests
- [ ] Verify tests pass: `go test ./handlers/...`

### Phase 5: Integration
- [ ] Register routes in `main.go`
- [ ] Update nginx config (if new routes)
- [ ] Test through gateway: `http://localhost:3000`
- [ ] Verify HTMX interactions work
- [ ] Check error handling (try invalid inputs)

### Phase 6: Documentation
- [ ] Update AI_CHANGELOG.md
- [ ] Add inline comments for complex logic
- [ ] Update API documentation (Swagger if applicable)

### Phase 7: Code Quality
- [ ] Run linter: `golangci-lint run`
- [ ] Fix any linting issues
- [ ] Format code: `gofmt -w .`
- [ ] Check test coverage: `go test -cover ./...`
- [ ] Ensure 70%+ coverage

### Phase 8: Commit and PR
- [ ] Stage changes: `git add .`
- [ ] Commit with conventional commit message:
      ```
      feat([service]): [description]

      - [What was added/changed]
      - [Why it was needed]

      Closes #[issue-number]
      ```
- [ ] Push to GitHub
- [ ] Create PR to `development`
- [ ] Fill out PR template completely
- [ ] Check acceptance criteria from issue

---

## Cognitive Load Optimization Notes

### For Intrinsic Complexity (Simplify)
- Use clear naming: `getUserByEmail` not `get`
- Break complex functions into smaller ones
- Add comments explaining "why" not "what"
- Use Go idioms (explicit error handling)

### For Extraneous Load (Reduce)
- No magic numbers - use named constants
- No global mutable state
- Explicit dependencies (pass via constructor)
- Clear error messages

### For Germane Load (Maximize)
- Follow existing patterns in codebase
- Respect bounded contexts and layering
- Use abstractions consistently
- Document architectural decisions

---

## Questions and Clarifications

### Before Starting Implementation
- [ ] Is the bounded context clear?
- [ ] Are the layering boundaries understood?
- [ ] Are all dependencies identified?
- [ ] Are acceptance criteria measurable?

### During Implementation
If you encounter:
- **Unclear requirements** → Ask in issue comments
- **Cross-context dependencies** → Red flag! Discuss with architect
- **Layer violations** → Refactor to respect separation
- **Scope creep** → Split into separate issues

---

## References
- ARCHITECTURE.md - Mental Models section
- ARCHITECTURE.md - Service Architecture section
- DevSmithRoles.md - Hybrid workflow
- DevsmithTDD.md - TDD approach
- Go documentation: https://go.dev/doc/

---

**Next Steps:**
1. Read this spec completely
2. Ask clarifying questions in issue #[number]
3. Follow the implementation checklist
4. Create PR when complete
