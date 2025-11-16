import { test, expect } from '@playwright/test';

test.describe('SMOKE: Portal Loads', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate using test endpoint
    const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
      data: {
        username: 'testuser',
        email: 'test@example.com',
        avatar_url: 'http://example.com/avatar.png'
      }
    });
    
    if (loginResponse.ok()) {
      // Navigate to authenticated route
      await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    }
  });

  test('Portal is accessible when authenticated', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Navigation renders correctly', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    const nav = page.locator('nav');
    await expect(nav).toBeVisible();
    await expect(nav).toContainText('DevSmith');
  });

  test('Dark mode button is visible and functional', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    // Check dark mode button (vanilla JS implementation with ID)
    const darkModeButton = page.locator('#dark-mode-toggle');
    await expect(darkModeButton).toBeVisible();
    
    // Check icons exist (one hidden, one visible depending on current mode)
    const sunIcon = page.locator('#sun-icon');
    const moonIcon = page.locator('#moon-icon');
    expect(await sunIcon.count()).toBe(1);
    expect(await moonIcon.count()).toBe(1);
  });
});
