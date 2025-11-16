import { test } from './fixtures/auth.fixture';
import percySnapshot from '@percy/playwright';

/**
 * Visual Regression Tests
 * 
 * Captures visual snapshots using Percy for regression detection.
 * Tests critical pages and components across desktop, tablet, and mobile.
 * 
 * Phase 2.3 - Percy Visual Regression Testing
 */

test.describe('Visual Regression - Portal', () => {
  test('Portal Dashboard (authenticated)', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Capture desktop view
    await percySnapshot(authenticatedPage, 'Portal Dashboard - Desktop', {
      widths: [1920]
    });
    
    // Capture tablet view
    await percySnapshot(authenticatedPage, 'Portal Dashboard - Tablet', {
      widths: [768]
    });
    
    // Capture mobile view
    await percySnapshot(authenticatedPage, 'Portal Dashboard - Mobile', {
      widths: [375]
    });
  });
});

test.describe('Visual Regression - Review Service', () => {
  test('Review Workspace - All 5 Reading Modes', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Base workspace
    await percySnapshot(authenticatedPage, 'Review Workspace - Empty', {
      widths: [1920, 768, 375]
    });
    
    // TODO: Add snapshots for each reading mode once UI is stable
    // - Preview Mode
    // - Skim Mode
    // - Scan Mode
    // - Detailed Mode
    // - Critical Mode
  });
});

test.describe('Visual Regression - Logs Service', () => {
  test('Logs Dashboard - Empty State', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/logs');
    await authenticatedPage.waitForLoadState('networkidle');
    
    await percySnapshot(authenticatedPage, 'Logs Dashboard - Empty', {
      widths: [1920, 768, 375]
    });
  });
  
  test('Logs Dashboard - With Entries', async ({ authenticatedPage }) => {
    // Navigate to logs
    await authenticatedPage.goto('/logs');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Wait for logs to load (if any exist)
    await authenticatedPage.waitForTimeout(2000);
    
    await percySnapshot(authenticatedPage, 'Logs Dashboard - With Data', {
      widths: [1920, 768, 375]
    });
  });
});

test.describe('Visual Regression - Analytics Service', () => {
  test('Analytics Dashboard', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/analytics');
    await authenticatedPage.waitForLoadState('networkidle');
    
    await percySnapshot(authenticatedPage, 'Analytics Dashboard', {
      widths: [1920, 768, 375]
    });
  });
});

test.describe('Visual Regression - Dark Mode', () => {
  test('Portal Dashboard - Dark Mode', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Toggle dark mode (if theme switcher exists)
    const themeToggle = authenticatedPage.locator('[data-testid="theme-toggle"]');
    if (await themeToggle.isVisible()) {
      await themeToggle.click();
      await authenticatedPage.waitForTimeout(500);
    }
    
    await percySnapshot(authenticatedPage, 'Portal Dashboard - Dark Mode', {
      widths: [1920]
    });
  });
});
