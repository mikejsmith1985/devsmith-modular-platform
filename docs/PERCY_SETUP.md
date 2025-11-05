# Percy Visual Regression Testing Setup

## Overview
Percy provides automated visual regression testing by capturing screenshots of your application and comparing them across builds to detect visual changes.

## Setup Steps

### 1. Create Percy Account
1. Visit https://percy.io/
2. Sign up with GitHub account
3. Create new project: "DevSmith Modular Platform"
4. Note your project token

### 2. Configure Percy Token
Add to your environment:

```bash
# Option A: Local development (.bashrc or .zshrc)
export PERCY_TOKEN="your_token_here"

# Option B: Docker Compose (docker-compose.yml)
# Add to portal service environment:
environment:
  - PERCY_TOKEN=${PERCY_TOKEN}
```

### 3. Run Visual Tests
```bash
# Run all visual regression tests
npx percy exec -- npx playwright test tests/e2e/visual-regression.spec.ts

# Run with specific browser
npx percy exec -- npx playwright test visual-regression.spec.ts --project=chromium
```

## Test Coverage

Our visual regression suite (`tests/e2e/visual-regression.spec.ts`) covers:

### Portal Service
- Dashboard (desktop, tablet, mobile)
- Dark mode variations

### Review Service  
- Workspace (all reading modes: preview, skim, scan, detailed, critical)
- Empty state
- Desktop, tablet, mobile views

### Logs Service
- Dashboard (empty and with data)
- Real-time log streaming
- Responsive views

### Analytics Service
- Dashboard
- All viewport sizes

## Configuration

Percy configuration is in `.percy.yml`:

```yaml
snapshot:
  widths: [375, 768, 1920]  # Mobile, tablet, desktop
  minHeight: 1024
  enableJavaScript: true     # For HTMX interactions
  responsiveSnapshotCapture: true
  networkIdleTimeout: 500    # Wait for HTMX updates
```

## Workflow

1. **First Run**: Establishes baseline snapshots
2. **Subsequent Runs**: Compares against baseline, highlights differences
3. **Review**: Visit Percy dashboard to approve/reject changes
4. **CI Integration**: Add to GitHub Actions:

```yaml
- name: Percy Visual Tests
  env:
    PERCY_TOKEN: ${{ secrets.PERCY_TOKEN }}
  run: |
    npx percy exec -- npx playwright test visual-regression.spec.ts
```

## Best Practices

### When to Update Baselines
- ✅ Intentional UI changes (new features, design updates)
- ✅ After reviewing and approving differences
- ❌ Don't approve without reviewing (defeats the purpose)

### Snapshot Strategy
- Capture key user workflows (login, dashboard, feature pages)
- Include responsive breakpoints (mobile, tablet, desktop)
- Test dark mode separately
- Capture empty states and error states

### Performance
- Percy snapshots add ~30-60 seconds to test runs
- Run separately from smoke tests: `npm run test:visual`
- Consider parallel execution for large test suites

## Troubleshooting

### "PERCY_TOKEN not set"
```bash
# Check if token is configured
echo $PERCY_TOKEN

# If empty, set it:
export PERCY_TOKEN="your_token_here"
```

### "Build not found"
- Ensure Percy project is created
- Check token matches project
- Verify network connectivity

### Visual Diffs Detected
1. Review Percy dashboard: https://percy.io/
2. Compare baseline vs. new snapshot
3. If intentional change: **Approve**
4. If regression: **Reject** and fix code
5. New baseline becomes reference for future builds

## Integration Status

- ✅ Percy CLI installed (`@percy/playwright`, `@percy/cli`)
- ✅ Configuration file created (`.percy.yml`)
- ✅ Test suite implemented (`visual-regression.spec.ts`)
- ⏳ Percy account setup (requires manual action)
- ⏳ Token configuration (requires Percy account)
- ⏳ Baseline snapshots capture (run after token setup)

## Next Steps

1. Create Percy account at https://percy.io/
2. Get project token
3. Set `PERCY_TOKEN` environment variable
4. Run: `npx percy exec -- npx playwright test visual-regression.spec.ts`
5. Review and approve baseline snapshots in Percy dashboard
6. Add Percy token to GitHub Secrets for CI integration

## Resources

- Percy Documentation: https://docs.percy.io/
- Percy Playwright Integration: https://docs.percy.io/docs/playwright
- Percy Pricing: https://percy.io/pricing (free for open source)
