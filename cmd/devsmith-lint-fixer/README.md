# devsmith-lint-fixer

Automated linting tool for the DevSmith platform designed to prevent and fix common code quality issues.

## Purpose

This tool addresses recurring linting issues that appear across the DevSmith codebase, enabling:
- **Automated Fixes**: Safe, common issues fixed automatically
- **Issue Detection**: Identifies categories of issues for developer review
- **Prevention**: Integrates with pre-push hooks to prevent regressions

## Usage

### Report Mode (Analyze Only)
```bash
go run ./cmd/devsmith-lint-fixer -report -path ./internal/ai
```

### Fix Mode (Apply Safe Fixes)
```bash
go run ./cmd/devsmith-lint-fixer -fix -path ./internal
```

## Supported Fixes

### Automatic Fixes
1. **Missing Package Comments** - Adds `// Package X` comments
2. **Empty String Tests** - Converts `len(s) > 0` to `s != ""`
3. **HTTP nil Body** - Suggests `http.NoBody` instead of `nil`

### Detected But Manual
1. **Field Alignment** - Recommend `betteralign -apply`
2. **Variable Shadowing** - Requires manual rename
3. **Repeated Strings** - Extract to constants
4. **Naming Conventions** - Type prefixes (AIProvider → Provider)

## Integration

### Pre-Push Hook
Add to `.git/hooks/pre-push`:
```bash
go run ./cmd/devsmith-lint-fixer -report -path ./internal
```

### CI/CD Pipeline
```bash
devsmith-lint-fixer -report -path ./
exit $?  # Fail if issues found
```

## Future Enhancements

1. **Config File Support** - `.devsmith-lint.yaml` for rules
2. **Custom Rules Engine** - User-defined fix patterns
3. **Integration Dashboard** - Track issues over time
4. **Automated Fixes for Naming** - AIProvider → Provider refactoring
5. **Shadow Variable Fixer** - Automatic variable renaming

## Related Tools

- `betteralign` - Fixes struct field alignment
- `gofmt` - Code formatting
- `goimports` - Import management
- `golangci-lint` - Comprehensive linting
