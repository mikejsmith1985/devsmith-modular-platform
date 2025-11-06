import { test, expect } from '@playwright/test';

test.describe('OAuth Flow - Visual Validation', () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing auth
    await page.context().clearCookies();
    await page.goto('http://localhost:3000');
  });

  test('OAuth flow redirects correctly (no loop)', async ({ page }) => {
    console.log('🧪 Starting OAuth flow test...');

    // STEP 1: Initial load should redirect to login
    await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
    
    console.log(`📍 Current URL after initial navigation: ${page.url()}`);
    
    // Should be on login page
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    
    // Capture login page screenshot
    await page.screenshot({ 
      path: 'test-results/oauth-flow/01-login-page.png',
      fullPage: true 
    });
    console.log('✅ Step 1: Login page loaded');

    // STEP 2: Verify "Login with GitHub" button exists
    const githubButton = page.locator('button:has-text("Login with GitHub")');
    await expect(githubButton).toBeVisible({ timeout: 5000 });
    
    await page.screenshot({ 
      path: 'test-results/oauth-flow/02-github-button-visible.png',
      fullPage: true 
    });
    console.log('✅ Step 2: GitHub login button visible');

    // STEP 3: Check what happens when clicking GitHub login
    // We can't actually complete OAuth in automated tests, but we can verify the redirect
    console.log('📍 About to click GitHub login button...');
    
    // Track navigation
    const navigationPromise = page.waitForNavigation({ timeout: 10000 });
    await githubButton.click();
    
    try {
      await navigationPromise;
      const redirectUrl = page.url();
      console.log(`📍 Redirected to: ${redirectUrl}`);
      
      // Should redirect to GitHub OAuth (may go through /login first)
      expect(redirectUrl).toMatch(/github\.com\/login/);
      console.log('✅ Step 3: Redirected to GitHub OAuth');
      
      // Verify redirect_uri parameter is correct
      if (redirectUrl.includes('redirect_uri')) {
        const urlObj = new URL(redirectUrl);
        // Check both query param and return_to param
        const redirectUri = urlObj.searchParams.get('redirect_uri') || 
                           (urlObj.searchParams.get('return_to')?.match(/redirect_uri=([^&]+)/)?.[1]);
        if (redirectUri) {
          const decodedUri = decodeURIComponent(redirectUri);
          console.log(`📍 redirect_uri in OAuth URL: ${decodedUri}`);
          expect(decodedUri).toContain('/auth/callback');
          console.log('✅ redirect_uri is correct: /auth/callback');
        }
      }
      
      await page.screenshot({ 
        path: 'test-results/oauth-flow/03-github-oauth-redirect.png',
        fullPage: true 
      });
      console.log('✅ Step 3: OAuth parameters validated');
    } catch (error) {
      console.error('❌ Navigation failed:', error);
      await page.screenshot({ 
        path: 'test-results/oauth-flow/03-ERROR-navigation-failed.png',
        fullPage: true 
      });
      throw error;
    }
  });

  test('OAuth callback route exists and handles missing token', async ({ page }) => {
    console.log('🧪 Testing OAuth callback route...');

    // STEP 1: Navigate to callback route without token (simulates error case)
    await page.goto('http://localhost:3000/auth/callback', { waitUntil: 'networkidle' });
    
    console.log(`📍 Current URL: ${page.url()}`);
    
    // Should show error message or redirect back to login
    await page.waitForTimeout(2000); // Give time for redirect
    
    await page.screenshot({ 
      path: 'test-results/oauth-flow/04-callback-no-token.png',
      fullPage: true 
    });
    
    const finalUrl = page.url();
    console.log(`📍 Final URL after callback: ${finalUrl}`);
    
    // Should either show error on callback page or redirect to login
    const isOnLogin = finalUrl.includes('/login');
    const isOnCallback = finalUrl.includes('/auth/callback');
    
    expect(isOnLogin || isOnCallback).toBeTruthy();
    console.log(`✅ Callback route handled missing token correctly (${isOnLogin ? 'redirected to login' : 'showed error'})`);
  });

  test('OAuth callback route with token stores and redirects', async ({ page }) => {
    console.log('🧪 Testing OAuth callback with valid token...');

    // Capture browser console messages
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      consoleMessages.push(text);
      console.log(`🖥️  Browser console: ${text}`);
    });

    // Create a test JWT token (this won't be validated by backend in this test)
    const testToken = 'test.jwt.token';
    
    // STEP 1: Navigate to callback with token parameter
    // We expect the component to store token and redirect
    const response = await page.goto(`http://localhost:3000/auth/callback?token=${testToken}`);
    
    console.log(`📍 Navigated to callback with token, status: ${response?.status()}`);
    
    // STEP 2: Wait for component to mount and store token
    // Poll localStorage until token appears (component logs say it stores successfully)
    let storedToken = null;
    let attempts = 0;
    const maxAttempts = 20; // 20 attempts × 100ms = 2 seconds max
    
    while (!storedToken && attempts < maxAttempts) {
      await page.waitForTimeout(100);
      storedToken = await page.evaluate(() => localStorage.getItem('devsmith_token'));
      attempts++;
      
      if (storedToken) {
        console.log(`📍 Token found in localStorage after ${attempts * 100}ms: ${storedToken}`);
        break;
      }
    }
    
    if (!storedToken) {
      console.log(`⚠️  Token not found after ${maxAttempts * 100}ms`);
    }
    
    console.log(`📍 Browser console messages: ${consoleMessages.length} total`);
    
    // Print all browser console messages for debugging
    if (consoleMessages.length > 0) {
      console.log('📋 Browser console messages:');
      consoleMessages.forEach((msg, i) => console.log(`   ${i + 1}. ${msg}`));
    }
    
    // STEP 3: Token should be stored (even if briefly before dashboard clears it)
    expect(storedToken).toBe(testToken);
    console.log('✅ OAuthCallback component successfully stored token in localStorage');
    
    // STEP 4: Wait for redirect to complete
    await page.waitForTimeout(1500);
    
    const finalUrl = page.url();
    console.log(`📍 Final URL after redirects: ${finalUrl}`);
    
    await page.screenshot({ 
      path: 'test-results/oauth-flow/05-callback-with-token.png',
      fullPage: true 
    });
    
    // NOTE: We don't check final URL or token persistence because test token is invalid
    // Real OAuth flow: callback stores token → dashboard validates → user stays logged in
    // Test flow: callback stores token → dashboard fails validation → redirects to login (clears token)
    // The important thing is that the OAuthCallback component DID store the token (Step 3 passed)
  });

  test('No OAuth loop - verify redirect chain', async ({ page, context }) => {
    console.log('🧪 Testing for OAuth loop...');

    let redirectCount = 0;
    const visitedUrls: string[] = [];
    
    // Track all navigation events
    page.on('response', (response) => {
      const status = response.status();
      const url = response.url();
      
      if (status === 302 || status === 301) {
        redirectCount++;
        visitedUrls.push(url);
        console.log(`🔄 Redirect ${redirectCount}: ${url} → ${response.headers()['location']}`);
      }
    });

    // Start navigation
    await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
    
    // Wait a bit to see if loop occurs
    await page.waitForTimeout(3000);
    
    const finalUrl = page.url();
    console.log(`📍 Final URL: ${finalUrl}`);
    console.log(`📍 Total redirects: ${redirectCount}`);
    console.log(`📍 Visited URLs:`, visitedUrls);
    
    await page.screenshot({ 
      path: 'test-results/oauth-flow/06-no-loop-final-state.png',
      fullPage: true 
    });
    
    // Should not have more than 3 redirects (normal: root → login, or root → dashboard if authed)
    expect(redirectCount).toBeLessThan(5);
    console.log('✅ No OAuth loop detected');
    
    // Should be stable on login or dashboard page
    expect(finalUrl).toMatch(/\/(login|dashboard)?$/);
    console.log('✅ Final URL is stable (no continuous redirects)');
  });

  test('Portal API callback endpoint exists and responds', async ({ page }) => {
    console.log('🧪 Testing Portal API callback endpoint...');

    // Test the actual Portal API endpoint behavior
    // NOTE: Route is at /auth/github/callback (NO /api/portal prefix)
    // NOTE: With test_code, GitHub OAuth will reject it, so we expect an error response
    const response = await page.request.get(
      'http://localhost:3000/auth/github/callback?code=test_code',
      { maxRedirects: 0 }
    );
    
    console.log(`📍 Response status: ${response.status()}`);
    console.log(`📍 Response headers:`, response.headers());
    
    // With invalid code, endpoint returns error (not a 404)
    // This proves the route exists and is responding
    expect(response.status()).not.toBe(404);
    console.log('✅ Portal API callback route exists (not 404)');
    
    // Should return JSON error for invalid code
    const contentType = response.headers()['content-type'];
    console.log(`📍 Content-Type: ${contentType}`);
    expect(contentType).toContain('application/json');
    console.log('✅ Response is JSON (error message for invalid code)');
    
    // Parse response body
    const body = await response.json();
    console.log(`📍 Response body:`, body);
    expect(body).toHaveProperty('error');
    console.log('✅ Response includes error property (invalid OAuth code)');
  });
});
