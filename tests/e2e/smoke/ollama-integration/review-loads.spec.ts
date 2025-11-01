import { test, expect } from '@playwright/test';

test.describe('SMOKE: Review Loads', () => {
  test('Review page is accessible', async ({ page }) => {
    const response = await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    expect(response?.status()).toBe(200);
  });

  test('Session creation form renders with all input methods', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Check for form
    const form = page.locator('form#review-session-form');
    await expect(form).toBeVisible();
    
    // Check for paste input
    await expect(page.locator('textarea[name="pasted_code"]')).toBeVisible();
    
    // Check for GitHub URL input
    await expect(page.locator('input[name="github_url"]')).toBeVisible();
    
    // Check for file upload
    await expect(page.locator('input[name="file"]')).toBeVisible();
  });

  test('Reading mode cards are visible and clickable', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Check for mode buttons with correct text (Select Preview, Select Skim, etc.)
    const previewButton = page.locator('button:has-text("Select Preview")');
    const skimButton = page.locator('button:has-text("Select Skim")');
    const scanButton = page.locator('button:has-text("Select Scan")');
    const detailedButton = page.locator('button:has-text("Select Detailed")');
    const criticalButton = page.locator('button:has-text("Select Critical")');
    
    await expect(previewButton).toBeVisible();
    await expect(skimButton).toBeVisible();
    await expect(scanButton).toBeVisible();
    await expect(detailedButton).toBeVisible();
    await expect(criticalButton).toBeVisible();
  });

  test('Submit button is present and enabled', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    const submitButton = page.locator('button[type="submit"]').filter({ hasText: 'Start Review' });
    await expect(submitButton).toBeVisible();
    await expect(submitButton).toBeEnabled();
  });
});
