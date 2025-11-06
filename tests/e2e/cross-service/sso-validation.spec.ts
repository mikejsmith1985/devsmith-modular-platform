/**
 * E2E Test: SSO Validation Across All Services
 * 
 * Purpose: Ensures that authentication in Portal propagates to all other services
 * without requiring re-authentication (Single Sign-On).
 * 
 * This test validates:
 * 1. User logs in via GitHub OAuth (Portal)
 * 2. Authentication persists across all services (Review, Logs, Analytics)
 * 3. No redirects to /auth/github/login after initial authentication
 * 
 * Context: This test was added after discovering critical bug where Review service
 * was using OptionalAuthMiddleware instead of RedisSessionAuthMiddleware, causing
 * authenticated users to be redirected to login when accessing Review.
 * 
 * See: .docs/ERROR_LOG.md - "SSO Authentication Failure - Critical Bug"
 */

import { test, expect } from '@playwright/test';

test.describe('SSO Validation', () => {
  test('User logs in once and can access all services without re-authentication', async ({ page }) => {
    // GIVEN: User navigates to Portal
    await page.goto('http://localhost:3000');

    // WHEN: User initiates GitHub OAuth login
    await page.click('text=Login with GitHub');

    // THEN: Redirected to GitHub OAuth (or test login in dev mode)
    // Note: In test environment with ENABLE_TEST_AUTH=true, this will be test login
    await page.waitForURL(/.*\/auth\/github\/(callback|login|test-login).*/);

    // WHEN: OAuth completes (simulated or real)
    // Test mode: auto-redirects to dashboard
    // Real mode: user approves, GitHub redirects with code
    await page.waitForURL('**/dashboard', { timeout: 10000 });

    // THEN: User lands on Portal dashboard
    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.locator('.welcome-message, h1, h2')).toBeVisible();

    console.log('✅ Portal authentication successful');

    // ==========================================
    // Test 1: Review Service SSO
    // ==========================================
    
    // WHEN: User clicks Review card from dashboard
    await page.click('text=Review', { timeout: 5000 }).catch(async () => {
      // Fallback: navigate directly if button not found
      await page.goto('http://localhost:3000/review');
    });

    // THEN: Review app loads WITHOUT redirect to login
    await page.waitForURL('**/review**', { timeout: 10000 });
    
    // Critical assertion: URL should NOT contain auth/github/login
    const reviewURL = page.url();
    expect(reviewURL).not.toContain('auth/github/login');
    expect(reviewURL).toMatch(/.*\/review.*/);

    // Verify Review workspace loads (not just redirect)
    await expect(page.locator('h1, h2, .workspace, .review-content')).toBeVisible({ timeout: 5000 });

    console.log('✅ Review service SSO validated - no re-authentication required');

    // ==========================================
    // Test 2: Logs Service SSO
    // ==========================================
    
    // WHEN: User navigates to Logs service
    await page.goto('http://localhost:3000/logs');

    // THEN: Logs dashboard loads WITHOUT redirect to login
    await page.waitForURL('**/logs**', { timeout: 10000 });
    
    const logsURL = page.url();
    expect(logsURL).not.toContain('auth/github/login');
    expect(logsURL).toMatch(/.*\/logs.*/);

    // Verify Logs dashboard loads
    await expect(page.locator('h1, h2, .dashboard, .logs-content')).toBeVisible({ timeout: 5000 });

    console.log('✅ Logs service SSO validated - no re-authentication required');

    // ==========================================
    // Test 3: Analytics Service SSO
    // ==========================================
    
    // WHEN: User navigates to Analytics service
    await page.goto('http://localhost:3000/analytics');

    // THEN: Analytics dashboard loads WITHOUT redirect to login
    await page.waitForURL('**/analytics**', { timeout: 10000 });
    
    const analyticsURL = page.url();
    expect(analyticsURL).not.toContain('auth/github/login');
    expect(analyticsURL).toMatch(/.*\/analytics.*/);

    // Verify Analytics dashboard loads
    await expect(page.locator('h1, h2, .dashboard, .analytics-content')).toBeVisible({ timeout: 5000 });

    console.log('✅ Analytics service SSO validated - no re-authentication required');

    // ==========================================
    // Final Validation: Session Persistence
    // ==========================================
    
    // Navigate back to Portal to verify session still active
    await page.goto('http://localhost:3000/dashboard');
    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.locator('.welcome-message, h1, h2')).toBeVisible();

    console.log('✅ Session persisted across all service navigations');
  });

  test('Unauthenticated user is redirected to login from protected routes', async ({ page }) => {
    // GIVEN: User has no valid session (new browser context)

    // WHEN: User tries to access Review directly
    await page.goto('http://localhost:3000/review');

    // THEN: Redirected to GitHub OAuth login
    await page.waitForURL(/.*\/auth\/github\/login.*/);
    expect(page.url()).toContain('auth/github/login');

    console.log('✅ Protected route correctly redirects unauthenticated users');
  });

  test('Session expires after logout and requires re-authentication', async ({ page }) => {
    // GIVEN: User is authenticated
    await page.goto('http://localhost:3000');
    await page.click('text=Login with GitHub');
    await page.waitForURL('**/dashboard', { timeout: 10000 });

    // WHEN: User logs out
    await page.click('text=Logout', { timeout: 5000 }).catch(async () => {
      // Fallback: navigate to logout endpoint
      await page.goto('http://localhost:3000/auth/logout');
    });

    // THEN: Session is cleared
    await page.waitForURL(/.*\/(login|auth|$)/, { timeout: 5000 }).catch(() => {
      // Some implementations redirect to home page
    });

    // WHEN: User tries to access Review after logout
    await page.goto('http://localhost:3000/review');

    // THEN: Redirected to login (session no longer valid)
    await page.waitForURL(/.*\/auth\/github\/login.*/);
    expect(page.url()).toContain('auth/github/login');

    console.log('✅ Logout correctly invalidates session');
  });
});
