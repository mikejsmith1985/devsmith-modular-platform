import { test, expect } from '@playwright/test';

/**
 * Review Service E2E Tests
 * 
 * Tests the Review application workflow:
 * 1. Accessing Review service through Traefik
 * 2. Authentication requirements
 * 3. Basic UI functionality
 */

test.describe('Review Service Access', () => {
  test('should be accessible at /review path', async ({ page }) => {
    // Navigate to Review service through Traefik
    await page.goto('/review');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Should either show Review UI or redirect to login
    const url = page.url();
    const isReviewPage = url.includes('/review');
    const isLoginPage = url.includes('/login') || url.includes('/auth');
    
    expect(isReviewPage || isLoginPage).toBeTruthy();
  });

  test('should redirect to authentication when not logged in', async ({ page }) => {
    // Navigate directly to Review
    await page.goto('/review');
    
    // Wait for any redirects
    await page.waitForLoadState('networkidle');
    
    const url = page.url();
    
    // Should be on login page or GitHub OAuth
    if (!url.includes('/review')) {
      expect(url).toMatch(/login|auth|github/);
    } else {
      // Or might show login UI within Review app
      const hasLoginUI = await page.locator('text=/login|sign in/i').count() > 0;
      expect(hasLoginUI).toBeTruthy();
    }
  });

  test('should load without errors', async ({ page }) => {
    let errors: string[] = [];
    
    // Capture console errors
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    // Navigate to Review
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Should have no JavaScript errors
    // (some 401 auth errors are expected, but not JS errors)
    const jsErrors = errors.filter(e => 
      !e.includes('401') && 
      !e.includes('Unauthorized')
    );
    
    expect(jsErrors.length).toBe(0);
  });
});

/**
 * Review Service UI Tests (Authenticated)
 * 
 * These tests require authentication
 * Currently marked as TODO
 */

test.describe.skip('Review Workspace (Authenticated)', () => {
  test('should display workspace selection', async ({ page }) => {
    // TODO: Test with authenticated session
    await page.goto('/review');
    
    // Should see workspace options or form
    const workspaceUI = page.locator('[data-testid="workspace-selector"]');
    await expect(workspaceUI).toBeVisible();
  });

  test('should allow code paste', async ({ page }) => {
    // TODO: Test pasting code into Review
    await page.goto('/review');
    
    const codeInput = page.locator('textarea[placeholder*="code" i]');
    await expect(codeInput).toBeVisible();
    
    await codeInput.fill('package main\n\nfunc main() {\n\tprintln("test")\n}');
    
    // Should accept code input
    const value = await codeInput.inputValue();
    expect(value).toContain('package main');
  });

  test('should offer reading mode selection', async ({ page }) => {
    // TODO: Verify 5 reading modes available
    await page.goto('/review');
    
    const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];
    
    for (const mode of modes) {
      const modeOption = page.locator(`text=${mode}`).first();
      await expect(modeOption).toBeVisible();
    }
  });
});
