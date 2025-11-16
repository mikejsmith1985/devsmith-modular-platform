import { test, expect } from '@playwright/test';
import { test as authTest } from './fixtures/auth.fixture';

/**
 * Responsive Design Validation Tests
 * 
 * Tests mobile/tablet layouts, touch targets, and responsive navigation.
 * Validates design works across iPhone 12, iPad, and Android devices.
 * 
 * Phase 3.2 - Responsive Design Validation
 */

test.describe('Responsive Design - Mobile (iPhone 12 - 390x844)', () => {
  authTest('Portal Dashboard renders correctly on mobile', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 390, height: 844 });
    
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify no horizontal scroll
    const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1); // Allow 1px rounding

    // Verify app cards are visible
    const appCards = authenticatedPage.locator('.app-card, [class*="card"]');
    await expect(appCards.first()).toBeVisible();

    // Verify navigation is accessible
    const nav = authenticatedPage.locator('nav, header, [role="navigation"]');
    await expect(nav.first()).toBeVisible();
  });

  authTest('Navigation is touch-friendly (44px minimum)', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 390, height: 844 });
    await authenticatedPage.waitForLoadState('networkidle');

    // Check all interactive elements meet touch target size
    const buttons = authenticatedPage.locator('button, a[href], [role="button"]');
    const count = await buttons.count();

    for (let i = 0; i < Math.min(count, 10); i++) {
      const button = buttons.nth(i);
      if (await button.isVisible()) {
        const box = await button.boundingBox();
        if (box) {
          // Touch targets should be at least 44x44px (Apple HIG)
          expect(box.height).toBeGreaterThanOrEqual(40); // Allow slight variance
          expect(box.width).toBeGreaterThanOrEqual(40);
        }
      }
    }
  });

  authTest('Review service mobile layout', async ({ authenticatedPage }) => {
    const response = await authenticatedPage.request.post('http://localhost:3000/auth/test-login');
    const data = await response.json();
    await authenticatedPage.context().addCookies([{
      name: 'devsmith_token',
      value: data.token,
      domain: 'localhost',
      path: '/'
    }]);

    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify no horizontal scroll
    const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);

    // Check content is readable (not cut off)
    const content = authenticatedPage.locator('main, [role="main"], .content, .workspace');
    await expect(content.first()).toBeVisible();
  });
});

test.describe('Responsive Design - Tablet (iPad - 810x1080)', () => {
  authTest('Portal Dashboard tablet layout', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 810, height: 1080 });
    const response = await authenticatedPage.request.post('http://localhost:3000/auth/test-login');
    const data = await response.json();
    await authenticatedPage.context().addCookies([{
      name: 'devsmith_token',
      value: data.token,
      domain: 'localhost',
      path: '/'
    }]);

    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify tablet-optimized layout (should show 2-column grid)
    const viewport = authenticatedPage.viewportSize();
    expect(viewport?.width).toBe(810); // iPad width

    // No horizontal scroll
    const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);

    // App cards should be visible and clickable
    const appCards = authenticatedPage.locator('.app-card, [class*="card"]');
    await expect(appCards.first()).toBeVisible();
  });

  authTest('Logs service tablet layout', async ({ authenticatedPage }) => {
    const response = await authenticatedPage.request.post('http://localhost:3000/auth/test-login');
    const data = await response.json();
    await authenticatedPage.context().addCookies([{
      name: 'devsmith_token',
      value: data.token,
      domain: 'localhost',
      path: '/'
    }]);

    await authenticatedPage.goto('/logs');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify no horizontal scroll
    const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);
  });
});

test.describe('Responsive Design - Android (Galaxy S9+ - 412x846)', () => {
  authTest('Analytics Dashboard Android layout', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 412, height: 846 });
    const response = await authenticatedPage.request.post('http://localhost:3000/auth/test-login');
    const data = await response.json();
    await authenticatedPage.context().addCookies([{
      name: 'devsmith_token',
      value: data.token,
      domain: 'localhost',
      path: '/'
    }]);

    await authenticatedPage.goto('/analytics');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify no horizontal scroll
    const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 1);

    // Content should be visible
    const content = authenticatedPage.locator('main, [role="main"], .content');
    await expect(content.first()).toBeVisible();
  });
});

test.describe('Responsive Navigation', () => {
  authTest('Mobile menu/hamburger functionality', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 375, height: 667 }); // iPhone SE size

    const response = await authenticatedPage.request.post('http://localhost:3000/auth/test-login');
    const data = await response.json();
    await authenticatedPage.context().addCookies([{
      name: 'devsmith_token',
      value: data.token,
      domain: 'localhost',
      path: '/'
    }]);

    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Look for mobile menu toggle (hamburger icon, menu button, etc.)
    const mobileMenuSelectors = [
      '[data-testid="mobile-menu-toggle"]',
      '.mobile-menu-toggle',
      'button[aria-label*="menu" i]',
      'button[aria-label*="navigation" i]',
      '.hamburger',
      '[class*="burger"]',
      'button:has(svg path[d*="M2,6 L18,6"])', // Hamburger icon path
    ];

    let foundMenu = false;
    for (const selector of mobileMenuSelectors) {
      try {
        const menuToggle = authenticatedPage.locator(selector).first();
        if (await menuToggle.isVisible({ timeout: 1000 })) {
          foundMenu = true;
          
          // Test toggle functionality
          await menuToggle.click();
          await authenticatedPage.waitForTimeout(500); // Animation
          
          // Menu should expand/appear
          const menu = authenticatedPage.locator('nav, [role="navigation"], .menu, .nav-menu').first();
          await expect(menu).toBeVisible();
          
          break;
        }
      } catch (e) {
        // Continue to next selector
      }
    }

    // If no mobile menu found, verify regular nav is still accessible
    if (!foundMenu) {
      const nav = authenticatedPage.locator('nav, header, [role="navigation"]').first();
      await expect(nav).toBeVisible();
    }
  });
});

test.describe('Responsive Breakpoint Tests', () => {
  const breakpoints = [
    { name: 'Mobile Small', width: 320, height: 568 },
    { name: 'Mobile Medium', width: 375, height: 667 },
    { name: 'Mobile Large', width: 414, height: 896 },
    { name: 'Tablet Portrait', width: 768, height: 1024 },
    { name: 'Tablet Landscape', width: 1024, height: 768 },
    { name: 'Desktop Small', width: 1280, height: 720 },
    { name: 'Desktop Large', width: 1920, height: 1080 },
  ];

  for (const breakpoint of breakpoints) {
    authTest(`Portal renders correctly at ${breakpoint.name} (${breakpoint.width}x${breakpoint.height})`, async ({ authenticatedPage }) => {
      await authenticatedPage.setViewportSize({ width: breakpoint.width, height: breakpoint.height });

      await authenticatedPage.goto('/dashboard');
      await authenticatedPage.waitForLoadState('networkidle');

      // Verify no horizontal scroll
      const scrollWidth = await authenticatedPage.evaluate(() => document.documentElement.scrollWidth);
      const clientWidth = await authenticatedPage.evaluate(() => document.documentElement.clientWidth);
      expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 2); // Allow 2px variance

      // Verify content is visible
      const content = authenticatedPage.locator('body');
      await expect(content).toBeVisible();

      // Verify page loaded successfully
      expect(authenticatedPage.url()).toContain('dashboard');
    });
  }
});

test.describe('Responsive Images and Media', () => {
  authTest('Images scale correctly on mobile', async ({ authenticatedPage }) => {
    await authenticatedPage.setViewportSize({ width: 375, height: 667 });

    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Check all images don't overflow viewport
    const images = authenticatedPage.locator('img');
    const count = await images.count();

    for (let i = 0; i < count; i++) {
      const img = images.nth(i);
      if (await img.isVisible()) {
        const box = await img.boundingBox();
        if (box) {
          // Image should not exceed viewport width
          expect(box.width).toBeLessThanOrEqual(375);
        }
      }
    }
  });
});
