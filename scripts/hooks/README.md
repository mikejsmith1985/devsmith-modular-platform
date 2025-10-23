# Git Hooks

This directory contains Git hooks that can be installed into your local `.git/hooks/` directory.

## Installation

Run the install script from the repository root:

```bash
./scripts/install-hooks.sh
```

## Available Hooks

### pre-commit

Enhanced pre-commit hook with:
- Code formatting validation
- Linting (golangci-lint)
- Test coverage requirements (40% minimum)
- Security vulnerability scanning (govulncheck)
- TDD workflow awareness (RED/GREEN/REFACTOR)
- Import cycle detection
- Conditional race detection

See [PRE-COMMIT-ENHANCEMENTS.md](../../.docs/PRE-COMMIT-ENHANCEMENTS.md) for full documentation.

### Configuration

- **Team config**: `.pre-commit-config.yaml` (at repo root, committed)
- **Local config**: `.git/hooks/pre-commit-local.yaml` (not committed)

Copy the example to customize locally:
```bash
cp .git/hooks/pre-commit-local.yaml.example .git/hooks/pre-commit-local.yaml
```

## Testing

```bash
# Quick validation
.git/hooks/pre-commit --quick

# Standard validation (default)
.git/hooks/pre-commit --standard

# Thorough validation
.git/hooks/pre-commit --thorough
```
