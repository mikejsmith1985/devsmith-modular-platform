#!/bin/bash

# OAuth Manual Verification Script
# Captures screenshots of the actual OAuth flow

set -e

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
SCREENSHOT_DIR="test-results/manual-verification-$(date +%Y%m%d)"
mkdir -p "$SCREENSHOT_DIR"

echo "========================================="
echo "OAuth Manual Verification Test"
echo "========================================="
echo "Screenshot directory: $SCREENSHOT_DIR"
echo ""

# Use Playwright's screenshot capability directly
cat > /tmp/oauth-test.js << 'EOJS'
const playwright = require('playwright');
const path = require('path');
const fs = require('fs');

const screenshotDir = process.argv[2];

(async () => {
  const browser = await playwright.chromium.launch();
  const context = await browser.newContext({ viewport: { width: 1280, height: 1024 } });
  const page = await context.newPage();
  
  console.log('\n[TEST] Step 1: Loading login page...');
  await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join(screenshotDir, '01-login-page.png'), fullPage: true });
  console.log('[TEST] ✅ Screenshot 1 saved: 01-login-page.png');
  
  console.log('\n[TEST] Step 2: Testing root page...');
  await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join(screenshotDir, '02-root-page.png'), fullPage: true });
  console.log('[TEST] ✅ Screenshot 2 saved: 02-root-page.png');
  
  console.log('\n[TEST] Step 3: Testing OAuth callback route directly...');
  await page.goto('http://localhost:3000/auth/github/callback?code=test&state=test', { waitUntil: 'networkidle' });
  const content = await page.content();
  await page.screenshot({ path: path.join(screenshotDir, '03-oauth-callback-direct.png'), fullPage: true });
  console.log('[TEST] ✅ Screenshot 3 saved: 03-oauth-callback-direct.png');
  
  // Check what was rendered
  const hasReactApp = content.includes('DevSmith Platform') || content.includes('<div id="root"');
  const hasBackendError = content.includes('OAUTH_STATE_INVALID') || content.includes('"error"');
  
  console.log('\n[TEST] Analysis:');
  console.log('[TEST]   - React app loaded:', hasReactApp);
  console.log('[TEST]   - Backend error shown:', hasBackendError);
  
  if (hasBackendError) {
    console.log('\n[TEST] ❌ FAILURE: Backend is handling OAuth callback instead of React!');
    console.log('[TEST] This means the legacy /auth/github/callback route is still active.');
    fs.writeFileSync(path.join(screenshotDir, 'ERROR.txt'), 'Backend route intercepting OAuth callback\n' + content.substring(0, 1000));
  } else if (hasReactApp) {
    console.log('\n[TEST] ✅ SUCCESS: React app is handling OAuth callback route!');
  } else {
    console.log('\n[TEST] ⚠️  WARNING: Unknown response');
    fs.writeFileSync(path.join(screenshotDir, 'UNKNOWN.txt'), content.substring(0, 1000));
  }
  
  console.log('\n[TEST] Step 4: Testing GitHub login flow initiation...');
  await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle' });
  
  // Set up console listener
  page.on('console', msg => {
    const text = msg.text();
    if (text.includes('[PKCE]') || text.includes('[OAuth]')) {
      console.log('[BROWSER]', text);
    }
  });
  
  // Click GitHub login button
  const githubButton = await page.locator('button:has-text("Login with GitHub")');
  if (await githubButton.isVisible()) {
    console.log('[TEST] Clicking GitHub login button...');
    
    // Wait for navigation to GitHub
    const navigationPromise = page.waitForURL(/github\.com/, { timeout: 5000 }).catch(() => null);
    await githubButton.click();
    
    const navigated = await navigationPromise;
    if (navigated) {
      console.log('[TEST] ✅ Redirected to GitHub OAuth');
      const url = page.url();
      console.log('[TEST] GitHub URL:', url);
      
      // Extract OAuth params
      const urlObj = new URL(url);
      console.log('[TEST] OAuth parameters:');
      console.log('[TEST]   - state:', urlObj.searchParams.get('state')?.substring(0, 30) + '...');
      console.log('[TEST]   - code_challenge:', urlObj.searchParams.get('code_challenge')?.substring(0, 30) + '...');
      
      await page.screenshot({ path: path.join(screenshotDir, '04-github-oauth-page.png'), fullPage: true });
      console.log('[TEST] ✅ Screenshot 4 saved: 04-github-oauth-page.png');
    } else {
      console.log('[TEST] ⚠️  Did not redirect to GitHub (might be expected in some setups)');
    }
  }
  
  await browser.close();
  
  console.log('\n=========================================');
  console.log('TEST COMPLETE');
  console.log('=========================================');
  console.log('Screenshots saved to:', screenshotDir);
  console.log('\nNext steps:');
  console.log('1. Review screenshots in', screenshotDir);
  console.log('2. Verify screenshot 03 shows React app (NOT JSON error)');
  console.log('3. If React app visible, OAuth routing is fixed ✅');
  console.log('4. If JSON error visible, routing still broken ❌\n');
})();
EOJS

node /tmp/oauth-test.js "$SCREENSHOT_DIR"
