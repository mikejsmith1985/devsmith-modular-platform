import { test } from '@playwright/test';

/**
 * UI Screenshot Capture Test
 * Captures screenshots of all services to verify visual appearance
 */

test.describe('UI Screenshot Capture', () => {
  test('Capture all service UIs', async ({ page }) => {
    // Portal
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/portal-homepage.png', fullPage: true });
    console.log('✓ Portal screenshot saved to test-results/portal-homepage.png');

    // Logs
    await page.goto('/logs');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/logs-dashboard.png', fullPage: true });
    console.log('✓ Logs screenshot saved to test-results/logs-dashboard.png');

    // Review
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/review-workspace.png', fullPage: true });
    console.log('✓ Review screenshot saved to test-results/review-workspace.png');

    // Analytics
    await page.goto('/analytics');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/analytics-dashboard.png', fullPage: true });
    console.log('✓ Analytics screenshot saved to test-results/analytics-dashboard.png');
  });
});
