import { test, expect } from '@playwright/test';

test.describe('SMOKE: Dark Mode Toggle', () => {
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

  test('Dark mode button is clickable', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    // Find dark mode button by ID
    const darkModeButton = page.locator('#dark-mode-toggle');
    await expect(darkModeButton).toBeVisible();
    await expect(darkModeButton).toBeEnabled();
  });

  test('Clicking dark mode toggle changes DOM class', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('#dark-mode-toggle');
    
    // Get initial dark class state
    const htmlElement = page.locator('html');
    const initialClass = await htmlElement.getAttribute('class');
    
    // Click toggle
    await darkModeButton.click();
    await page.waitForTimeout(300);
    
    // Check that class changed
    const updatedClass = await htmlElement.getAttribute('class');
    
    // Either 'dark' was added or removed
    const wasDark = initialClass?.includes('dark');
    const isDark = updatedClass?.includes('dark');
    expect(wasDark).not.toBe(isDark);
  });

  test('Dark mode preference persists in localStorage', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('#dark-mode-toggle');
    
    // Click dark mode toggle
    await darkModeButton.click();
    await page.waitForTimeout(300);
    
    // Check localStorage
    const darkModeSetting = await page.evaluate(() => localStorage.getItem('darkMode'));
    expect(darkModeSetting).toBeTruthy();
  });

  test('Dark mode persists across page navigation', async ({ page }) => {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    // Enable dark mode
    const darkModeButton = page.locator('#dark-mode-toggle');
    await darkModeButton.click();
    await page.waitForTimeout(300);
    
    // Navigate to dashboard again
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    
    // Dark mode should still be active
    const htmlElement = page.locator('html');
    const classAttr = await htmlElement.getAttribute('class');
    expect(classAttr).toContain('dark');
  });
});
