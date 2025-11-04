# Playwright E2E Testing Setup

## Quick Setup Instructions

### Step 1: Configure Test Credentials

Edit `.env.playwright` and fill in your GitHub account credentials:

```bash
# Your GitHub login credentials (for automated testing)
GITHUB_TEST_USERNAME=YOUR_GITHUB_USERNAME_HERE
GITHUB_TEST_PASSWORD=YOUR_GITHUB_PASSWORD_HERE

# Already configured (OAuth app credentials)
GITHUB_CLIENT_ID=Ov23liaV4He3p1k7VziT
GITHUB_CLIENT_SECRET=3646316980e7f068a7a4d0bc21de510869251dc7
BASE_URL=http://localhost:3000
LOGGING_API=http://localhost:3000/api/logs
```

### Step 2: Run the Real User Flow Test

```bash
# Make sure services are running
docker-compose up -d

# Run the test (headed mode to watch it work)
npx playwright test tests/e2e/real-user-flow.spec.ts --headed

# Or run headless (faster)
npx playwright test tests/e2e/real-user-flow.spec.ts
```

### What the Test Does

1. ✅ Navigates to `http://localhost:3000/`
2. ✅ Clicks "Login with GitHub" button
3. ✅ Completes GitHub OAuth (automated or manual)
4. ✅ Lands on dashboard
5. ✅ Clicks Review card
6. ✅ Verifies navigation to Review app (NOT redirected back to GitHub)
7. ✅ **Captures all network traffic and console logs**
8. ✅ **Sends debug session to Logging service** for future debugging

## Alternative: Manual Login Mode

If you don't want to put your GitHub password in the config:

1. Leave `GITHUB_TEST_USERNAME` and `GITHUB_TEST_PASSWORD` empty in `.env.playwright`
2. Run test in headed mode: `npx playwright test tests/e2e/real-user-flow.spec.ts --headed`
3. Test will pause at GitHub login page - manually log in
4. Test continues automatically after OAuth callback

## Debugging Failed Tests

If the test fails, check:

1. **Debug output in console** - Shows all network requests, redirects, errors
2. **Logging service** - Test automatically sends debug session to `http://localhost:3000/api/logs/browser-debug`
3. **Cookie settings** - Test prints all cookies and their SameSite settings

## Security Note

⚠️ **NEVER commit `.env.playwright` to git!** 

It's already in `.gitignore`, but double-check:

```bash
# Verify it's ignored
git status | grep .env.playwright
# Should show nothing (ignored)
```

## What This Tests That Mock Tests Can't

- ✅ **Real GitHub OAuth flow** - Not mocked, actual GitHub redirect
- ✅ **Real cookie setting** - Portal sets cookie with correct SameSite
- ✅ **Real cookie sending** - Browser includes cookie in /review request
- ✅ **Real redirect behavior** - See if Review redirects or shows content
- ✅ **Real network timing** - Catch race conditions
- ✅ **Real browser security** - SameSite, CORS, etc.

This is the **only** way to test what the user actually experiences.
