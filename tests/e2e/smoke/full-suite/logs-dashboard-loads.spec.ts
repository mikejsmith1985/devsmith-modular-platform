import { test, expect } from '@playwright/test';

test.describe('SMOKE: Logs Dashboard Loads', () => {
  test('Logs dashboard is accessible', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/logs', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Dashboard renders with main controls', async ({ page }) => {
    await page.goto('http://localhost:3000/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for main heading
    await expect(page.locator('h1')).toContainText('Logs');
    
    // Check for control buttons
    await expect(page.locator('button:has-text("Pause")')).toBeVisible();
    await expect(page.locator('button:has-text("Clear")')).toBeVisible();
  });

  test('Log cards render with Tailwind styling', async ({ page }) => {
    await page.goto('http://localhost:3000/logs', { waitUntil: 'domcontentloaded' });
    
    // Wait for log output container
    const logsContainer = page.locator('#logs-output');
    await expect(logsContainer).toBeVisible();
    
    // Check for Tailwind classes indicating cards
    const html = await page.content();
    expect(html).toContain('rounded-lg');
    expect(html).toContain('shadow-sm');
  });

  test('Filter controls are present', async ({ page }) => {
    await page.goto('http://localhost:3000/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for level filter
    await expect(page.locator('select[name="level"]')).toBeVisible();
    
    // Check for service filter
    await expect(page.locator('select[name="service"]')).toBeVisible();
    
    // Check for search input
    await expect(page.locator('input[type="search"]')).toBeVisible();
  });

  test('WebSocket connection status indicator is present', async ({ page }) => {
    await page.goto('http://localhost:3000/logs', { waitUntil: 'domcontentloaded' });
    
    // Check for connection status div
    const statusIndicator = page.locator('[class*="connection"]');
    expect(await statusIndicator.count()).toBeGreaterThan(0);
  });
});
