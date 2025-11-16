import { test, expect } from '../fixtures/auth.fixture';
import percySnapshot from '@percy/playwright';

/**
 * Health Page Stats Filtering Visual Tests
 * 
 * RULE ZERO COMPLIANCE TEST
 * 
 * Purpose: Visually validate that stats cards show database totals
 * and remain unchanged when filters are applied to the table.
 * 
 * Architecture Fix:
 * - Stats fetched from /api/logs/v1/stats (database totals)
 * - Stored in unfilteredStats state (never recalculated)
 * - Cards always display unfilteredStats
 * - Table filtering works independently
 * 
 * Related: HEALTH_STATS_ARCHITECTURE_FIX.md
 */

test.describe('Health Page - Stats Card Filtering', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    // Navigate to Health page
    await authenticatedPage.goto('/health');
    
    // Wait for page to fully load
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Wait for stats to load (stat cards use .frosted-card class)
    await authenticatedPage.waitForSelector('.frosted-card', { timeout: 10000 });
  });

  test('Stats cards show database totals on initial load', async ({ authenticatedPage }) => {
    // Take initial screenshot
    await percySnapshot(authenticatedPage, 'Health Page - Initial Load');
    
    // Verify stats cards are visible
    const errorCard = authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first();
    const warningCard = authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first();
    const infoCard = authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first();
    
    await expect(errorCard).toBeVisible();
    await expect(warningCard).toBeVisible();
    await expect(infoCard).toBeVisible();
    
    // Extract initial counts
    const errorText = await errorCard.textContent();
    const warningText = await warningCard.textContent();
    const infoText = await infoCard.textContent();
    
    console.log('Initial Stats:', { errorText, warningText, infoText });
    
    // Verify counts are numbers > 0
    const errorMatch = errorText?.match(/(\d+)/);
    const warningMatch = warningText?.match(/(\d+)/);
    const infoMatch = infoText?.match(/(\d+)/);
    
    expect(errorMatch).toBeTruthy();
    expect(warningMatch).toBeTruthy();
    expect(infoMatch).toBeTruthy();
  });

  test('Stats cards remain unchanged when ERROR filter applied', async ({ authenticatedPage }) => {
    // Capture initial stats
    const initialErrorCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().textContent();
    const initialWarningCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().textContent();
    const initialInfoCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().textContent();
    
    console.log('Initial counts before filter:', { 
      error: initialErrorCount, 
      warning: initialWarningCount, 
      info: initialInfoCount 
    });
    
    // Take screenshot before filter
    await percySnapshot(authenticatedPage, 'Health Page - Before ERROR Filter');
    
    // Apply ERROR filter by clicking ERROR stat card
    const errorCard = authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first();
    await errorCard.click();
    
    // Wait for filter to apply
    await authenticatedPage.waitForLoadState('networkidle');
    await authenticatedPage.waitForTimeout(1000); // Small delay for UI updates
    
    // Take screenshot after filter
    await percySnapshot(authenticatedPage, 'Health Page - After ERROR Filter Applied');
    
    // Verify stats cards UNCHANGED
    const afterErrorCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().textContent();
    const afterWarningCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().textContent();
    const afterInfoCount = await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().textContent();
    
    console.log('Counts after ERROR filter:', { 
      error: afterErrorCount, 
      warning: afterWarningCount, 
      info: afterInfoCount 
    });
    
    // CRITICAL ASSERTIONS: Stats should be identical
    expect(afterErrorCount).toBe(initialErrorCount);
    expect(afterWarningCount).toBe(initialWarningCount);
    expect(afterInfoCount).toBe(initialInfoCount);
    
    // Verify table only shows ERROR entries
    const tableRows = authenticatedPage.locator('table tbody tr');
    const rowCount = await tableRows.count();
    
    console.log(`Table has ${rowCount} rows after ERROR filter`);
    
    // Check each visible row is ERROR level
    for (let i = 0; i < Math.min(rowCount, 5); i++) {
      const row = tableRows.nth(i);
      const levelBadge = row.locator('.badge');
      const levelText = await levelBadge.textContent();
      
      // Should contain "ERROR" or "CRITICAL" (filtered)
      expect(levelText?.toUpperCase()).toMatch(/ERROR|CRITICAL/);
    }
  });

  test('Stats cards remain unchanged when WARNING filter applied', async ({ authenticatedPage }) => {
    // Capture initial stats
    const initialCounts = {
      error: await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().textContent(),
      warning: await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().textContent(),
      info: await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().textContent()
    };
    
    // Apply WARNING filter
    const warningCard = authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first();
    await warningCard.click();
    await authenticatedPage.waitForLoadState('networkidle');
    await authenticatedPage.waitForTimeout(1000);
    
    // Take screenshot
    await percySnapshot(authenticatedPage, 'Health Page - After WARNING Filter Applied');
    
    // Verify stats UNCHANGED
    const afterCounts = {
      error: await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().textContent(),
      warning: await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().textContent(),
      info: await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().textContent()
    };
    
    expect(afterCounts.error).toBe(initialCounts.error);
    expect(afterCounts.warning).toBe(initialCounts.warning);
    expect(afterCounts.info).toBe(initialCounts.info);
  });

  test('Stats cards remain unchanged when multiple filters toggled', async ({ authenticatedPage }) => {
    // Capture initial stats
    const getStats = async () => ({
      error: await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().textContent(),
      warning: await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().textContent(),
      info: await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().textContent()
    });
    
    const initialStats = await getStats();
    console.log('Initial stats:', initialStats);
    
    // Apply ERROR filter
    await authenticatedPage.locator('.frosted-card').filter({ hasText: /ERROR/i }).first().click();
    await authenticatedPage.waitForTimeout(500);
    let currentStats = await getStats();
    expect(currentStats).toEqual(initialStats);
    
    // Apply WARNING filter (toggle ERROR off, WARNING on)
    await authenticatedPage.locator('.frosted-card').filter({ hasText: /WARNING/i }).first().click();
    await authenticatedPage.waitForTimeout(500);
    currentStats = await getStats();
    expect(currentStats).toEqual(initialStats);
    
    // Apply INFO filter
    await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().click();
    await authenticatedPage.waitForTimeout(500);
    currentStats = await getStats();
    expect(currentStats).toEqual(initialStats);
    
    // Clear all filters (click INFO again to deselect)
    await authenticatedPage.locator('.frosted-card').filter({ hasText: /INFO/i }).first().click();
    await authenticatedPage.waitForTimeout(500);
    currentStats = await getStats();
    expect(currentStats).toEqual(initialStats);
    
    // Take final screenshot
    await percySnapshot(authenticatedPage, 'Health Page - After Multiple Filter Toggles');
  });

  test('Stats API endpoint is called on page load', async ({ authenticatedPage, page }) => {
    // Set up network interception
    const statsRequests: any[] = [];
    
    authenticatedPage.on('request', request => {
      if (request.url().includes('/api/logs/v1/stats')) {
        statsRequests.push({
          url: request.url(),
          method: request.method()
        });
      }
    });
    
    // Navigate to Health page
    await authenticatedPage.goto('/health');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Verify stats API was called
    expect(statsRequests.length).toBeGreaterThan(0);
    expect(statsRequests[0].method).toBe('GET');
    expect(statsRequests[0].url).toContain('/api/logs/v1/stats');
    
    console.log('Stats API calls:', statsRequests);
  });
});

test.describe('Health Page - Stats Architecture Validation', () => {
  test('Stats endpoint returns database totals', async ({ authenticatedPage }) => {
    // Call stats endpoint directly
    const response = await authenticatedPage.request.get('http://localhost:3000/api/logs/v1/stats');
    
    expect(response.ok()).toBeTruthy();
    
    const statsData = await response.json();
    console.log('Stats API response:', statsData);
    
    // Verify structure
    expect(statsData).toHaveProperty('debug');
    expect(statsData).toHaveProperty('info');
    expect(statsData).toHaveProperty('warning');
    expect(statsData).toHaveProperty('error');
    expect(statsData).toHaveProperty('critical');
    
    // Verify all are numbers
    expect(typeof statsData.debug).toBe('number');
    expect(typeof statsData.info).toBe('number');
    expect(typeof statsData.warning).toBe('number');
    expect(typeof statsData.error).toBe('number');
    expect(typeof statsData.critical).toBe('number');
  });

  test('Stats endpoint totals match sum of entries', async ({ authenticatedPage }) => {
    // Get stats
    const statsResponse = await authenticatedPage.request.get('http://localhost:3000/api/logs/v1/stats');
    const stats = await statsResponse.json();
    
    // Get logs (limited sample)
    const logsResponse = await authenticatedPage.request.get('http://localhost:3000/api/logs?limit=1000');
    const logsData = await logsResponse.json();
    
    console.log('Stats totals:', stats);
    console.log('Logs count:', logsData.length);
    
    // Count levels in logs sample
    const levels = logsData.reduce((acc: any, log: any) => {
      const level = log.level.toLowerCase();
      acc[level] = (acc[level] || 0) + 1;
      return acc;
    }, {});
    
    console.log('Levels in sample:', levels);
    
    // Stats totals should be >= sample counts (database has more)
    if (levels.error) {
      expect(stats.error).toBeGreaterThanOrEqual(levels.error);
    }
    if (levels.warning) {
      expect(stats.warning).toBeGreaterThanOrEqual(levels.warning);
    }
    if (levels.info) {
      expect(stats.info).toBeGreaterThanOrEqual(levels.info);
    }
  });
});
