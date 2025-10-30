# Duplicate Code Detection

## Overview

The `scripts/check-duplicates.sh` tool helps detect duplicate code blocks BEFORE the linter complains. This enables proactive refactoring and code quality improvements.

## Why Detect Duplicates Early?

- **Maintainability**: Duplicated code is harder to maintain (fix bug in one place, forget the other)
- **Quality**: Duplicates hide refactoring opportunities
- **Testing**: Reduces test surface area (test once, use everywhere)
- **Linting**: Fails CI if duplicates remain (golangci-lint)

## Running the Tool

### Manual Detection

```bash
# Quick scan for obvious duplicates
bash scripts/check-duplicates.sh

# Output shows potential duplicates and recommendations
```

### Integrate into Pre-Push Hook

Add to `.git/hooks/pre-push`:

```bash
#!/bin/bash

# Run duplicate check
bash scripts/check-duplicates.sh
if [ $? -ne 0 ]; then
    echo "❌ Duplicate code detected - fix before pushing"
    exit 1
fi

# Continue with other checks...
```

### Advanced: Use dupl Tool

For comprehensive analysis, install the `dupl` tool:

```bash
go install github.com/remyoudompheng/dupl@latest

# Run comprehensive scan (default similarity threshold)
dupl ./internal ./apps ./cmd

# Use custom threshold (default: 3-4 lines)
dupl -t 5 ./internal
```

## How Duplicates Happen

1. **Copy-Paste**: Fastest way to code (also fastest to create bugs)
2. **Similar Requirements**: Two handlers that parse limit differently
3. **Database Patterns**: Same query logic with different WHERE clauses
4. **Error Handling**: Same defer/cleanup patterns repeated

## Common Duplicate Patterns

### Pattern 1: Query Wrapper Functions

```go
// BEFORE (Duplicate)
func GetRecentChecks(ctx context.Context, limit int) {
    rows, err := db.QueryContext(ctx, query1, limit)
    defer rows.Close()
    for rows.Next() {
        var item Item
        rows.Scan(&item.Field1, &item.Field2)
        items = append(items, item)
    }
    return items
}

func GetOldChecks(ctx context.Context, limit int) {
    rows, err := db.QueryContext(ctx, query2, limit)
    defer rows.Close()  // DUPLICATE PATTERN
    for rows.Next() {   // DUPLICATE PATTERN
        var item Item
        rows.Scan(&item.Field1, &item.Field2) // DUPLICATE
        items = append(items, item)
    }
    return items
}

// AFTER (Refactored)
func queryItems(ctx context.Context, query string, arg interface{}) ([]Item, error) {
    rows, err := db.QueryContext(ctx, query, arg)
    defer rows.Close()
    var items []Item
    for rows.Next() {
        var item Item
        rows.Scan(&item.Field1, &item.Field2)
        items = append(items, item)
    }
    return items, rows.Err()
}

func GetRecentChecks(ctx context.Context, limit int) {
    return queryItems(ctx, query1, limit)
}

func GetOldChecks(ctx context.Context, limit int) {
    return queryItems(ctx, query2, limit)
}
```

### Pattern 2: Parameter Parsing

```go
// BEFORE (Duplicate)
func Handler1(c *gin.Context) {
    limit := 50
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
            limit = l
        }
    }
    // Use limit...
}

func Handler2(c *gin.Context) {
    limit := 50  // DUPLICATE
    if limitStr := c.Query("limit"); limitStr != "" {  // DUPLICATE
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {  // DUPLICATE
            limit = l  // DUPLICATE
        }
    }
    // Use limit...
}

// AFTER (Refactored)
func parseLimit(c *gin.Context, defaultLimit, maxLimit int) int {
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= maxLimit {
            return l
        }
    }
    return defaultLimit
}

func Handler1(c *gin.Context) {
    limit := parseLimit(c, 50, 1000)
    // Use limit...
}

func Handler2(c *gin.Context) {
    limit := parseLimit(c, 50, 1000)
    // Use limit...
}
```

## Detection Threshold

The tool considers >10 lines as potential duplicates. For stricter detection, use dupl:

- **Default (dupl)**: 3-4 lines
- **Custom threshold**: `dupl -t 5` checks for 5+ line duplicates

## DevSmith Integration

**Recommended workflow:**

1. **Before Writing Code**: Check similar functions for duplicates
2. **During Development**: Run `check-duplicates.sh` before commit
3. **Before Push**: Pre-push hook runs full validation
4. **CI Validation**: golangci-lint catches any missed duplicates

## False Positives

The tool may flag similar patterns that aren't truly duplicates:

✅ **True duplicates** (fix these):
- Identical error handling in multiple functions
- Copy-pasted business logic
- Same database query patterns

❌ **False positives** (ignore these):
- Similar but unrelated code (different purposes)
- Coincidental structural similarities
- Different error handling strategies

## Best Practices

### DO:
- ✅ Extract helpers for >10 line patterns
- ✅ Use helper functions for database queries
- ✅ Consolidate parameter parsing
- ✅ Share error handling patterns

### DON'T:
- ❌ Create functions that are too generic (lose readability)
- ❌ Force refactoring of truly independent code
- ❌ Over-abstract small utilities

## Examples from Phase 3

### Fixed Duplicate #1

**File**: `internal/logs/services/health_storage_service.go`

**Duplicate pattern**: GetRecentChecks + GetCheckHistory (52 lines)

**Solution**: Created `queryHealthChecks()` helper

**Lines saved**: 52 → 2 per function + 1 helper

### Fixed Duplicate #2

**File**: `cmd/logs/handlers/health_history_handler.go`

**Duplicate pattern**: GetHealthHistory + GetRepairHistory (24 lines)

**Solution**: Created `parseLimit()` + `sendJSONResponse()` helpers

**Lines saved**: 24 → 2 per function + 2 helpers

## Troubleshooting

### Script not running?

```bash
chmod +x scripts/check-duplicates.sh
bash scripts/check-duplicates.sh
```

### Want more detailed output?

```bash
bash scripts/check-duplicates.sh --verbose
```

### Want to use dupl but it's not installed?

```bash
go install github.com/remyoudompheng/dupl@latest
```

## References

- [DRY Principle](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself)
- [golangci-lint documentation](https://golangci-lint.run/)
- [dupl GitHub](https://github.com/remyoudompheng/dupl)
