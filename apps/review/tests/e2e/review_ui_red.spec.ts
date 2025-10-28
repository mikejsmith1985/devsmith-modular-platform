// RED PHASE: Failing Playwright E2E tests for Review UI/UX (Feature 24)
// These tests will fail until the UI is implemented.
import { test, expect } from '@playwright/test';

test('Session creation form validates input', async ({ page }) => {
  await page.goto('/review');
  await expect(page.locator('#session-form')).toBeVisible();
  await page.click('button[type=submit]');
  await expect(page.locator('.error')).toBeVisible();
});

test('Code input supports paste/upload/GitHub', async ({ page }) => {
  await page.goto('/review');
  await expect(page.locator('#code-input')).toBeVisible();
  await expect(page.locator('#github-url')).toBeVisible();
  await expect(page.locator('#file-upload')).toBeVisible();
});

test('Preview mode results display', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-preview');
  await expect(page.locator('.preview-results')).toBeVisible();
});

test('Skim mode function list with expand', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-skim');
  await expect(page.locator('.function-list')).toBeVisible();
  await page.click('.function-list .expand');
  await expect(page.locator('.function-details')).toBeVisible();
});

test('Scan mode search interface', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-scan');
  await expect(page.locator('#scan-search')).toBeVisible();
});

test('Detailed mode line-by-line view', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-detailed');
  await expect(page.locator('.line-by-line')).toBeVisible();
});

test('Critical mode issue list with severity badges', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-critical');
  await expect(page.locator('.issue-list')).toBeVisible();
  await expect(page.locator('.severity-badge')).toBeVisible();
});

test('Mode transitions work fluidly', async ({ page }) => {
  await page.goto('/review');
  await page.click('#mode-preview');
  await page.click('#go-to-skim');
  await expect(page.locator('.function-list')).toBeVisible();
  await page.click('#go-to-detailed');
  await expect(page.locator('.line-by-line')).toBeVisible();
  await page.click('#go-to-critical');
  await expect(page.locator('.issue-list')).toBeVisible();
});

test('Mobile responsive', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 });
  await page.goto('/review');
  await expect(page.locator('#session-form')).toBeVisible();
});

test('Accessibility (WCAG 2.1 AA)', async ({ page }) => {
  await page.goto('/review');
  // Placeholder: real test would use axe-core or similar
  await expect(page.locator('body')).toHaveAttribute('data-wcag-aa', 'true');
});
