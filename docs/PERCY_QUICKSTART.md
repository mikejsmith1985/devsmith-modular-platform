# Percy Visual Testing - Quick Start

## What is Percy?

Percy automatically captures screenshots of your app and compares them to baseline images to detect visual changes. It's perfect for catching UI regressions.

## Setup (5 minutes)

### 1. Get Your Percy Token

1. Go to https://percy.io and sign in
2. Click "Create new project"
3. Name it: `devsmith-modular-platform`
4. Go to Project Settings
5. Copy your `PERCY_TOKEN`

### 2. Add Token to Environment

```bash
# Add to .env file
echo "PERCY_TOKEN=your_token_here" >> .env
```

Or export it temporarily:
```bash
export PERCY_TOKEN=your_token_here
```

### 3. Run Percy Tests

```bash
# Make sure services are running
docker-compose up -d

# Option 1: Run with npm script (recommended)
npm run test:visual

# Option 2: Run with npx directly
npx percy exec -- npx playwright test tests/e2e/visual-regression.spec.ts

# Option 3: Run tests locally without Percy (for debugging)
npm run test:visual:local
```

**First run**: Percy will capture baseline screenshots
**Subsequent runs**: Percy will compare against baselines and flag any visual changes

## What Gets Tested

Our visual regression tests capture:

✅ **Portal Dashboard** (desktop, tablet, mobile)
✅ **Review Workspace** (all 5 reading modes)
✅ **Logs Dashboard** (empty state and with data)
✅ **Analytics Dashboard**
✅ **Dark mode variations**

## Viewing Results

After running tests:
1. Go to https://percy.io/devsmith-modular-platform
2. Click on the latest build
3. Review any visual changes
4. Approve or reject changes

## Troubleshooting

**Problem**: `PERCY_TOKEN not set`
**Solution**: Add token to `.env` or export it: `export PERCY_TOKEN=your_token`

**Problem**: Services not accessible
**Solution**: Start services: `docker-compose up -d`

**Problem**: "No snapshots captured"
**Solution**: Check Playwright tests are running: `npx playwright test --list`

## CI/CD Integration

### GitHub Actions

Add to your `.github/workflows/playwright.yml`:

```yaml
- name: Run Percy Visual Tests
  env:
    PERCY_TOKEN: ${{ secrets.PERCY_TOKEN }}
  run: |
    docker-compose up -d
    npm run test:visual
```

**Setup**:
1. Go to your GitHub repo → Settings → Secrets → Actions
2. Add new secret: `PERCY_TOKEN` with your token from percy.io
3. Percy will automatically run on PRs and compare visual changes

### Local Testing Without Percy

To run tests locally without sending snapshots to Percy:

```bash
npm run test:visual:local
```

This runs the same tests but skips Percy snapshot upload.

## Best Practices

✅ **DO**: Run Percy tests before merging PRs
✅ **DO**: Review all visual changes in Percy dashboard
✅ **DO**: Approve or reject changes explicitly
✅ **DO**: Use descriptive snapshot names

❌ **DON'T**: Auto-approve all changes without review
❌ **DON'T**: Skip visual tests for "small" changes
❌ **DON'T**: Commit PERCY_TOKEN to version control
