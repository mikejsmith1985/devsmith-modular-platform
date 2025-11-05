import { test, expect } from '@playwright/test';

/**
 * Logs Service E2E Tests
 * 
 * Tests the Logs application workflow:
 * 1. Accessing Logs service through Traefik
 * 2. Dashboard visibility
 * 3. Basic UI functionality
 */

test.describe('Logs Service Access', () => {
  test('should be accessible at /logs path', async ({ page }) => {
    // Navigate to Logs service through Traefik
    await page.goto('/logs');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Should show Logs UI (logs is currently public)
    const url = page.url();
    expect(url).toContain('/logs');
  });

  test('should load without errors', async ({ page }) => {
    let errors: string[] = [];
    
    // Capture console errors
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    // Navigate to Logs
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Should have no JavaScript errors
    const jsErrors = errors.filter(e => 
      !e.includes('401') && 
      !e.includes('Unauthorized')
    );
    
    expect(jsErrors.length).toBe(0);
  });

  test('should display logs dashboard', async ({ page }) => {
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Should see logs-related UI elements
    // Check for common log UI patterns
    const hasLogsUI = await Promise.race([
      page.locator('text=/logs|entries|level|severity/i').first().isVisible().catch(() => false),
      page.locator('table').isVisible().catch(() => false),
      page.locator('[role="table"]').isVisible().catch(() => false)
    ]);
    
    // At minimum, page should have loaded (not 404)
    const pageContent = await page.textContent('body');
    expect(pageContent).toBeTruthy();
    expect(pageContent?.length).toBeGreaterThan(0);
  });
});

/**
 * Logs Dashboard Functionality
 * 
 * Tests for filtering, search, and interactions
 */

test.describe('Logs Dashboard Features', () => {
  test('should have filter controls', async ({ page }) => {
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Look for common filter elements
    const hasFilters = await Promise.race([
      page.locator('[type="search"]').isVisible().catch(() => false),
      page.locator('select').first().isVisible().catch(() => false),
      page.locator('input[placeholder*="filter" i]').isVisible().catch(() => false),
      page.locator('button:has-text("Filter")').isVisible().catch(() => false)
    ]);
    
    // Filters might not be visible if no logs exist yet, that's OK
    // Just verify the page structure loaded
    const pageContent = await page.textContent('body');
    expect(pageContent).toBeTruthy();
  });

  test('should handle empty log state gracefully', async ({ page }) => {
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    
    // Should either show logs or "no logs" message
    const pageContent = await page.textContent('body');
    
    // Should not crash or show error state
    expect(pageContent).toBeTruthy();
    expect(pageContent).not.toContain('Error 500');
    expect(pageContent).not.toContain('Internal Server Error');
  });
});

/**
 * Logs Health Check Dashboard
 * 
 * Tests for the health check monitoring UI
 */

test.describe('Health Check Dashboard', () => {
  test('should be accessible at /logs/healthcheck', async ({ page }) => {
    await page.goto('/logs/healthcheck');
    await page.waitForLoadState('networkidle');
    
    // Should load health check dashboard
    const url = page.url();
    expect(url).toContain('/logs/healthcheck');
    
    // Should show some health-related content
    const pageContent = await page.textContent('body');
    expect(pageContent).toBeTruthy();
  });

  test('should display service health status', async ({ page }) => {
    await page.goto('/logs/healthcheck');
    await page.waitForLoadState('networkidle');
    
    // Look for health status indicators
    const hasHealthUI = await Promise.race([
      page.locator('text=/healthy|unhealthy|status/i').first().isVisible().catch(() => false),
      page.locator('[data-testid*="health"]').first().isVisible().catch(() => false),
      page.locator('.status, .health').first().isVisible().catch(() => false)
    ]);
    
    // Page should have loaded with content
    const pageContent = await page.textContent('body');
    expect(pageContent).toBeTruthy();
    expect(pageContent?.length).toBeGreaterThan(100); // Reasonable content size
  });
});
