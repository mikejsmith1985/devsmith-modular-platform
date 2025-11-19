/**
 * Cross-Service Session Persistence Test
 * 
 * Validates that users can:
 * 1. Log into Portal ONCE via GitHub OAuth
 * 2. Access Review without re-authentication
 * 3. Access Logs without re-authentication
 * 4. Access Analytics without re-authentication
 * 5. See consistent user menu across all services
 * 6. Logout from any service clears session everywhere
 * 
 * This is the CORE acceptance criteria for Redis session store implementation.
 */

import { test, expect } from '../fixtures/auth.fixture';
import percySnapshot from '@percy/playwright';

test.describe('Cross-Service Session Persistence', () => {
  test('should maintain session across Portal → Review → Logs → Analytics', async ({ authenticatedPage: page }) => {
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 1: Verify user is authenticated in Portal
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
      await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    // Verify user menu is visible (indicates authentication)
    const userMenu = page.locator('#user-menu, button[aria-label*="user menu"]');
    await expect(userMenu).toBeVisible();
    
    await percySnapshot(page, 'SSO Test - Portal Authenticated');
    console.log('✓ User authenticated in Portal');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 2: Navigate to Review - should NOT redirect to login
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
      await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Should NOT be on login page
    expect(page.url()).not.toContain('auth/github/login');
    expect(page.url()).toContain('/review');
    
    // Verify user menu is visible in Review
    const reviewUserMenu = page.locator('#user-menu, button[aria-label*="user menu"]');
    await expect(reviewUserMenu).toBeVisible();
    
    await percySnapshot(page, 'SSO Test - Review Without Re-Auth');
    console.log('✓ User accessed Review without re-authentication');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 3: Navigate to Logs - should NOT redirect to login
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
      await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Should NOT be on login page
    expect(page.url()).not.toContain('auth/github/login');
    expect(page.url()).toContain('/logs');
    
    // Verify user menu is visible in Logs
    const logsUserMenu = page.locator('#user-menu, button[aria-label*="user menu"]');
    await expect(logsUserMenu).toBeVisible();
    
    await percySnapshot(page, 'SSO Test - Logs Without Re-Auth');
    console.log('✓ User accessed Logs without re-authentication');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 4: Navigate to Analytics - should NOT redirect to login
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
      await page.goto('/analytics');
    await page.waitForLoadState('networkidle');
    
    // Should NOT be on login page
    expect(page.url()).not.toContain('auth/github/login');
    expect(page.url()).toContain('/analytics');
    
    // Verify user menu is visible in Analytics
    const analyticsUserMenu = page.locator('#user-menu, button[aria-label*="user menu"]');
    await expect(analyticsUserMenu).toBeVisible();
    
    await percySnapshot(page, 'SSO Test - Analytics Without Re-Auth');
    console.log('✓ User accessed Analytics without re-authentication');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 5: Verify navigation between services works seamlessly
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Analytics → Portal
    const portalNavLink = page.locator('a[href="/"], a[href="/dashboard"]');
    await portalNavLink.click();
    await page.waitForURL(/\/(dashboard)?$/);
    expect(page.url()).not.toContain('auth/github/login');
    console.log('✓ Navigation Analytics → Portal works without re-auth');
    
    // Portal → Logs
    const logsNavLink = page.locator('a[href="/logs"]');
    await logsNavLink.click();
    await page.waitForURL(/\/logs/);
    expect(page.url()).not.toContain('auth/github/login');
    console.log('✓ Navigation Portal → Logs works without re-auth');
    
    // Logs → Review
    const reviewNavLink = page.locator('a[href="/review"]');
    await reviewNavLink.click();
    await page.waitForURL(/\/review/);
    expect(page.url()).not.toContain('auth/github/login');
    console.log('✓ Navigation Logs → Review works without re-auth');
    
    await percySnapshot(page, 'SSO Test - Seamless Navigation Complete');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 6: Verify consistent user identity across services
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Get username from Portal
      await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    const portalUserMenu = page.locator('#user-menu, button[aria-label*="user menu"]');
    await portalUserMenu.click();
    await page.waitForTimeout(300);
    const portalUsername = await page.locator('[data-username], .username').textContent();
    
    // Get username from Review
      await page.goto('/review');
    await page.waitForLoadState('networkidle');
    const reviewUserMenuBtn = page.locator('#user-menu, button[aria-label*="user menu"]');
    await reviewUserMenuBtn.click();
    await page.waitForTimeout(300);
    const reviewUsername = await page.locator('[data-username], .username').textContent();
    
    // Verify usernames match
    expect(reviewUsername).toBe(portalUsername);
    console.log(`✓ User identity consistent: ${portalUsername}`);
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 7: Logout from Review and verify session cleared everywhere
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Stay on Review page and logout
    const logoutButton = page.locator('button:has-text("Logout"), a[href*="logout"]');
    await logoutButton.click();
    await page.waitForTimeout(1000);
    
    // Should redirect to login page
    await page.waitForURL(/auth\/github\/login/);
    expect(page.url()).toContain('auth/github/login');
    console.log('✓ Logout from Review redirects to login');
    
    await percySnapshot(page, 'SSO Test - After Logout');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 8: Verify session cleared - accessing other services requires re-auth
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Try to access Portal - should redirect to login
      await page.goto('/dashboard');
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('auth/github/login');
    console.log('✓ Portal requires re-authentication after logout');
    
    // Try to access Logs - should redirect to login
      await page.goto('/logs');
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('auth/github/login');
    console.log('✓ Logs requires re-authentication after logout');
    
    // Try to access Analytics - should redirect to login
      await page.goto('/analytics');
    await page.waitForTimeout(1000);
    expect(page.url()).toContain('auth/github/login');
    console.log('✓ Analytics requires re-authentication after logout');
    
    await percySnapshot(page, 'SSO Test - All Services Require Re-Auth');
    
    console.log('\n✅ ALL CROSS-SERVICE SESSION PERSISTENCE TESTS PASSED');
    console.log('✅ SSO validated: Login once, access all services');
    console.log('✅ Logout validated: Session cleared across all services');
  });
  
  test('should handle expired session gracefully', async ({ page }) => {
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST: Simulate expired session (no valid cookie)
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
      await page.goto('/dashboard');
    
    // Should redirect to login
    await page.waitForURL(/auth\/github\/login/, { timeout: 5000 });
    expect(page.url()).toContain('auth/github/login');
    
    await percySnapshot(page, 'SSO Test - Expired Session Redirect');
    console.log('✓ Expired session redirects to login');
    
    // Verify same behavior for all services
    const services = ['/review', '/logs', '/analytics'];
    for (const service of services) {
        await page.goto(service);
      await page.waitForTimeout(1000);
      expect(page.url()).toContain('auth/github/login');
      console.log(`✓ Expired session redirects to login for ${service}`);
    }
    
    console.log('\n✅ EXPIRED SESSION HANDLING VALIDATED');
  });
});
