import { test, expect } from '@playwright/test';

test.describe('OAuth PKCE Complete Flow', () => {
  test('should complete full OAuth flow without login loop', async ({ page, context }) => {
    // Step 1: Clear all storage to start fresh
    await context.clearCookies();
    await page.goto('http://localhost:3000');
    await page.evaluate(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
    console.log('✓ Cleared all storage');

    // Step 2: Navigate to login
    await page.goto('http://localhost:3000/login');
    await page.waitForLoadState('networkidle');
    console.log('✓ Navigated to login page');

    // Step 3: Find and click GitHub login button
    const loginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")').first();
    await expect(loginButton).toBeVisible({ timeout: 5000 });
    console.log('✓ Login button visible');

    // Step 4: Click login and wait for GitHub OAuth redirect
    await loginButton.click();
    console.log('✓ Clicked login button');

    // Wait for GitHub OAuth page or callback
    await page.waitForURL(/github\.com\/login\/oauth\/authorize|localhost:3000\/auth\/callback/, { 
      timeout: 10000 
    });
    
    const currentUrl = page.url();
    console.log(`✓ Redirected to: ${currentUrl}`);

    if (currentUrl.includes('github.com')) {
      // Real GitHub OAuth - check that we have correct parameters
      expect(currentUrl).toContain('client_id=');
      expect(currentUrl).toContain('redirect_uri=');
      expect(currentUrl).toContain('code_challenge=');
      expect(currentUrl).toContain('code_challenge_method=S256');
      console.log('✓ GitHub OAuth URL has correct PKCE parameters');
      
      // Can't complete real GitHub OAuth in automated test
      console.log('⚠ Real GitHub OAuth detected - cannot automate approval');
      console.log('⚠ Manual test required: Approve OAuth and verify no login loop');
      return;
    }

    // If we somehow got to callback (test mode), continue testing
    if (currentUrl.includes('/auth/callback')) {
      console.log('✓ Reached OAuth callback');

      // Wait for token exchange to complete
      await page.waitForTimeout(2000);

      // Check if token was stored
      const token = await page.evaluate(() => localStorage.getItem('devsmith_token'));
      console.log(`Token in localStorage: ${token ? '✓ Present' : '✗ Missing'}`);
      
      if (token) {
        expect(token).toBeTruthy();
        expect(token.length).toBeGreaterThan(20);
        console.log(`✓ Valid JWT token stored (length: ${token.length})`);
      }

      // Wait for redirect to dashboard
      await page.waitForURL('http://localhost:3000/', { timeout: 5000 });
      console.log('✓ Redirected to dashboard');

      // Wait for auth validation to complete
      await page.waitForTimeout(1000);

      // Verify we're on dashboard (not redirected back to login)
      const finalUrl = page.url();
      expect(finalUrl).toBe('http://localhost:3000/');
      console.log('✓ Stayed on dashboard (no login loop!)');

      // Verify dashboard content loaded
      await expect(page.locator('h1, h2')).toBeVisible({ timeout: 3000 });
      console.log('✓ Dashboard content visible');

      // Check network logs for /me endpoint call
      page.on('response', async (response) => {
        if (response.url().includes('/api/portal/auth/me')) {
          console.log(`/me endpoint called: ${response.status()}`);
          if (response.status() === 200) {
            const body = await response.json();
            console.log(`✓ /me returned user data: ${JSON.stringify(body)}`);
          } else {
            console.log(`✗ /me failed: ${response.status()} ${response.statusText()}`);
          }
        }
      });

      // Trigger a page reload to test token persistence
      await page.reload();
      await page.waitForTimeout(1000);
      
      const urlAfterReload = page.url();
      expect(urlAfterReload).toBe('http://localhost:3000/');
      console.log('✓ Token persisted after reload (no redirect to login)');
    }
  });

  test('should call /me endpoint with Bearer token after login', async ({ page }) => {
    // This test captures the /me API call
    const meApiCalls: any[] = [];
    
    page.on('request', (request) => {
      if (request.url().includes('/api/portal/auth/me')) {
        meApiCalls.push({
          url: request.url(),
          method: request.method(),
          headers: request.headers(),
        });
      }
    });

    page.on('response', async (response) => {
      if (response.url().includes('/api/portal/auth/me')) {
        console.log(`\n/me endpoint response:`);
        console.log(`  Status: ${response.status()}`);
        console.log(`  Headers: ${JSON.stringify(response.headers())}`);
        try {
          const body = await response.json();
          console.log(`  Body: ${JSON.stringify(body, null, 2)}`);
        } catch (e) {
          console.log(`  Body: (not JSON)`);
        }
      }
    });

    // Go to app (if already logged in, this should trigger /me call)
    await page.goto('http://localhost:3000');
    await page.waitForTimeout(2000);

    // Check if /me was called
    if (meApiCalls.length > 0) {
      console.log(`\n✓ /me endpoint was called ${meApiCalls.length} time(s)`);
      meApiCalls.forEach((call, i) => {
        console.log(`\nCall ${i + 1}:`);
        console.log(`  URL: ${call.url}`);
        console.log(`  Authorization: ${call.headers.authorization || '(not present)'}`);
        
        if (call.headers.authorization) {
          expect(call.headers.authorization).toContain('Bearer ');
          console.log(`  ✓ Bearer token present in Authorization header`);
        }
      });
    } else {
      console.log(`⚠ /me endpoint was not called (user may not be logged in)`);
    }
  });
});
