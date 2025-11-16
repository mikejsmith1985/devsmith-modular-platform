import { test, expect } from '@playwright/test';

test.describe('SMOKE: Analytics Dashboard Loads', () => {
  test('Analytics dashboard is accessible', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Dashboard renders with heading', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for main heading (use more specific selector)
    await expect(page.locator('main h1')).toContainText('Analytics');
  });

  test('Chart.js is loaded', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for chart.js in page
    const html = await page.content();
    expect(html).toContain('chart');
  });

  test('HTMX filters are present', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for time range filter - look by ID since template may use IDs
    const timeRangeSelect = page.locator('select#time_range, select[name="time_range"]');
    const count = await timeRangeSelect.count();
    expect(count).toBeGreaterThan(0);
  });

  test('Dashboard content container exists', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for analytics content div that HTMX will populate
    const contentContainer = page.locator('#analytics-content');
    expect(await contentContainer.count()).toBeGreaterThan(0);
  });

  test('Alpine.js and Tailwind are loaded', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for Alpine.js
    const html = await page.content();
    expect(html).toContain('alpinejs');
    
    // Check for Tailwind
    expect(html).toContain('tailwindcss');
  });
});
