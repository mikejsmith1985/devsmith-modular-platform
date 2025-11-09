import { test, expect } from '../fixtures/auth.fixture';

/**
 * Phase 0: Health App Rename Tests
 * 
 * Validates that "Logs" has been renamed to "Health" throughout the UI.
 * 
 * TDD Approach: RED → GREEN → REFACTOR
 * - RED: This test should FAIL initially (still shows "Logs")
 * - GREEN: After renaming, test should PASS
 * - REFACTOR: Clean up implementation if needed
 */

test.describe('Phase 0: Health App Rename', () => {
  
  test.beforeEach(async ({ authenticatedPage }) => {
    // Navigate to dashboard
    await authenticatedPage.goto('http://localhost:3000');
    
    // Wait for dashboard to load
    await authenticatedPage.waitForLoadState('networkidle');
  });

  test('Dashboard shows "Health" card instead of "Logs"', async ({ authenticatedPage }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Looking at app cards
    const healthCard = authenticatedPage.locator('.frosted-card').filter({ hasText: 'Health' });
    // Look for card with title "Logs" or "Development Logs" (not just text containing "logs")
    const logsCard = authenticatedPage.locator('.frosted-card h5').filter({ hasText: /^(Development )?Logs$/ });
    
    // THEN: "Health" card should exist
    await expect(healthCard).toBeVisible();
    
    // AND: "Logs" card title should NOT exist
    await expect(logsCard).not.toBeVisible();
  });

  test('Health card has correct description', async ({ authenticatedPage }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Looking at Health card
    const healthCard = authenticatedPage.locator('.frosted-card').filter({ hasText: 'Health' });
    
    // THEN: Description mentions "System Health", not "Development Logs"
    await expect(healthCard).toContainText('System Health');
    await expect(healthCard).not.toContainText('Development Logs');
  });

  test('Health card links to /health route', async ({ authenticatedPage }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Clicking Health card
    const healthCard = authenticatedPage.locator('.frosted-card').filter({ hasText: 'Health' });
    await healthCard.click();
    
    // THEN: Should navigate to /health
    await expect(authenticatedPage).toHaveURL(/\/health/);
  });

  test('Health page has correct title', async ({ authenticatedPage }) => {
    // GIVEN: User navigates to Health app
    await authenticatedPage.goto('http://localhost:3000/health');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // WHEN: Looking at page title
    const title = authenticatedPage.locator('h1, h2').first();
    
    // THEN: Title should say "Health", not "Logs"
    await expect(title).toContainText('Health');
    await expect(title).not.toContainText('Logs');
  });

  test('Navigation shows "Health" link', async ({ authenticatedPage }) => {
    // GIVEN: User is on any page
    await authenticatedPage.goto('http://localhost:3000/health');
    
    // WHEN: Looking at navigation
    const nav = authenticatedPage.locator('nav');
    
    // THEN: Navigation should show "Health" link
    await expect(nav.locator('a').filter({ hasText: 'Health' })).toBeVisible();
    await expect(nav.locator('a').filter({ hasText: 'Logs' })).not.toBeVisible();
  });
});
