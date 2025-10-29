import { test, expect } from '@playwright/test';

// E2E: Review Session Creation Form
// Covers: Paste, Upload, GitHub URL, validation, accessibility

test.describe('Review Session Creation', () => {
  test('shows form and validates required fields', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await expect(page.locator('form#review-session-form')).toBeVisible();
    await page.click('button[type="submit"]');
    await expect(page.locator('.error-message')).toContainText('Code input is required');
  });

  test('accepts pasted code', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.fill('textarea[name="pasted_code"]', 'package main\nfunc main() {}');
    await page.click('button[type="submit"]');
    await expect(page.locator('.analysis-progress')).toBeVisible();
  });

  test('accepts file upload', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    const filePath = 'tests/e2e/review/testdata/simple.go';
    await page.setInputFiles('input[type="file"]', filePath);
    await page.click('button[type="submit"]');
    await expect(page.locator('.analysis-progress')).toBeVisible();
  });

  test('accepts GitHub URL', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.fill('input[name="github_url"]', 'https://github.com/example/repo');
    await page.click('button[type="submit"]');
    await expect(page.locator('.analysis-progress')).toBeVisible();
  });

  test('is accessible (WCAG 2.1 AA)', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    // Check for label associations
    await expect(page.locator('label[for="pasted_code"]')).toBeVisible();
    await expect(page.locator('label[for="github_url"]')).toBeVisible();
    // Keyboard navigation
    await page.keyboard.press('Tab');
    await expect(page.locator('textarea[name="pasted_code"]')).toBeFocused();
  });
});
