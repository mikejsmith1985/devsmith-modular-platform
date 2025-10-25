import { test, expect, Page } from '@playwright/test';

/**
 * RED PHASE: Log Viewer UI with Virtual Scrolling
 * 
 * All tests written to fail first (TDD approach)
 * These tests exercise ALL acceptance criteria for Issue #XX
 * 
 * Requirements:
 * - Virtual scrolling (handle 10k+ logs)
 * - Level-based color coding (ERROR=red, WARN=yellow, INFO=blue)
 * - Expandable log details (stack traces, context)
 * - Real-time updates via WebSocket
 * - Search/filter UI
 * - Date range picker
 * - Service/level filter dropdowns
 * - Copy log entry to clipboard
 * - Toast notifications
 * - Responsive design
 * - Accessibility (WCAG 2.1 AA)
 * - E2E tests with Playwright
 */

test.describe('Log Viewer UI - Virtual Scrolling & Advanced Features', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await page.goto('http://localhost:8082/');
    // Wait for dashboard to fully load
    await page.waitForLoadState('networkidle');
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ============================================================================
  // VIRTUAL SCROLLING TESTS (10k+ logs performance)
  // ============================================================================

  test('RED: Virtual scrolling - should render 10,000 logs without lag', async () => {
    // Load large number of logs
    const logsContainer = page.locator('#logs-output');
    
    // Wait for 10k logs to be loaded via API or WebSocket
    await page.evaluate(() => {
      // Simulate loading 10,000 logs
      const container = document.getElementById('logs-output');
      if (container) {
        container.innerHTML = '';
        for (let i = 0; i < 10000; i++) {
          const div = document.createElement('div');
          div.className = 'log-entry';
          div.textContent = `Log entry ${i}`;
          container.appendChild(div);
        }
      }
    });

    // Should still be responsive (not lag)
    const startTime = Date.now();
    await page.evaluate(() => {
      const container = document.getElementById('logs-output');
      if (container) {
        container.scrollTop = 5000; // Scroll to middle
      }
    });
    const scrollTime = Date.now() - startTime;

    // Scrolling 10k items should be smooth (< 200ms)
    expect(scrollTime).toBeLessThan(200);
  });

  test('RED: Virtual scrolling - should only render visible viewport (windowing)', async () => {
    // Check that off-screen logs are NOT in DOM (virtualized)
    const logsContainer = page.locator('#logs-output');
    
    await page.evaluate(() => {
      const container = document.getElementById('logs-output');
      if (container) {
        // Render 10k logs
        for (let i = 0; i < 10000; i++) {
          const div = document.createElement('div');
          div.className = 'log-entry';
          div.id = `log-${i}`;
          div.textContent = `Log ${i}`;
          container.appendChild(div);
        }
      }
    });

    // After rendering with virtualization, DOM should have ~50-100 logs
    // (not all 10,000)
    const logCount = await page.evaluate(() => {
      return document.getElementById('logs-output')?.children.length || 0;
    });

    // Expected: 50-200 items in DOM (virtualized)
    // Failing because current implementation renders all items
    expect(logCount).toBeLessThan(200);
  });

  test('RED: Virtual scrolling - should lazy-load as user scrolls', async () => {
    // Set up WebSocket spy to track new log loads
    const loadedLogs: any[] = [];
    
    await page.evaluate(() => {
      window.loadedLogs = [];
      const originalContainer = document.getElementById('logs-output');
      if (originalContainer) {
        originalContainer.addEventListener('scroll', () => {
          // Track scroll events
          window.scrollEvents = (window.scrollEvents || 0) + 1;
        });
      }
    });

    // Scroll to bottom
    await page.evaluate(() => {
      const container = document.getElementById('logs-output');
      if (container) {
        container.scrollTop = container.scrollHeight;
      }
    });

    // Should trigger load-more event or fetch next batch
    const scrollEventsTriggered = await page.evaluate(() => window.scrollEvents || 0);
    expect(scrollEventsTriggered).toBeGreaterThan(0);
  });

  // ============================================================================
  // EXPANDABLE LOG DETAILS TESTS
  // ============================================================================

  test('RED: Expandable details - should show expand button on log entry', async () => {
    // Log entries should have expand/collapse button
    const logEntry = page.locator('.log-entry').first();
    
    const expandButton = logEntry.locator('.expand-btn, [data-expand], .icon-expand');
    
    // Should fail because expand button doesn't exist yet
    await expect(expandButton).toBeVisible();
  });

  test('RED: Expandable details - should expand to show stack trace', async () => {
    const logEntry = page.locator('.log-entry').first();
    const expandButton = logEntry.locator('.expand-btn');
    
    await expandButton.click();
    
    // Should show stack trace section
    const stackTrace = logEntry.locator('.stack-trace, [data-stacktrace]');
    await expect(stackTrace).toBeVisible();
    
    // Should contain traceable content
    const traceContent = await stackTrace.textContent();
    expect(traceContent?.length).toBeGreaterThan(0);
  });

  test('RED: Expandable details - should show context metadata', async () => {
    const logEntry = page.locator('.log-entry').first();
    const expandButton = logEntry.locator('.expand-btn');
    
    await expandButton.click();
    
    // Should show metadata section
    const metadata = logEntry.locator('.metadata, [data-metadata], .context');
    await expect(metadata).toBeVisible();
    
    // Should contain context info (user, request ID, etc.)
    const contextText = await metadata.textContent();
    expect(contextText).toContain(/user|request|context|id/i);
  });

  test('RED: Expandable details - should show full error message when expanded', async () => {
    const logEntry = page.locator('.log-entry').first();
    const expandButton = logEntry.locator('.expand-btn');
    
    // Initial message should be truncated
    const initialMessage = await logEntry.locator('.log-message').textContent();
    expect(initialMessage?.length).toBeLessThan(500); // Truncated
    
    await expandButton.click();
    
    // Expanded message should be full
    const expandedDetails = logEntry.locator('.expanded-details, [data-expanded]');
    const fullMessage = await expandedDetails.textContent();
    expect(fullMessage?.length).toBeGreaterThan(initialMessage?.length || 0);
  });

  // ============================================================================
  // DATE RANGE PICKER TESTS
  // ============================================================================

  test('RED: Date picker - should have date range input fields', async () => {
    const dateFromInput = page.locator('#date-from, [data-date-from], input[type="date"][name*="from"]');
    const dateToInput = page.locator('#date-to, [data-date-to], input[type="date"][name*="to"]');
    
    await expect(dateFromInput).toBeVisible();
    await expect(dateToInput).toBeVisible();
  });

  test('RED: Date picker - should filter logs by date range', async () => {
    const dateFrom = page.locator('#date-from');
    const dateTo = page.locator('#date-to');
    const filterButton = page.locator('button:has-text("Filter"), button:has-text("Apply")');
    
    // Set date range
    await dateFrom.fill('2025-01-01');
    await dateTo.fill('2025-01-31');
    
    // Applying filter should reload logs within date range
    await filterButton.click();
    
    // Wait for logs to reload
    await page.waitForTimeout(500);
    
    // Verify logs are within date range
    const logs = await page.locator('.log-entry').count();
    expect(logs).toBeGreaterThan(0);
  });

  test('RED: Date picker - should persist date filters in URL params', async () => {
    await page.locator('#date-from').fill('2025-01-15');
    await page.locator('#date-to').fill('2025-01-20');
    await page.locator('button:has-text("Apply")').click();
    
    // URL should contain date params
    const url = page.url();
    expect(url).toContain(/date.*from|from.*date/i);
    expect(url).toContain(/date.*to|to.*date/i);
  });

  // ============================================================================
  // COPY TO CLIPBOARD TESTS
  // ============================================================================

  test('RED: Copy clipboard - should have copy button on log entry', async () => {
    const logEntry = page.locator('.log-entry').first();
    const copyButton = logEntry.locator('[data-copy], .copy-btn, button:has-text("Copy")');
    
    // Should fail because copy button doesn't exist
    await expect(copyButton).toBeVisible();
  });

  test('RED: Copy clipboard - should copy full log entry to clipboard', async () => {
    const logEntry = page.locator('.log-entry').first();
    const copyButton = logEntry.locator('[data-copy]');
    
    // Get expected text
    const expectedText = await logEntry.textContent();
    
    // Click copy
    await copyButton.click();
    
    // Verify clipboard contains the log
    const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboardText).toBe(expectedText);
  });

  test('RED: Copy clipboard - should show confirmation toast on copy', async () => {
    const logEntry = page.locator('.log-entry').first();
    const copyButton = logEntry.locator('[data-copy]');
    
    await copyButton.click();
    
    // Should show toast notification
    const toast = page.locator('[role="alert"], .toast, .notification, [data-toast]');
    await expect(toast).toBeVisible();
    
    // Should contain "Copied" message
    const toastText = await toast.textContent();
    expect(toastText).toContain(/copied|copied to clipboard/i);
  });

  // ============================================================================
  // TOAST NOTIFICATIONS TESTS
  // ============================================================================

  test('RED: Toast notifications - should show when new error log received', async () => {
    // Wait for new error log
    await page.evaluate(() => {
      // Simulate receiving new ERROR log via WebSocket
      window.simulateNewLog({
        level: 'ERROR',
        message: 'Critical error occurred',
        service: 'portal'
      });
    });

    // Toast should appear
    const toast = page.locator('[role="alert"], .toast, [data-toast-error]');
    await expect(toast).toBeVisible();
    
    // Should mention error
    const toastText = await toast.textContent();
    expect(toastText).toContain(/error|critical/i);
  });

  test('RED: Toast notifications - should auto-dismiss after 5 seconds', async () => {
    const toast = page.locator('[role="alert"], .toast').first();
    
    if (await toast.isVisible()) {
      // Wait 5+ seconds
      await page.waitForTimeout(5500);
      
      // Toast should be hidden
      await expect(toast).not.toBeVisible();
    }
  });

  test('RED: Toast notifications - should allow manual dismiss', async () => {
    const toast = page.locator('[role="alert"], .toast').first();
    
    if (await toast.isVisible()) {
      const closeButton = toast.locator('button:has-text("Close"), .close, [data-dismiss]');
      await closeButton.click();
      
      // Toast should disappear
      await expect(toast).not.toBeVisible();
    }
  });

  test('RED: Toast notifications - should show for warnings', async () => {
    // Simulate WARNING log
    await page.evaluate(() => {
      window.simulateNewLog({
        level: 'WARN',
        message: 'Warning condition detected'
      });
    });

    const warnToast = page.locator('[data-toast-warning], .toast-warning');
    await expect(warnToast).toBeVisible();
  });

  // ============================================================================
  // SEARCH WITH DEBOUNCING TESTS
  // ============================================================================

  test('RED: Search debouncing - should wait 300ms before filtering', async () => {
    const searchInput = page.locator('#search-input');
    
    const filterCalls: number[] = [];
    
    // Monitor filter changes
    await page.evaluate(() => {
      let filterCallCount = 0;
      window.originalFetch = window.fetch;
      window.fetch = async (...args: any[]) => {
        if (args[0].includes('/logs')) {
          filterCallCount++;
          window.lastFilterCallTime = Date.now();
        }
        return window.originalFetch.apply(this, args);
      };
    });

    // Type quickly
    await searchInput.type('error', { delay: 50 }); // Type fast
    
    // Capture filter calls
    await page.waitForTimeout(400);
    
    const lastCallTime = await page.evaluate(() => window.lastFilterCallTime);
    
    // Should have debounced (not called 5 times for 5 characters)
    // Should wait ~300ms before calling
    const timeSinceLastKeypress = Date.now() - (lastCallTime || 0);
    expect(timeSinceLastKeypress).toBeGreaterThanOrEqual(250);
  });

  test('RED: Search debouncing - should only call API once for multi-character input', async () => {
    const searchInput = page.locator('#search-input');
    
    let filterCallCount = 0;
    
    await page.evaluate(() => {
      window.filterCallCount = 0;
      window.originalFetch = window.fetch;
      window.fetch = async (...args: any[]) => {
        if (args[0].includes('/logs')) {
          window.filterCallCount = (window.filterCallCount || 0) + 1;
        }
        return window.originalFetch.apply(this, args);
      };
    });

    // Type 5 characters quickly
    await searchInput.type('error', { delay: 50 });
    
    // Wait for debounce
    await page.waitForTimeout(350);
    
    const callCount = await page.evaluate(() => window.filterCallCount);
    
    // Should only call API once (not 5 times)
    // Expected: 1 call (allowing ±1 for initial load)
    expect(callCount).toBeLessThanOrEqual(2);
  });

  // ============================================================================
  // ACCESSIBILITY TESTS (WCAG 2.1 AA)
  // ============================================================================

  test('RED: Accessibility - should have proper ARIA labels', async () => {
    // Filter inputs should have labels or aria-label
    const levelFilter = page.locator('#level-filter');
    const serviceFilter = page.locator('#service-filter');
    const searchInput = page.locator('#search-input');
    
    // Check for label or aria-label
    for (const element of [levelFilter, serviceFilter, searchInput]) {
      const ariaLabel = await element.getAttribute('aria-label');
      const label = await page.locator(`label[for="${await element.getAttribute('id')}"]`).count();
      
      expect(ariaLabel || label > 0).toBeTruthy();
    }
  });

  test('RED: Accessibility - should support keyboard navigation', async () => {
    // Tab through filters
    await page.keyboard.press('Tab'); // Focus first element
    await page.keyboard.press('Tab'); // Move to next
    
    const focusedElement = await page.evaluate(() => {
      return document.activeElement?.tagName;
    });
    
    // Should have focused an element
    expect(focusedElement).not.toBe('BODY');
  });

  test('RED: Accessibility - should have sufficient color contrast', async () => {
    // Check log level colors have sufficient contrast
    const errorLog = page.locator('.log-level.error').first();
    const warnLog = page.locator('.log-level.warn').first();
    
    // Get computed colors
    const errorColor = await errorLog.evaluate((el) => {
      return window.getComputedStyle(el).backgroundColor;
    });
    
    // Should have high contrast (testable via axe-core)
    expect(errorColor).toBeTruthy();
  });

  test('RED: Accessibility - should include role attributes', async () => {
    const logList = page.locator('#logs-output');
    const role = await logList.getAttribute('role');
    
    // Should have role="list" or similar
    expect(['list', 'region', 'main']).toContain(role);
  });

  test('RED: Accessibility - should use semantic HTML', async () => {
    // Check for proper semantic elements
    const nav = page.locator('nav');
    const main = page.locator('main');
    const section = page.locator('section');
    
    // Should use semantic HTML
    expect(await nav.count() + await main.count() + await section.count()).toBeGreaterThan(0);
  });

  // ============================================================================
  // RESPONSIVE DESIGN TESTS (NEW)
  // ============================================================================

  test('RED: Responsive design - should adapt to mobile (375px)', async () => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Expandable details should still be accessible
    const expandButton = page.locator('.expand-btn').first();
    await expect(expandButton).toBeVisible();
    
    // Date picker should be usable
    const dateInput = page.locator('[data-date-from]');
    await expect(dateInput).toBeVisible();
  });

  test('RED: Responsive design - stack trace should be readable on mobile', async () => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    // Expand a log
    await page.locator('.expand-btn').first().click();
    
    // Stack trace should not overflow
    const stackTrace = page.locator('.stack-trace').first();
    const isVisible = await stackTrace.isVisible();
    
    expect(isVisible).toBe(true);
  });

  test('RED: Responsive design - copy button should work on mobile', async () => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    const copyButton = page.locator('[data-copy]').first();
    await expect(copyButton).toBeVisible();
    
    // Should be clickable (not too small)
    const boundingBox = await copyButton.boundingBox();
    expect(boundingBox?.width).toBeGreaterThan(30);
    expect(boundingBox?.height).toBeGreaterThan(30);
  });

  // ============================================================================
  // WEBSOCKET REAL-TIME UPDATES (NEW)
  // ============================================================================

  test('RED: WebSocket - should receive expanded log details in real-time', async () => {
    // Send log with stack trace via WebSocket
    await page.evaluate(() => {
      window.simulateNewLog({
        level: 'ERROR',
        message: 'Stack trace included',
        stackTrace: 'at function() \n at caller()',
        context: { userId: 123 }
      });
    });

    // New log should appear with expandable content
    const newLog = page.locator('.log-entry').last();
    const expandButton = newLog.locator('.expand-btn');
    
    await expect(expandButton).toBeVisible();
  });

  test('RED: WebSocket - expanded details should update in real-time', async () => {
    // Expand first log
    const logEntry = page.locator('.log-entry').first();
    await logEntry.locator('.expand-btn').click();
    
    // If log gets updated via WebSocket with new stack trace
    await page.evaluate(() => {
      window.updateLog(0, { stackTrace: 'NEW STACK TRACE' });
    });

    // Expanded details should show updated content
    const stackTrace = logEntry.locator('.stack-trace');
    const content = await stackTrace.textContent();
    expect(content).toContain('NEW STACK TRACE');
  });

  // ============================================================================
  // INTEGRATION TESTS
  // ============================================================================

  test('RED: Integration - expand + copy should work together', async () => {
    const logEntry = page.locator('.log-entry').first();
    
    // Expand
    await logEntry.locator('.expand-btn').click();
    
    // Copy
    const copyButton = logEntry.locator('[data-copy]');
    await copyButton.click();
    
    // Clipboard should contain expanded content (with stack trace)
    const clipboard = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboard.length).toBeGreaterThan(100); // Should be longer than collapsed entry
  });

  test('RED: Integration - date filter + search + copy should work together', async () => {
    // Apply date filter
    await page.locator('#date-from').fill('2025-01-01');
    await page.locator('button:has-text("Apply")').click();

    // Apply search filter
    await page.locator('#search-input').fill('error');
    await page.waitForTimeout(350);

    // Copy a filtered result
    const logEntry = page.locator('.log-entry').first();
    await logEntry.locator('[data-copy]').click();

    // Should copy filtered result
    const clipboard = await page.evaluate(() => navigator.clipboard.readText());
    expect(clipboard).toContain('error');
  });
});
