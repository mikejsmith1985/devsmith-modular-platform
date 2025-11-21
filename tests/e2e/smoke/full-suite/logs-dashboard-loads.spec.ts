import { test, expect } from '@playwright/test';

test.describe('SMOKE: Logs Dashboard Loads', () => {
  test('Logs dashboard is accessible', async ({ page }) => {
    const response = await page.goto('/logs', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Dashboard renders with main controls', async ({ page }) => {
    await page.goto('/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for main heading (use more specific selector to avoid strict mode)
    await expect(page.locator('.logs-header h1')).toContainText('Logs');
    
    // Check for control buttons (look by ID or text content)
    await expect(page.locator('#pause-btn')).toBeVisible();
    await expect(page.locator('#clear-btn')).toBeVisible();
  });

  test('Log cards render with Tailwind styling', async ({ page }) => {
    await page.goto('/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for log output container exists
    const logsContainer = page.locator('#logs-output');
    expect(await logsContainer.count()).toBeGreaterThan(0);
    
    // Check for styling elements in the page
    const html = await page.content();
    expect(html).toContain('logs-container');  // Main container class
    expect(html).toContain('logs-header');     // Header styling class
    expect(html).toContain('logs-output');     // Output area class
    expect(html).toContain('btn-control');     // Button styling
  });

  test('Filter controls are present', async ({ page }) => {
    await page.goto('/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for level filter
    await expect(page.locator('select#level-filter')).toBeVisible();
    
    // Check for service filter
    await expect(page.locator('select#service-filter')).toBeVisible();
    
    // Check for search input
    await expect(page.locator('input#search-input')).toBeVisible();
  });

  test('WebSocket connection status indicator is present', async ({ page }) => {
    await page.goto('/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for connection status span
    const statusIndicator = page.locator('#connection-status');
    await expect(statusIndicator).toBeVisible();
    await expect(statusIndicator).toHaveClass(/status-indicator/);
  });
});
