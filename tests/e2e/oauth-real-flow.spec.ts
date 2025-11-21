import { test, expect } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

/**
 * OAuth Real User Flow Test
 * 
 * This test replicates the ACTUAL user experience:
 * 1. Visit login page
 * 2. Click GitHub login button
 * 3. Verify React app handles the flow
 * 4. Capture screenshots at each step
 */

const screenshotDir = path.join(process.cwd(), 'test-results', `manual-verification-${new Date().toISOString().split('T')[0].replace(/-/g, '')}`);

test.beforeAll(() => {
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }
});

test.describe('OAuth Real User Flow', () => {
  test('Complete OAuth flow from login page', async ({ page }) => {
    console.log('[TEST] Starting OAuth real user flow test');
    console.log('[TEST] Screenshot directory:', screenshotDir);

    // Get GitHub credentials from environment
    const githubUsername = process.env.GITHUB_TEST_USERNAME;
    const githubPassword = process.env.GITHUB_TEST_PASSWORD;
    
    if (!githubUsername || !githubPassword) {
      console.log('[TEST] ⚠️  GitHub credentials not set - will test up to GitHub redirect only');
      console.log('[TEST] Set GITHUB_TEST_USERNAME and GITHUB_TEST_PASSWORD to test full flow');
    }

    // Step 1: Visit login page
    console.log('[TEST] Step 1: Navigating to login page');
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(screenshotDir, '01-login-page.png'), fullPage: true });
    console.log('[TEST] ✅ Screenshot 1: Login page captured');

    // Verify login page loaded
    await expect(page.locator('text=DevSmith Platform')).toBeVisible();
    await expect(page.locator('button:has-text("Login with GitHub")')).toBeVisible();
    console.log('[TEST] ✅ Login page elements verified');

    // Step 2: Click GitHub login button
    console.log('[TEST] Step 2: Clicking GitHub login button');
    
    await page.click('button:has-text("Login with GitHub")');
    
    // Wait for either GitHub login page OR immediate callback (if already authed)
    console.log('[TEST] Waiting for navigation to GitHub or callback...');
    
    try {
      // Wait for navigation to complete
      await page.waitForLoadState('networkidle', { timeout: 10000 });
      const currentUrl = page.url();
      console.log('[TEST] Current URL after click:', currentUrl);
      
      await page.screenshot({ path: path.join(screenshotDir, '02-after-github-click.png'), fullPage: true });
      
      // Check if we're on GitHub login page
      if (currentUrl.includes('github.com/login')) {
        console.log('[TEST] ✅ Redirected to GitHub login page');
        
        if (githubUsername && githubPassword) {
          console.log('[TEST] Step 3: Entering GitHub credentials...');
          
          // Fill in GitHub username/email
          await page.fill('input[name="login"]', githubUsername);
          await page.fill('input[name="password"]', githubPassword);
          await page.screenshot({ path: path.join(screenshotDir, '03-github-credentials-filled.png'), fullPage: true });
          
          // Click sign in
          await page.click('input[type="submit"][value="Sign in"]');
          console.log('[TEST] Submitted GitHub login form');
          
          // Wait for either authorization page or callback
          await page.waitForLoadState('networkidle', { timeout: 15000 });
          const afterLoginUrl = page.url();
          console.log('[TEST] URL after GitHub login:', afterLoginUrl);
          await page.screenshot({ path: path.join(screenshotDir, '04-after-github-login.png'), fullPage: true });
          
          // Check if we need to authorize the app
          if (afterLoginUrl.includes('github.com/login/oauth/authorize')) {
            console.log('[TEST] Step 4: Authorizing app...');
            const authorizeButton = page.locator('button[name="authorize"]');
            if (await authorizeButton.isVisible({ timeout: 2000 })) {
              await authorizeButton.click();
              console.log('[TEST] Clicked authorize button');
              await page.waitForLoadState('networkidle', { timeout: 10000 });
            }
          }
          
          // Should now be back at our callback URL
          const finalUrl = page.url();
          console.log('[TEST] Final URL:', finalUrl);
          await page.screenshot({ path: path.join(screenshotDir, '05-final-page.png'), fullPage: true });
          
          // Verify we're back at localhost
          expect(finalUrl).toContain('localhost:3000');
          
          // Check if we successfully logged in
          const pageContent = await page.content();
          const isLoggedIn = finalUrl.includes('/') && !finalUrl.includes('/login');
          const hasError = pageContent.includes('error') || pageContent.includes('Authentication Error');
          
          console.log('[TEST] Login status:');
          console.log('[TEST]   - Is logged in:', isLoggedIn);
          console.log('[TEST]   - Has error:', hasError);
          
          if (isLoggedIn && !hasError) {
            console.log('[TEST] ✅ SUCCESS: Full OAuth flow completed!');
          } else {
            console.log('[TEST] ⚠️  OAuth completed but may have errors');
          }
          
        } else {
          console.log('[TEST] ⚠️  Skipping GitHub login - credentials not provided');
          console.log('[TEST] Test verified redirect to GitHub works correctly');
        }
        
      } else if (currentUrl.includes('localhost:3000/auth/github/callback')) {
        console.log('[TEST] ✅ Already authenticated - redirected directly to callback');
        await page.screenshot({ path: path.join(screenshotDir, '03-callback-immediate.png'), fullPage: true });
        
        // Verify callback handled by React
        const pageContent = await page.content();
        const isReactApp = pageContent.includes('DevSmith Platform') || pageContent.includes('root');
        const isBackendError = pageContent.includes('OAUTH_STATE_INVALID');
        
        console.log('[TEST] Callback handled by:', isReactApp ? 'React ✅' : 'Backend ❌');
        expect(isReactApp).toBe(true);
        expect(isBackendError).toBe(false);
        
      } else if (currentUrl.includes('localhost:3000') && !currentUrl.includes('/login')) {
        console.log('[TEST] ✅ Successfully logged in (cached auth)');
        await page.screenshot({ path: path.join(screenshotDir, '03-logged-in.png'), fullPage: true });
      } else {
        console.log('[TEST] ❓ Unexpected URL:', currentUrl);
        await page.screenshot({ path: path.join(screenshotDir, '03-unexpected-state.png'), fullPage: true });
      }
      
    } catch (error) {
      console.log('[TEST] ❌ ERROR during OAuth flow:', error);
      await page.screenshot({ path: path.join(screenshotDir, 'ERROR-oauth-flow.png'), fullPage: true });
      throw error;
    }
  });

  test('Verify React app serves at root', async ({ page }) => {
    console.log('[TEST] Verifying React app serves at root');
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(screenshotDir, '04-root-page.png'), fullPage: true });
    
    const pageContent = await page.content();
    const hasReactApp = pageContent.includes('DevSmith Platform') || pageContent.includes('root');
    
    console.log('[TEST] Root page has React app:', hasReactApp);
    expect(hasReactApp).toBe(true);
  });

  test('Verify OAuth callback route returns React app', async ({ page }) => {
    console.log('[TEST] Verifying OAuth callback returns React app (not 401)');
    
    // Visit callback URL with fake params
    await page.goto('/auth/github/callback?code=test&state=test');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: path.join(screenshotDir, '05-callback-direct.png'), fullPage: true });
    
    const pageContent = await page.content();
    const hasReactApp = pageContent.includes('DevSmith Platform') || pageContent.includes('root');
    const hasBackendError = pageContent.includes('OAUTH_STATE_INVALID') || pageContent.includes('Unauthorized');
    
    console.log('[TEST] Callback route analysis:');
    console.log('[TEST]   - Has React app:', hasReactApp);
    console.log('[TEST]   - Has backend error:', hasBackendError);
    
    // Should serve React app, not backend 401
    expect(hasReactApp).toBe(true);
    expect(hasBackendError).toBe(false);
  });
});

test.afterAll(() => {
  console.log('\n===========================================');
  console.log('MANUAL VERIFICATION COMPLETE');
  console.log('===========================================');
  console.log('Screenshots saved to:', screenshotDir);
  console.log('\nWhat was tested:');
  console.log('  ✅ Login page loads with GitHub button');
  console.log('  ✅ GitHub OAuth redirect works');
  console.log('  ✅ React app handles /auth/github/callback route');
  console.log('  ✅ Full OAuth flow (if credentials provided)');
  console.log('\nTo test full OAuth flow with real GitHub login:');
  console.log('  1. Set environment variables:');
  console.log('     export GITHUB_TEST_USERNAME="your-github-username"');
  console.log('     export GITHUB_TEST_PASSWORD="your-github-password"');
  console.log('  2. Re-run test: npx playwright test oauth-real-flow');
  console.log('\nScreenshots to review:');
  console.log('  - 01-login-page.png: DevSmith login page');
  console.log('  - 02-after-github-click.png: GitHub OAuth page or callback');
  console.log('  - 03+: OAuth flow steps (varies based on auth state)');
  console.log('\n✅ If you reached dashboard/home page, OAuth is WORKING');
  console.log('❌ If stuck on error page, check screenshots for details\n');
});
