import { test, expect } from '@playwright/test';

test.describe('SMOKE: Logs Dashboard - HTMX Filters', () => {
  test('Logs dashboard loads and has filter controls', async ({ page }) => {
    const response = await page.goto('http://localhost:8082', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
    
    // Check for filter controls - use ID selectors from actual template
    const levelFilter = page.locator('select#level-filter');
    const serviceFilter = page.locator('select#service-filter');
    const searchInput = page.locator('input#search-input');
    
    expect(await levelFilter.count()).toBeGreaterThan(0);
    expect(await serviceFilter.count()).toBeGreaterThan(0);
    expect(await searchInput.count()).toBeGreaterThan(0);
    
    console.log('✅ Logs dashboard has filter controls');
  });
  
  test('Dark mode toggle is visible on logs dashboard', async ({ page }) => {
    await page.goto('http://localhost:8082', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('#dark-mode-toggle');
    await expect(darkModeButton).toBeVisible();
    
    console.log('✅ Dark mode button visible on logs dashboard');
  });
  
  test('Logs output container is present', async ({ page }) => {
    await page.goto('http://localhost:8082', { waitUntil: 'domcontentloaded' });
    
    const logsOutput = page.locator('#logs-output, .logs-output, [class*="logs"]');
    expect(await logsOutput.count()).toBeGreaterThan(0);
    
    console.log('✅ Logs output container present');
  });
});
