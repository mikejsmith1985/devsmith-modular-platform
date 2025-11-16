# Smoke Test Packages

Fast, focused smoke tests for rapid validation during development. Run only what you need.

## Quick Start

```bash
# Run all smoke tests (full validation)
npx playwright test tests/e2e/smoke/ --project=smoke --workers=6

# Run specific package during focused work
npx playwright test tests/e2e/smoke/ollama-integration/ --project=smoke --workers=4
npx playwright test tests/e2e/smoke/ui-rendering/ --project=smoke --workers=2
npx playwright test tests/e2e/smoke/full-suite/ --project=smoke --workers=6
```

## Test Packages

### 1. **ollama-integration/** - AI Model Analysis Features
**Purpose**: Validate that Ollama integration works for all 5 reading modes  
**Duration**: ~8-12 seconds  
**When to run**: After changes to:
- Review service mode handlers (preview, skim, scan, detailed, critical)
- Ollama client adapter
- AI analysis logic
- Mode buttons/UI in review app

**What it tests**:
- Review page loads and renders mode buttons
- Each mode button is clickable and triggers API request
- Ollama service returns valid analysis results
- Results display correctly in UI

**Skip if**: You're fixing dark mode, portal navigation, or analytics dashboards

```bash
npx playwright test tests/e2e/smoke/ollama-integration/ --project=smoke
```

---

### 2. **ui-rendering/** - Navigation & Theme
**Purpose**: Validate basic UI rendering, navigation, and dark mode toggle  
**Duration**: ~6-8 seconds  
**When to run**: After changes to:
- Navigation component (nav.templ)
- Dark mode toggle implementation
- Alpine.js integration
- Layout templates

**What it tests**:
- Portal page loads
- Navigation renders with correct links
- Dark mode toggle button is visible
- Dark mode toggle functionality (click, persistence, theme changes)
- Alpine.js attributes render correctly

**Skip if**: You're fixing Ollama integration, logs WebSocket, or analytics charts

```bash
npx playwright test tests/e2e/smoke/ui-rendering/ --project=smoke
```

---

### 3. **full-suite/** - Analytics & Logging Services
**Purpose**: Validate dashboard rendering and service health  
**Duration**: ~10-14 seconds  
**When to run**: After changes to:
- Analytics dashboard layout or filters
- Logs dashboard rendering
- Service health endpoints
- Dashboard CSS/styling

**What it tests**:
- Analytics dashboard loads and renders
- Chart.js library loads
- HTMX filters are functional
- Logs dashboard loads
- Log cards with Tailwind styling render
- Filter controls are present
- WebSocket connection indicators work

**Skip if**: You're fixing individual reading modes or dark mode

```bash
npx playwright test tests/e2e/smoke/full-suite/ --project=smoke
```

---

## Recommended Workflow

### Fixing Ollama Integration
```bash
# Only run tests that validate Ollama features
npx playwright test tests/e2e/smoke/ollama-integration/ --project=smoke --workers=4

# When confident, validate UI still works
npx playwright test tests/e2e/smoke/ui-rendering/ --project=smoke --workers=2

# Final validation - run all
npx playwright test tests/e2e/smoke/ --project=smoke --workers=6
```

### Fixing Dark Mode / Alpine.js
```bash
# Only run UI rendering tests
npx playwright test tests/e2e/smoke/ui-rendering/ --project=smoke

# Don't need to run Ollama or full-suite tests
```

### Fixing Analytics Dashboard
```bash
# Only run analytics tests
npx playwright test tests/e2e/smoke/full-suite/ --project=smoke

# Skip Ollama and UI rendering tests
```

### Pre-Push Validation (5-10 seconds)
```bash
# Run all smoke tests
npx playwright test tests/e2e/smoke/ --project=smoke --workers=6 --timeout=15000
```

---

## Test Duration Targets

| Package | Target | Typical | Max |
|---------|--------|---------|-----|
| ollama-integration | <12s | 8-10s | 15s |
| ui-rendering | <8s | 6-8s | 10s |
| full-suite | <14s | 10-12s | 15s |
| **all** | <30s | 20-25s | 30s |

---

## Scripts

Use convenience scripts to run packages by name:

```bash
# Run with prettier output
./scripts/validate-feature.sh ollama       # Run ollama-integration tests
./scripts/validate-feature.sh ui           # Run ui-rendering tests
./scripts/validate-feature.sh all          # Run everything
```

---

## Troubleshooting

### Tests timeout (> 15s)
- Check if services are healthy: `docker-compose ps`
- Check container logs: `docker-compose logs <service>`
- Reduce `--workers` to free up resources
- Ensure Ollama is running (for ollama-integration tests)

### Dark mode tests fail but Ollama tests pass
- You don't need to fix dark mode immediately - skip those tests
- Focus on Ollama, then come back to UI rendering

### Review loads fail but dark mode works
- Review service might be down - check logs
- Run only `ui-rendering` tests while fixing Review

---

## Adding New Tests

1. Determine which package it belongs to (ollama, ui, or full-suite)
2. Create the `.spec.ts` file in that package directory
3. Follow naming: `feature-name.spec.ts`
4. Add to this README under the appropriate section
5. Ensure test completes in < 5 seconds (< 15s total per package)

Example:
```bash
# New test for Preview mode specifically
touch tests/e2e/smoke/ollama-integration/preview-mode.spec.ts
```

---

## Performance Tips

- Use `--workers=2-4` for UI rendering tests (Alpine.js DOM manipulation)
- Use `--workers=6` for Ollama integration tests (API calls, parallel-safe)
- Use `--workers=6` for full-suite tests (independent dashboard tests)
- Don't use `--workers=8+` on WSL2 - diminishing returns
