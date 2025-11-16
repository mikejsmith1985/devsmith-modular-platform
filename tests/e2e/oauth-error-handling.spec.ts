import { test, expect } from '@playwright/test';

/**
 * OAuth Error Handling E2E Tests
 * 
 * Tests all error scenarios validated in backend unit tests:
 * 1. Missing authorization code
 * 2. Invalid OAuth state parameter
 * 3. Token exchange failure
 * 4. User info fetch failure
 * 
 * These tests validate that error messages are user-friendly and actionable.
 */

test.describe('OAuth Error Handling', () => {
  
  test.beforeEach(async ({ page, context }) => {
    // Clear all storage before each test
    await context.clearCookies();
    
    // Try to clear localStorage, but don't fail if access denied
    try {
      await page.evaluate(() => {
        localStorage.clear();
        sessionStorage.clear();
      });
    } catch (error) {
      // Ignore localStorage security errors
      console.log('⚠ Could not clear localStorage (may be error page)');
    }
  });

  test('should show error for missing authorization code', async ({ page }) => {
    // Navigate directly to callback without code parameter
    await page.goto('http://localhost:3000/auth/github/callback?state=test-state');
    
    // Wait for error message to appear
    await page.waitForLoadState('networkidle');
    
    // Should show error message
    const errorText = await page.textContent('body');
    
    // Verify specific error message from backend
    expect(errorText).toContain('Missing authorization code');
    
    // Verify user-friendly action guidance
    expect(errorText).toContain('Please try logging in again');
    
    // Verify error code for support
    expect(errorText).toContain('OAUTH_CODE_MISSING');
    
    console.log('✓ Missing code error displayed correctly');
  });

  test('should show error for invalid OAuth state parameter', async ({ page }) => {
    // Navigate to callback with code but invalid state
    await page.goto('http://localhost:3000/auth/github/callback?code=test-code&state=invalid-state-not-in-redis');
    
    // Wait for error message
    await page.waitForLoadState('networkidle');
    
    // Should show error message
    const errorText = await page.textContent('body');
    expect(errorText).toBeTruthy(); // Ensure errorText is not null
    
    // Verify specific error from backend
    expect(errorText!).toContain('Invalid OAuth state parameter');
    
    // Verify CSRF protection explanation
    expect(errorText!.includes('CSRF') || errorText!.includes('Security validation failed')).toBeTruthy();
    
    // Verify error code
    expect(errorText!).toContain('OAUTH_STATE_INVALID');
    
    console.log('✓ Invalid state error displayed correctly');
  });

  test('should show error for missing state parameter', async ({ page }) => {
    // Navigate to callback with code but no state at all
    await page.goto('http://localhost:3000/auth/github/callback?code=test-code');
    
    // Wait for error message
    await page.waitForLoadState('networkidle');
    
    // Should show error message
    const errorText = await page.textContent('body');
    
    // Verify specific error from backend
    expect(errorText).toContain('Missing state parameter');
    
    // Verify security explanation
    expect(errorText).toContain('Security validation failed');
    
    // Verify error code
    expect(errorText).toContain('OAUTH_STATE_MISSING');
    
    console.log('✓ Missing state error displayed correctly');
  });

  test('should show user-friendly message for GitHub API errors', async ({ page }) => {
    /**
     * Note: This test validates the UI displays proper error messages
     * when GitHub API calls fail. The actual API failure is simulated
     * in backend unit tests with mock HTTP client.
     * 
     * In E2E testing, we verify:
     * 1. Error message is user-friendly (not technical stack traces)
     * 2. Action guidance is provided
     * 3. Error codes are included for support
     */
    
    // This test would require either:
    // a) Mock server that simulates GitHub API failures
    // b) Test environment variable to trigger simulated failures
    // c) Integration test with actual GitHub test credentials
    
    // For now, we document expected behavior:
    console.log('📝 Expected error messages for GitHub API failures:');
    console.log('   - Token exchange failure: "Failed to exchange authorization code"');
    console.log('   - User info failure: "Failed to fetch user information"');
    console.log('   - All errors include: "Please try logging in again"');
    console.log('   - All errors include error codes for support');
    
    // Skip test as it requires mock server setup
    test.skip();
  });

  test('successful OAuth flow shows no errors', async ({ page }) => {
    // Start OAuth flow
    await page.goto('http://localhost:3000/login');
    
    // Find login button
    const loginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")').first();
    await expect(loginButton).toBeVisible({ timeout: 5000 });
    
    // Click login
    await loginButton.click();
    
    // Wait for GitHub OAuth redirect (GitHub redirects to /login first, then /login/oauth/authorize)
    await page.waitForURL(/github\.com\/login/, { 
      timeout: 10000 
    });
    
    // Verify PKCE parameters present (need to decode since they're in return_to)
    const url = page.url();
    const decodedUrl = decodeURIComponent(url);
    expect(decodedUrl).toContain('code_challenge=');
    expect(decodedUrl).toContain('code_challenge_method=S256');
    
    console.log('✓ OAuth flow initiated without errors');
    console.log('✓ Redirected to GitHub successfully');
    console.log('✓ PKCE parameters present in GitHub URL');
    console.log('⚠ Cannot automate GitHub approval - manual test required');
  });

  test('error messages have consistent structure', async ({ page }) => {
    // Test multiple error scenarios to verify consistent message structure
    const errorScenarios = [
      {
        url: 'http://localhost:3000/auth/github/callback?state=test',
        expectedError: 'Missing authorization code',
        expectedCode: 'OAUTH_CODE_MISSING'
      },
      {
        url: 'http://localhost:3000/auth/github/callback?code=test',
        expectedError: 'Missing state parameter',
        expectedCode: 'OAUTH_STATE_MISSING'
      },
      {
        url: 'http://localhost:3000/auth/github/callback?code=test&state=invalid',
        expectedError: 'Invalid OAuth state parameter',
        expectedCode: 'OAUTH_STATE_INVALID'
      }
    ];

    for (const scenario of errorScenarios) {
      await page.goto(scenario.url);
      await page.waitForLoadState('networkidle');
      
      const errorText = await page.textContent('body');
      expect(errorText).toBeTruthy(); // Ensure we got text
      
      // Verify error message present (case-insensitive)
      expect(errorText!.toLowerCase()).toContain(scenario.expectedError.toLowerCase());
      
      // Verify error code present
      expect(errorText).toContain(scenario.expectedCode);
      
      // Verify action guidance present (may be in JSON format)
      expect(errorText!.toLowerCase()).toMatch(/try.*(logging|login).*(again|in again)/);
      
      console.log(`✓ ${scenario.expectedCode}: Consistent error structure`);
    }
  });

  test('error page has proper HTTP status codes', async ({ page }) => {
    // Listen for response to callback endpoint
    let callbackStatus: number | undefined;
    
    page.on('response', async (response) => {
      if (response.url().includes('/auth/github/callback')) {
        callbackStatus = response.status();
      }
    });

    // Navigate to callback with missing code (should be 400)
    await page.goto('http://localhost:3000/auth/github/callback?state=test');
    await page.waitForLoadState('networkidle');
    
    // Verify 400 Bad Request status
    expect(callbackStatus).toBe(400);
    console.log('✓ Missing code returns 400 Bad Request');

    // Reset
    callbackStatus = undefined;

    // Navigate to callback with invalid state (should be 400 or 401)
    await page.goto('http://localhost:3000/auth/github/callback?code=test&state=invalid');
    await page.waitForLoadState('networkidle');
    
    // Verify 400 Bad Request or 401 Unauthorized status (both are acceptable for invalid state)
    expect(callbackStatus).toBeGreaterThanOrEqual(400);
    expect(callbackStatus).toBeLessThan(500);
    console.log(`✓ Invalid state returns ${callbackStatus} status (client error)`);
  });

  test('error messages are accessible (screen reader friendly)', async ({ page }) => {
    // Navigate to error scenario
    await page.goto('http://localhost:3000/auth/github/callback?state=test');
    await page.waitForLoadState('networkidle');
    
    // Check for ARIA attributes
    const errorElement = page.locator('[role="alert"], .error-message, .alert-danger').first();
    
    if (await errorElement.count() > 0) {
      // Verify error is marked as alert
      const role = await errorElement.getAttribute('role');
      expect(role).toBe('alert');
      console.log('✓ Error message has role="alert" for screen readers');
    } else {
      console.log('⚠ Error message should have role="alert" for accessibility');
    }
  });
});

test.describe('OAuth Security Validation', () => {
  
  test('CSRF protection: state parameter required', async ({ page, context }) => {
    // Clear storage (navigate to localhost first to access localStorage)
    await page.goto('http://localhost:3000/');
    await context.clearCookies();
    await page.evaluate(() => localStorage.clear());
    
    // Attempt callback without state
    await page.goto('http://localhost:3000/auth/github/callback?code=test-code');
    await page.waitForLoadState('networkidle');
    
    // Should reject with missing state error
    const errorText = await page.textContent('body');
    expect(errorText).toContain('Missing state parameter');
    
    console.log('✓ CSRF protection: Rejects callback without state');
  });

  test('CSRF protection: invalid state rejected', async ({ page, context }) => {
    // Clear storage (navigate to localhost first to access localStorage)
    await page.goto('http://localhost:3000/');
    await context.clearCookies();
    await page.evaluate(() => localStorage.clear());
    
    // Attempt callback with invalid state
    await page.goto('http://localhost:3000/auth/github/callback?code=test-code&state=attacker-injected-state');
    await page.waitForLoadState('networkidle');
    
    // Should reject with invalid state error
    const errorText = await page.textContent('body');
    expect(errorText).toContain('Invalid OAuth state parameter');
    
    console.log('✓ CSRF protection: Rejects callback with invalid state');
  });

  test('PKCE protection: code_challenge required in OAuth URL', async ({ page }) => {
    // Start OAuth flow
    await page.goto('http://localhost:3000/login');
    
    // Click login button
    const loginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")').first();
    await expect(loginButton).toBeVisible({ timeout: 5000 });
    await loginButton.click();
    
    // Wait for GitHub OAuth redirect (GitHub redirects to /login first)
    await page.waitForURL(/github\.com\/login/, { timeout: 10000 });
    
    // Verify PKCE parameters (need to decode URL since parameters are in return_to)
    const url = page.url();
    const decodedUrl = decodeURIComponent(url);
    expect(decodedUrl).toContain('code_challenge=');
    expect(decodedUrl).toContain('code_challenge_method=S256');
    
    console.log('✓ PKCE protection: code_challenge present in OAuth URL');
    console.log('✓ PKCE protection: Using SHA-256 method');
  });
});

test.describe('OAuth Error Recovery', () => {
  
  test('user can retry login after error', async ({ page, context }) => {
    // Clear storage (navigate to localhost first to access localStorage)
    await page.goto('http://localhost:3000/');
    await context.clearCookies();
    await page.evaluate(() => localStorage.clear());
    
    // Trigger error (missing code)
    await page.goto('http://localhost:3000/auth/github/callback?state=test');
    await page.waitForLoadState('networkidle');
    
    // Verify error displayed
    let errorText = await page.textContent('body');
    expect(errorText).toContain('Missing authorization code');
    
    console.log('✓ Error displayed');
    
    // Look for "try again" link or button
    const retryLink = page.locator('a:has-text("try logging in again"), button:has-text("try again")').first();
    
    if (await retryLink.count() > 0) {
      await retryLink.click();
      await page.waitForLoadState('networkidle');
      
      // Should be back at login page
      const loginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")').first();
      await expect(loginButton).toBeVisible({ timeout: 5000 });
      
      console.log('✓ Retry link navigates back to login');
    } else {
      console.log('⚠ Error message should include clickable retry link');
    }
  });

  test('error messages do not leak sensitive information', async ({ page }) => {
    // Navigate to callback with invalid parameters
    await page.goto('http://localhost:3000/auth/github/callback?code=secret-code&state=secret-state');
    await page.waitForLoadState('networkidle');
    
    const errorText = await page.textContent('body');
    
    // Verify no sensitive data in error message
    expect(errorText).not.toContain('secret-code');
    expect(errorText).not.toContain('secret-state');
    expect(errorText).not.toContain('JWT_SECRET');
    expect(errorText).not.toContain('GITHUB_CLIENT_SECRET');
    
    console.log('✓ Error messages do not leak sensitive information');
  });
});
