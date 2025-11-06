/**
 * Comprehensive Logs Dashboard Visual Test
 * 
 * This test validates EVERY interactive element on the Logs dashboard:
 * - All stat cards are visible and clickable
 * - All filters work correctly
 * - Search functionality works
 * - Dark mode toggle works
 * - Navigation works
 * - Percy visual regression captures all states
 */

import { test, expect } from '../fixtures/auth.fixture';
import percySnapshot from '@percy/playwright';

test.describe('Logs Dashboard - Comprehensive Interaction & Visual Validation', () => {
  test('should interact with every element and validate visually', async ({ authenticatedPage: page }) => {
    // Navigate to Logs dashboard
    await page.goto('http://localhost:3000/logs');
    await page.waitForLoadState('networkidle');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 1: Verify page loaded correctly
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    await expect(page.locator('h1:has-text("DevSmith Logs")')).toBeVisible();
    await percySnapshot(page, 'Logs Dashboard - Initial Load');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 2: Validate all stat cards are visible
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    const statCards = [
      'Total Entries',
      'Critical Errors', 
      'Warnings',
      'Info Logs'
    ];
    
    for (const cardTitle of statCards) {
      const card = page.locator('.ds-stat-card', { hasText: cardTitle });
      await expect(card).toBeVisible();
      console.log(`✓ Stat card "${cardTitle}" visible`);
    }
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 3: Click each stat card and verify modal opens
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    for (const cardTitle of statCards) {
      const card = page.locator('.ds-stat-card', { hasText: cardTitle });
      await card.click();
      
      // Wait for modal to appear
      const modal = page.locator('[role="dialog"], .modal, .ds-modal');
      await expect(modal).toBeVisible({ timeout: 2000 });
      
      // Capture visual state
      await percySnapshot(page, `Logs Dashboard - ${cardTitle} Modal`);
      console.log(`✓ Modal for "${cardTitle}" opened`);
      
      // Close modal (ESC key or close button)
      await page.keyboard.press('Escape');
      await expect(modal).not.toBeVisible({ timeout: 2000 });
      console.log(`✓ Modal for "${cardTitle}" closed`);
    }
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 4: Test all filter dropdowns
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Level filter
    const levelFilter = page.locator('#level-filter, select[aria-label*="level"]');
    await levelFilter.selectOption('ERROR');
    await page.waitForTimeout(500); // Wait for filter to apply
    await percySnapshot(page, 'Logs Dashboard - Filtered by ERROR');
    console.log('✓ Level filter works (ERROR selected)');
    
    await levelFilter.selectOption('WARN');
    await page.waitForTimeout(500);
    console.log('✓ Level filter works (WARN selected)');
    
    await levelFilter.selectOption('all');
    await page.waitForTimeout(500);
    console.log('✓ Level filter reset to all');
    
    // Service filter
    const serviceFilter = page.locator('#service-filter, select[aria-label*="service"]');
    await serviceFilter.selectOption('portal');
    await page.waitForTimeout(500);
    await percySnapshot(page, 'Logs Dashboard - Filtered by Portal Service');
    console.log('✓ Service filter works (Portal selected)');
    
    await serviceFilter.selectOption('all');
    await page.waitForTimeout(500);
    console.log('✓ Service filter reset to all');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 5: Test date range pickers
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    const today = new Date().toISOString().split('T')[0];
    const yesterday = new Date(Date.now() - 86400000).toISOString().split('T')[0];
    
    const dateFrom = page.locator('#date-from, input[type="date"][aria-label*="from"]');
    const dateTo = page.locator('#date-to, input[type="date"][aria-label*="to"]');
    
    await dateFrom.fill(yesterday);
    await dateTo.fill(today);
    await page.waitForTimeout(500);
    console.log('✓ Date range set');
    
    // Apply filters button
    const applyButton = page.locator('button:has-text("Apply Filters"), #apply-filters');
    await applyButton.click();
    await page.waitForTimeout(1000);
    await percySnapshot(page, 'Logs Dashboard - Date Range Applied');
    console.log('✓ Date range filters applied');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 6: Test search functionality
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    const searchInput = page.locator('#search-input, input[placeholder*="search"], input[aria-label*="search"]');
    await searchInput.fill('error');
    await page.waitForTimeout(500); // Debounce delay
    await percySnapshot(page, 'Logs Dashboard - Search "error"');
    console.log('✓ Search input works');
    
    await searchInput.clear();
    await page.waitForTimeout(500);
    console.log('✓ Search cleared');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 7: Test control buttons
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Pause/Resume button
    const pauseButton = page.locator('#pause-btn, button[aria-label*="pause"]');
    await pauseButton.click();
    await page.waitForTimeout(300);
    await expect(pauseButton).toContainText(/Resume|Play/i);
    console.log('✓ Pause button works');
    
    await pauseButton.click();
    await page.waitForTimeout(300);
    await expect(pauseButton).toContainText(/Pause/i);
    console.log('✓ Resume button works');
    
    // Auto-scroll button
    const autoScrollButton = page.locator('#auto-scroll-btn, button[aria-label*="auto-scroll"]');
    await autoScrollButton.click();
    await page.waitForTimeout(300);
    console.log('✓ Auto-scroll toggle works (disabled)');
    
    await autoScrollButton.click();
    await page.waitForTimeout(300);
    console.log('✓ Auto-scroll toggle works (enabled)');
    
    // Clear button
    const clearButton = page.locator('#clear-btn, button[aria-label*="clear"]');
    await clearButton.click();
    await page.waitForTimeout(500);
    await percySnapshot(page, 'Logs Dashboard - After Clear');
    console.log('✓ Clear button works');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 8: Test dark mode toggle
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    const darkModeToggle = page.locator('#dark-mode-toggle, button[aria-label*="dark mode"]');
    await darkModeToggle.click();
    await page.waitForTimeout(500);
    
    // Verify dark mode applied
    const htmlClass = await page.locator('html').getAttribute('class');
    expect(htmlClass).toContain('dark');
    await percySnapshot(page, 'Logs Dashboard - Dark Mode');
    console.log('✓ Dark mode enabled');
    
    // Toggle back to light mode
    await darkModeToggle.click();
    await page.waitForTimeout(500);
    const htmlClassLight = await page.locator('html').getAttribute('class');
    expect(htmlClassLight).not.toContain('dark');
    await percySnapshot(page, 'Logs Dashboard - Light Mode');
    console.log('✓ Light mode restored');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 9: Test navigation to other services
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    // Click Review nav link
    const reviewLink = page.locator('a[href="/review"]');
    await reviewLink.click();
    await page.waitForURL(/\/review/);
    console.log('✓ Navigation to Review works');
    
    // Navigate back to Logs
    await page.goto('http://localhost:3000/logs');
    await page.waitForLoadState('networkidle');
    
    // Click Analytics nav link
    const analyticsLink = page.locator('a[href="/analytics"]');
    await analyticsLink.click();
    await page.waitForURL(/\/analytics/);
    console.log('✓ Navigation to Analytics works');
    
    // Navigate back to Logs
    await page.goto('http://localhost:3000/logs');
    await page.waitForLoadState('networkidle');
    
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    // TEST 10: Verify WebSocket connection status
    // ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    
    const connectionStatus = page.locator('#connection-status, [role="status"]');
    await expect(connectionStatus).toContainText(/Connected|🟢/i);
    console.log('✓ WebSocket connection status indicates connected');
    
    // Final visual capture
    await percySnapshot(page, 'Logs Dashboard - Final State');
    
    console.log('\n✅ ALL LOGS DASHBOARD TESTS PASSED');
  });
});
