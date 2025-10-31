# DevSmith Lint Fixer

Automated linting fixes for the DevSmith platform to maintain consistent code quality across all services.

## Purpose

This tool applies safe, automated fixes to common linting violations identified by golangci-lint:
- Missing package comments
- Magic string constants
- Improper HTTP request body handling
- Field alignment optimization
- And more

## Installation

```bash
go install ./cmd/devsmith-lint-fixer
```

## Usage

### Dry Run (Default - Shows what would be changed)

```bash
# Analyze and show proposed changes (no files modified)
devsmith-lint-fixer --all --path ./internal/ai

# Fix specific issues
devsmith-lint-fixer --fix-comments --path ./internal/security
devsmith-lint-fixer --fix-strings --path ./internal/ai
devsmith-lint-fixer --fix-http --path ./internal
```

### Apply Fixes

```bash
# Apply all fixes to files
devsmith-lint-fixer --all --path ./internal/ai --dry-run=false

# Apply specific fix types
devsmith-lint-fixer --fix-comments --fix-http --path ./internal --dry-run=false
```

## Fix Types

### --fix-comments
Adds missing package-level documentation comments for packages that lack them.

```go
// Before
package ai

// After
// Package ai provides AI provider abstraction, routing, and cost monitoring.
package ai
```

### --fix-strings
Extracts magic string literals that appear multiple times into named constants.

```go
// Before
if r.URL.Path == "/api/generate" { }
if r.URL.Path == "/api/generate" { }

// After
const OllamaGenerateEndpoint = "/api/generate"

if r.URL.Path == OllamaGenerateEndpoint { }
if r.URL.Path == OllamaGenerateEndpoint { }
```

### --fix-http
Replaces `nil` with `http.NoBody` for HTTP request bodies (Go 1.20+).

```go
// Before
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

// After
req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
```

## Integration with Pre-Push Hook

The pre-push hook runs `golangci-lint` on modified files. This tool can be used to automatically fix issues before they're caught:

```bash
# In your pre-commit/pre-push hook
devsmith-lint-fixer --all --path . --dry-run=false
go fmt ./...
golangci-lint run ./...
```

## Future Enhancements

- Auto-fix field alignment using betteralign
- Interactive mode for complex fixes
- CI integration for reporting
- Custom DevSmith-specific checks
- Integration with health check dashboard for code quality metrics

## Development

To test the tool:

```bash
# Build locally
go build -o devsmith-lint-fixer ./cmd/devsmith-lint-fixer

# Test on a directory
./devsmith-lint-fixer --all --path ./internal/ai
```

## Philosophy

- **Safe first**: Only apply fixes that are guaranteed to be correct
- **Transparent**: Always show what will be changed (dry-run by default)
- **Extensible**: Easy to add new fix types
- **Automated**: Prevents manual, repetitive fixes
