import { test, expect } from '@playwright/test';

test.describe('SMOKE: Portal Loads', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate using test endpoint
    await page.goto('http://localhost:3000');
    
    // Use test login if available
    const response = await page.request.post('http://localhost:3000/auth/test-login');
    if (response.ok()) {
      // Wait for redirect to complete
      await page.waitForURL('http://localhost:3000/', { timeout: 5000 }).catch(() => {});
      await page.waitForLoadState('networkidle').catch(() => {});
    }
  });

  test('Portal is accessible at root', async ({ page }) => {
    const response = await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Navigation renders correctly', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    const nav = page.locator('nav');
    await expect(nav).toBeVisible();
    await expect(nav).toContainText('DevSmith');
  });

  test('Dark mode toggle is visible and has Alpine.js attributes', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    // Check that dark mode button exists (button inside Alpine container)
    const darkModeButton = page.locator('button[type="button"]').filter({ 
      has: page.locator('svg') 
    }).first();
    await expect(darkModeButton).toBeVisible({ timeout: 5000 });
    
    // Verify Alpine.js attributes are present
    const alpineContainer = page.locator('[x-data*="dark"]');
    const count = await alpineContainer.count();
    expect(count).toBeGreaterThan(0);
  });
});
