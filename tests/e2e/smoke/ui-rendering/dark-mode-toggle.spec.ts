import { test, expect } from '@playwright/test';

test.describe('SMOKE: Dark Mode Toggle', () => {
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

  test('Dark mode button renders with Alpine.js attributes', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    // Check for x-data directive on parent container
    const alpineContainer = page.locator('[x-data*="dark"]').first();
    await expect(alpineContainer).toBeVisible();
  });

  test('Dark mode button is clickable', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    // Find dark mode button (should have aria-label about mode)
    const darkModeButton = page.locator('button[type="button"]').filter({ has: page.locator('svg') }).first();
    await expect(darkModeButton).toBeVisible();
    await expect(darkModeButton).toBeEnabled();
  });

  test('Clicking dark mode toggle changes DOM class', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('button[type="button"]').filter({ has: page.locator('svg') }).first();
    
    // Get initial dark class state
    const htmlElement = page.locator('html');
    const initialClass = await htmlElement.getAttribute('class');
    
    // Click toggle
    await darkModeButton.click();
    
    // Wait for class change
    await page.waitForTimeout(300);
    
    // Check that class changed
    const updatedClass = await htmlElement.getAttribute('class');
    
    // Either 'dark' was added or removed
    const wasDark = initialClass?.includes('dark');
    const isDark = updatedClass?.includes('dark');
    expect(wasDark).not.toBe(isDark);
  });

  test('Dark mode preference persists in localStorage', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('button[type="button"]').filter({ has: page.locator('svg') }).first();
    
    // Click dark mode toggle
    await darkModeButton.click();
    await page.waitForTimeout(300);
    
    // Check localStorage
    const darkModeSetting = await page.evaluate(() => localStorage.getItem('darkMode'));
    expect(darkModeSetting).toBeTruthy();
  });

  test('Dark mode persists across page navigation', async ({ page }) => {
    await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
    
    // Enable dark mode
    const darkModeButton = page.locator('button[type="button"]').filter({ has: page.locator('svg') }).first();
    await darkModeButton.click();
    await page.waitForTimeout(300);
    
    // Navigate to review
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Dark mode should still be active
    const htmlElement = page.locator('html');
    const classAttr = await htmlElement.getAttribute('class');
    expect(classAttr).toContain('dark');
  });
});
