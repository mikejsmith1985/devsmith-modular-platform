import { test, expect } from '@playwright/test';

test.describe('SMOKE: Analytics Dashboard Loads', () => {
  test('Analytics dashboard is accessible', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Dashboard renders with heading', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for main heading
    await expect(page.locator('h1')).toContainText('Analytics');
  });

  test('Chart.js is loaded', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for chart.js in page
    const html = await page.content();
    expect(html).toContain('chart.js');
  });

  test('HTMX filters are present', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for time range filter with HTMX attributes
    const timeRangeSelect = page.locator('select[name="time_range"]');
    await expect(timeRangeSelect).toBeVisible();
    
    // Should have hx-get attribute (HTMX)
    const hxGet = await timeRangeSelect.getAttribute('hx-get');
    expect(hxGet).toBeTruthy();
  });

  test('Dashboard content container exists', async ({ page }) => {
    await page.goto('http://localhost:3000/analytics', { waitUntil: 'domcontentloaded' });
    
    // Check for analytics content div that HTMX will populate
    const contentContainer = page.locator('#analytics-content');
    await expect(contentContainer).toBeVisible();
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
