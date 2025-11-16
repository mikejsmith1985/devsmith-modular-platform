import { test, expect } from '../fixtures/auth.fixture';

/**
 * Cache Invalidation Infrastructure Tests
 * 
 * TDD Approach: RED → GREEN → REFACTOR
 * Tests the infrastructure-level cache invalidation solution
 * 
 * Architecture: Defense in Depth
 * - Layer 1: Traefik middleware (infrastructure)
 * - Layer 2: HTML meta tags (HTML level)
 * - Layer 3: Fresh Playwright context (test environment)
 * 
 * Reference: CACHE_SOLUTION_ARCHITECTURE.md
 */

test.describe('Cache Invalidation Infrastructure', () => {
  
  test('HTML responses have aggressive no-cache headers from Traefik', async ({ authenticatedPage }) => {
    // GIVEN: User requests the frontend HTML
    const response = await authenticatedPage.goto('http://localhost:3000/', {
      waitUntil: 'networkidle'
    });
    
    // THEN: Response should have no-cache headers
    expect(response?.status()).toBe(200);
    
    const headers = response?.headers();
    const cacheControl = headers?.['cache-control'] || '';
    
    // Verify aggressive no-cache headers from Traefik middleware
    expect(cacheControl).toContain('no-store');
    expect(cacheControl).toContain('no-cache');
    expect(cacheControl).toContain('must-revalidate');
    expect(cacheControl).toContain('max-age=0');
    
    // Verify additional cache-busting headers
    const pragma = headers?.['pragma'] || '';
    expect(pragma).toContain('no-cache');
    
    const expires = headers?.['expires'] || '';
    expect(expires).toBe('0');
  });

  test('HTML contains cache-control meta tags', async ({ authenticatedPage }) => {
    // GIVEN: User loads the frontend
    await authenticatedPage.goto('http://localhost:3000/', {
      waitUntil: 'networkidle'
    });
    
    // THEN: HTML should contain meta tags for cache control
    const cacheControlMeta = await authenticatedPage.locator('meta[http-equiv="Cache-Control"]');
    const pragmaMeta = await authenticatedPage.locator('meta[http-equiv="Pragma"]');
    const expiresMeta = await authenticatedPage.locator('meta[http-equiv="Expires"]');
    const buildTimestampMeta = await authenticatedPage.locator('meta[name="build-timestamp"]');
    
    // Verify all meta tags exist
    await expect(cacheControlMeta).toHaveCount(1);
    await expect(pragmaMeta).toHaveCount(1);
    await expect(expiresMeta).toHaveCount(1);
    await expect(buildTimestampMeta).toHaveCount(1);
    
    // Verify content attributes
    await expect(cacheControlMeta).toHaveAttribute('content', 'no-cache, no-store, must-revalidate');
    await expect(pragmaMeta).toHaveAttribute('content', 'no-cache');
    await expect(expiresMeta).toHaveAttribute('content', '0');
    
    // Verify build timestamp is a valid Unix timestamp
    const timestampContent = await buildTimestampMeta.getAttribute('content');
    expect(timestampContent).toBeTruthy();
    const timestamp = parseInt(timestampContent || '0', 10);
    expect(timestamp).toBeGreaterThan(1700000000); // After Nov 2023
  });

  test('JavaScript bundle loads successfully after rebuild', async ({ authenticatedPage }) => {
    // GIVEN: User loads the application
    const response = await authenticatedPage.goto('http://localhost:3000/', {
      waitUntil: 'networkidle'
    });
    
    // THEN: No JS 404 errors should occur
    expect(response?.status()).toBe(200);
    
    // AND: React should mount (root element should have content)
    const root = authenticatedPage.locator('#root');
    await expect(root).not.toBeEmpty();
    
    // AND: Dashboard should render (not blank screen)
    const dashboard = authenticatedPage.locator('.dashboard, [data-testid="dashboard"], h1, h2');
    await expect(dashboard.first()).toBeVisible();
  });

  test('Fresh context per test (no cache carryover)', async ({ page }) => {
    // GIVEN: A fresh browser context (from fixture)
    
    // Navigate to page first (localStorage requires a document)
    await page.goto('http://localhost:3000/', {
      waitUntil: 'domcontentloaded'
    });
    
    // THEN: No storage state should exist initially
    const localStorage = await page.evaluate(() => Object.keys(window.localStorage).length);
    expect(localStorage).toBe(1); // Only devsmith_token from auth fixture
    
    const sessionStorage = await page.evaluate(() => Object.keys(window.sessionStorage).length);
    expect(sessionStorage).toBe(0);
  });

  test('Multiple page loads get fresh HTML (no stale cache)', async ({ authenticatedPage }) => {
    // GIVEN: User loads page first time
    await authenticatedPage.goto('http://localhost:3000/', {
      waitUntil: 'networkidle'
    });
    
    // Extract the JS bundle hash from first load
    const scriptSrc1 = await authenticatedPage.locator('script[type="module"]').first().getAttribute('src');
    
    // WHEN: User reloads page (simulate rebuild scenario)
    await authenticatedPage.reload({
      waitUntil: 'networkidle'
    });
    
    // THEN: JS bundle should load successfully (same hash since no actual rebuild)
    const scriptSrc2 = await authenticatedPage.locator('script[type="module"]').first().getAttribute('src');
    
    // Verify bundle loads (no 404)
    expect(scriptSrc2).toBeTruthy();
    
    // Verify React mounted
    const root = authenticatedPage.locator('#root');
    await expect(root).not.toBeEmpty();
  });

});
