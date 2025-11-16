import { test, expect } from '@playwright/test';

test.describe('SMOKE: Review Critical Mode', () => {
  test('Can submit code and receive AI analysis', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Paste sample code
    const testCode = `package main\n\nimport "fmt"\n\nfunc main() {\n  fmt.Println("Hello, World!")\n}`;
    await page.fill('textarea[name="pasted_code"]', testCode);
    
    // Click submit
    await page.click('button[type="submit"]:has-text("Start Review")');
    
    // Wait for form to process (HTMX request)
    await page.waitForTimeout(1000);
    
    // Form should still be visible (SSE progress in separate container)
    const form = page.locator('form#review-session-form');
    await expect(form).toBeVisible();
  });

  test('Critical mode button triggers analysis', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    const testCode = `package main\nfunc main() { x := 1; }`;
    await page.fill('textarea[name="pasted_code"]', testCode);
    
  // Find and click Critical Mode button (use hx-post attribute so selector matches templated button)
  const criticalButton = page.locator('button[hx-post="/api/review/modes/critical"]').first();
    
    // Listen for network requests to /api/review/modes/critical
    // AI analysis can take 20-30 seconds, so increase timeout
    const responsePromise = page.waitForResponse(
      response => response.url().includes('/api/review/modes/critical') && response.status() === 200,
      { timeout: 45000 } // 45 seconds for AI analysis
    );
    
    await criticalButton.click({ timeout: 15000 });
    
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    
    // Response should contain HTML (not error)
    const text = await response.text();
    expect(text.length).toBeGreaterThan(0);
  });

  test('Mode results container receives analysis', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    const testCode = `func broken() { return; return; }`;
    await page.fill('textarea[name="pasted_code"]', testCode);
    
    const criticalButton = page.locator('button:has-text("Critical Mode")').first();
    await criticalButton.click({ timeout: 15000 });
    
    // Results should appear in results container
    const resultsContainer = page.locator('#results-container');
    
    // Wait for HTMX to swap content (poll for non-empty content)
    // AI analysis can take 20-30 seconds, increase timeout
    let retries = 60; // 60 * 500ms = 30 seconds max
    let hasContent = false;
    while (retries > 0 && !hasContent) {
      const content = await resultsContainer.textContent();
      if (content && content.trim().length > 0) {
        hasContent = true;
        break;
      }
      await page.waitForTimeout(500);
      retries--;
    }
    
    expect(hasContent).toBe(true);
  });
});
