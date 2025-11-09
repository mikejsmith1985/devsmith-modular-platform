import { test, expect } from '@playwright/test';

test.describe('Issue #2: Prompt Details Button Fix', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate first (use your existing auth fixture or OAuth flow)
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL('**/review**', { timeout: 30000 });
  });

  test('Quick Learn mode should load prompt successfully', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    
    // Ensure Quick Learn is selected (default)
    await expect(page.locator('input[value="quick"]')).toBeChecked();
    
    // Click Details button
    await page.click('button:has-text("Details")');
    
    // Wait for modal to open
    await page.waitForSelector('.modal:visible', { timeout: 5000 });
    
    // Check for success - NO error message should appear
    await expect(page.locator('text=/HTTP 500.*Failed to retrieve prompt/i')).not.toBeVisible();
    
    // Verify prompt loaded (should have "Prompt Template" heading)
    await expect(page.locator('text=Prompt Template')).toBeVisible();
    
    // Verify textarea is not empty
    const textarea = page.locator('textarea[placeholder*="prompt template"]');
    const content = await textarea.inputValue();
    expect(content.length).toBeGreaterThan(0);
  });

  test('Full Learn (detailed) mode should load prompt successfully', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    
    // Select Full Learn mode
    await page.click('input[value="detailed"]');
    await expect(page.locator('input[value="detailed"]')).toBeChecked();
    
    // Click Details button
    await page.click('button:has-text("Details")');
    
    // Wait for modal to open
    await page.waitForSelector('.modal:visible', { timeout: 5000 });
    
    // Check for success - NO error message should appear
    await expect(page.locator('text=/HTTP 500.*Failed to retrieve prompt/i')).not.toBeVisible();
    
    // Verify prompt loaded
    await expect(page.locator('text=Prompt Template')).toBeVisible();
    
    // Verify textarea has content
    const textarea = page.locator('textarea[placeholder*="prompt template"]');
    const content = await textarea.inputValue();
    expect(content.length).toBeGreaterThan(0);
    
    // Verify it contains reasoning/analysis keywords (specific to 'detailed' prompts)
    expect(content).toMatch(/step|analysis|process|reasoning/i);
  });

  test('Scan mode with detailed output should load prompt successfully', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    
    // Select Scan mode from dropdown
    await page.selectOption('select[name="mode"]', 'scan');
    
    // Select Full Learn (detailed) output mode
    await page.click('input[value="detailed"]');
    
    // Click Details button
    await page.click('button:has-text("Details")');
    
    // Wait for modal
    await page.waitForSelector('.modal:visible', { timeout: 5000 });
    
    // Should NOT show error
    await expect(page.locator('text=/HTTP 500.*Failed to retrieve prompt/i')).not.toBeVisible();
    
    // Should show Scan mode prompt
    await expect(page.locator('text=Prompt Template')).toBeVisible();
  });
});
