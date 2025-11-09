import { test, expect } from '@playwright/test';

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
  
  test.beforeEach(async ({ page }) => {
    // Navigate to dashboard
    await page.goto('http://localhost:3000');
    
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');
  });

  test('Dashboard shows "Health" card instead of "Logs"', async ({ page }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Looking at app cards
    const healthCard = page.locator('.card').filter({ hasText: 'Health' });
    const logsCard = page.locator('.card').filter({ hasText: 'Logs' });
    
    // THEN: "Health" card should exist
    await expect(healthCard).toBeVisible();
    
    // AND: "Logs" card should NOT exist
    await expect(logsCard).not.toBeVisible();
  });

  test('Health card has correct description', async ({ page }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Looking at Health card
    const healthCard = page.locator('.card').filter({ hasText: 'Health' });
    
    // THEN: Description mentions "System Health", not "Development Logs"
    await expect(healthCard).toContainText('System Health');
    await expect(healthCard).not.toContainText('Development Logs');
  });

  test('Health card links to /health route', async ({ page }) => {
    // GIVEN: User is on dashboard
    
    // WHEN: Clicking Health card
    const healthCard = page.locator('.card').filter({ hasText: 'Health' });
    await healthCard.click();
    
    // THEN: Should navigate to /health
    await expect(page).toHaveURL(/\/health/);
  });

  test('Health page has correct title', async ({ page }) => {
    // GIVEN: User navigates to Health app
    await page.goto('http://localhost:3000/health');
    await page.waitForLoadState('networkidle');
    
    // WHEN: Looking at page title
    const title = page.locator('h1, h2').first();
    
    // THEN: Title should say "Health", not "Logs"
    await expect(title).toContainText('Health');
    await expect(title).not.toContainText('Logs');
  });

  test('Navigation shows "Health" link', async ({ page }) => {
    // GIVEN: User is on any page
    await page.goto('http://localhost:3000/health');
    
    // WHEN: Looking at navigation
    const nav = page.locator('nav');
    
    // THEN: Navigation should show "Health" link
    await expect(nav.locator('a').filter({ hasText: 'Health' })).toBeVisible();
    await expect(nav.locator('a').filter({ hasText: 'Logs' })).not.toBeVisible();
  });
});
