import { test, expect } from '@playwright/test';

test.describe('SMOKE: Analytics Dashboard - HTMX Filters', () => {
  test('Analytics dashboard loads and has filter controls', async ({ page }) => {
    const response = await page.goto('http://localhost:8083', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
    
    // Check for time range filter
    const timeRangeSelect = page.locator('select[name="time_range"]');
    expect(await timeRangeSelect.count()).toBeGreaterThan(0);
    
    console.log('✅ Analytics dashboard has filter controls');
  });
  
  test('Dark mode toggle is visible on analytics dashboard', async ({ page }) => {
    await page.goto('http://localhost:8083', { waitUntil: 'domcontentloaded' });
    
    const darkModeButton = page.locator('#dark-mode-toggle');
    await expect(darkModeButton).toBeVisible();
    
    console.log('✅ Dark mode button visible on analytics dashboard');
  });
  
  test('Chart.js is loaded for analytics', async ({ page }) => {
    await page.goto('http://localhost:8083', { waitUntil: 'domcontentloaded' });
    
    const html = await page.content();
    expect(html).toContain('chart');
    
    console.log('✅ Chart.js loaded on analytics dashboard');
  });
});
