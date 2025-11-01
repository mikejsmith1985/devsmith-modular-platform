import { test, expect } from '@playwright/test';

test.describe('SMOKE: Portal Loads', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate using test endpoint with proper credentials
    const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
      data: {
        username: 'testuser',
        email: 'test@example.com',
        avatar_url: 'http://example.com/avatar.png'
      }
    });
    
    if (loginResponse.ok()) {
      // Token set in cookie, navigate to authenticated page
      await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
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
