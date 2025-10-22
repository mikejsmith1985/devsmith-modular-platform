# DevSmith Modular Platform

AI-supervised code review and analytics platform with modular microservices architecture.

## 📚 Documentation

### For Developers
- **[Pre-Commit Hook Guide](.docs/PRE-COMMIT-HOOK.md)** - Understanding validation output and fixing issues
- **[Architecture](ARCHITECTURE.md)** - System design and coding standards
- **[TDD Workflow](DevsmithTDD.md)** - Test-driven development approach
- **[Workflow Guide](.docs/WORKFLOW-GUIDE.md)** - Development process

### For AI Agents (Copilot)
- **[Copilot Instructions](.github/copilot-instructions.md)** - Step-by-step implementation guide
- **[Issue Templates](.docs/issues/)** - Feature specifications

### Additional Resources
- **[Troubleshooting](.docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Activity Log](.docs/devlog/copilot-activity.md)** - Development history

## 🚀 Quick Start

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd devsmith-modular-platform
   ```

2. **Understand the pre-commit hook**

   Every commit is automatically validated. Read [Pre-Commit Hook Guide](.docs/PRE-COMMIT-HOOK.md) to understand:
   - What gets checked (fmt, vet, lint, tests)
   - How to interpret the dashboard output
   - HIGH vs LOW priority issues
   - Using `--json` for detailed analysis

3. **Start development**

   See [Copilot Instructions](.github/copilot-instructions.md) for the full workflow.

## 🛡️ Quality Gates

All commits must pass:
- ✅ Code formatting (`go fmt`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint`)
- ✅ Tests (`go test -short`)

The pre-commit hook blocks commits automatically if validation fails.

## 📖 Key Documents

- **PRE-COMMIT-HOOK.md** - You're here! Start here to understand validation
- **ARCHITECTURE.md** - System design, standards, patterns
- **copilot-instructions.md** - Step-by-step implementation guide
