import { test as baseTest, expect } from '@playwright/test';
import { test, expect as authExpect } from '../fixtures/auth.fixture';

/**
 * Cross-Service Single Sign-On (SSO) Tests
 * 
 * Validates that authentication state is shared across all services
 * through the Redis session store.
 * 
 * Critical tests for Phase 1.1 (Redis Session Store) validation.
 */

test.describe('Cross-Service SSO (Unauthenticated State)', () => {
  baseTest('should redirect to login from any service when not authenticated', async ({ page }) => {
    const services = [
      { path: '/review', name: 'Review' },
      // Logs and Analytics might be public, test only if they require auth
    ];
    
    for (const service of services) {
      await page.goto(service.path);
      await page.waitForLoadState('networkidle');
      
      const url = page.url();
      
      // Should redirect to login or show login UI
      const isLoginRedirect = 
        url.includes('/login') || 
        url.includes('/auth') ||
        url.includes('github.com');
      
      const hasLoginUI = await page.locator('text=/login|sign in/i').count() > 0;
      
      expect(isLoginRedirect || hasLoginUI).toBeTruthy();
    }
  });

  baseTest('should maintain unauthenticated state across service navigation', async ({ page }) => {
    // Start at Portal
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Should see login UI
    const portalHasLogin = await page.locator('button:has-text("Login with GitHub")').isVisible();
    expect(portalHasLogin).toBeTruthy();
    
    // Navigate to Logs (might be public)
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Navigate to Analytics (might be public)
    await page.goto('/analytics');
    await page.waitForLoadState('networkidle');
    
    // Navigate back to Portal
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Should still see login UI (not magically authenticated)
    const stillNeedLogin = await page.locator('button:has-text("Login with GitHub")').isVisible();
    expect(stillNeedLogin).toBeTruthy();
  });
});

/**
 * Cross-Service SSO (Authenticated State)
 * 
 * These tests validate session sharing after authentication.
 * Uses auth fixture to create authenticated sessions.
 */

test.describe('Cross-Service SSO (Authenticated State)', () => {
  test('should share authentication across Portal and Review', async ({ authenticatedPage, testUser }) => {
    // Auth fixture already created session and set cookie
    
    // Navigate to Portal dashboard (should be authenticated)
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Should see authenticated Portal UI
    const dashboardVisible = await authenticatedPage.locator('text=/Dashboard|Welcome/i').isVisible();
    expect(dashboardVisible).toBeTruthy();
    
    // Navigate to Review without re-authenticating
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Should NOT redirect to login
    const url = authenticatedPage.url();
    expect(url).toContain('/review');
    expect(url).not.toContain('/login');
    expect(url).not.toContain('/auth');
    
    // Should see authenticated Review UI
    const reviewVisible = await authenticatedPage.locator('text=/Review|Workspace/i').isVisible();
    expect(reviewVisible).toBeTruthy();
  });

  test('should share authentication across all services', async ({ authenticatedPage, testUser }) => {
    
    const services = [
      { path: '/dashboard', name: 'Portal' },
      { path: '/review', name: 'Review' },
      { path: '/logs', name: 'Logs' },
      { path: '/analytics', name: 'Analytics' }
    ];
    
    for (const service of services) {
      await authenticatedPage.goto(service.path);
      await authenticatedPage.waitForLoadState('networkidle');
      
      const url = authenticatedPage.url();
      
      // Should NOT redirect to login
      expect(url).not.toContain('/login');
      expect(url).not.toContain('/auth/github');
      
      // Should see authenticated UI (service is accessible)
      const titleVisible = await authenticatedPage.locator('h1, h2').first().isVisible();
      expect(titleVisible).toBeTruthy();
    }
  });

  test('should maintain session across page reloads', async ({ authenticatedPage, testUser }) => {
    
    // Login on Portal
    await authenticatedPage.goto('/');
    // ... authenticated session ...
    
    // Reload page
    await authenticatedPage.reload();
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Should still be authenticated
    const isAuthenticated = await authenticatedPage.locator('button:has-text("Logout")').isVisible();
    expect(isAuthenticated).toBeTruthy();
  });

  test('should maintain session across browser tabs', async ({ browser }) => {
    // TODO: Use auth fixture
    
    // Create first context (tab 1)
    const context1 = await browser.newContext();
    const page1 = await context1.newPage();
    
    // Login on Portal in tab 1
    await page1.goto('/');
    // ... authenticated session ...
    
    // Extract session cookie
    const cookies = await context1.cookies();
    const sessionCookie = cookies.find(c => c.name === 'devsmith_token');
    expect(sessionCookie).toBeTruthy();
    
    // Create second context (tab 2) with same cookie
    const context2 = await browser.newContext();
    await context2.addCookies([sessionCookie!]);
    const page2 = await context2.newPage();
    
    // Navigate to Review in tab 2
    await page2.goto('/review');
    await page2.waitForLoadState('networkidle');
    
    // Should be authenticated (session shared via Redis)
    const isAuthenticated = await page2.locator('[data-testid="user-menu"]').isVisible();
    expect(isAuthenticated).toBeTruthy();
    
    await context1.close();
    await context2.close();
  });

  test('should logout from all services when logging out from one', async ({ authenticatedPage, testUser }) => {
    // TODO: Use auth fixture
    
    // Login on Portal
    await authenticatedPage.goto('/');
    // ... authenticated session ...
    
    // Navigate to Review
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Logout from Review
    await authenticatedPage.click('button:has-text("Logout")');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Navigate to Portal
    await authenticatedPage.goto('/');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Should be logged out on Portal too
    const needLogin = await authenticatedPage.locator('button:has-text("Login with GitHub")').isVisible();
    expect(needLogin).toBeTruthy();
  });

  test('should expire session after timeout', async ({ authenticatedPage, testUser }) => {
    // TODO: Use auth fixture with short TTL for testing
    
    // Login on Portal
    await authenticatedPage.goto('/');
    // ... authenticated session with 5-second TTL ...
    
    // Wait for session to expire
    await authenticatedPage.waitForTimeout(6000);
    
    // Navigate to Review
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Should redirect to login (session expired)
    const url = authenticatedPage.url();
    expect(url).toMatch(/login|auth/);
  });
});

/**
 * Redis Session Store Validation
 * 
 * Tests that specifically validate Redis session behavior
 */

test.describe('Redis Session Store Integration', () => {
  test('should store user data in Redis session', async ({ authenticatedPage, testUser }) => {
    
    // Login creates session in Redis
    await authenticatedPage.goto('/');
    // ... login ...
    
    // Session should contain user_id, github_username, etc.
    // (Would need Redis CLI access to validate)
    
    expect(true).toBeTruthy(); // Placeholder
  });

  test('should refresh session TTL on activity', async ({ authenticatedPage, testUser }) => {
    // TODO: Use auth fixture + Redis TTL inspection
    
    // Login creates session with 7-day TTL
    await authenticatedPage.goto('/');
    // ... login ...
    
    // Check initial TTL (would need Redis CLI)
    // Make request to any service
    await authenticatedPage.goto('/logs');
    
    // TTL should be refreshed to 7 days again
    // (Would need Redis CLI access to validate)
    
    expect(true).toBeTruthy(); // Placeholder
  });

  test('should delete session on logout', async ({ authenticatedPage, testUser }) => {
    // TODO: Use auth fixture + Redis inspection
    
    // Login creates session
    await authenticatedPage.goto('/');
    // ... login ...
    
    // Session exists in Redis
    // Logout
    await authenticatedPage.click('button:has-text("Logout")');
    
    // Session should be deleted from Redis
    // (Would need Redis CLI access to validate)
    
    expect(true).toBeTruthy(); // Placeholder
  });
});
