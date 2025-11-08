/**
 * LLM Configuration Page E2E Tests
 * 
 * Phase 5, Task 5.2 - TDD RED Phase
 * 
 * Tests the LLM configuration management UI:
 * - Display user's AI model configurations
 * - Add new LLM config
 * - Edit existing config
 * - Delete config
 * - Set app-specific preferences
 * - View usage summary
 */

import { test, expect } from '@playwright/test';

test.describe('LLM Configuration Page', () => {
  test.beforeEach(async ({ page }) => {
    // Login first (reuse existing auth fixture if available)
    await page.goto('/login');
    // TODO: Add proper login flow when auth is working
    // For now, might need to mock authentication
  });

  test('page loads at /llm-config', async ({ page }) => {
    await page.goto('/llm-config');
    
    await expect(page).toHaveURL('/llm-config');
    await expect(page.locator('h2')).toContainText('AI Model Configuration');
  });

  test('"Your AI Models" section displays empty state', async ({ page }) => {
    await page.goto('/llm-config');
    
    const modelsSection = page.locator('text=Your AI Models').locator('..');
    await expect(modelsSection).toBeVisible();
    
    // Should show empty state message or table
    await expect(page.locator('text=No AI models configured')).toBeVisible();
  });

  test('"Add Model" button opens modal', async ({ page }) => {
    await page.goto('/llm-config');
    
    await page.click('button:has-text("Add Model")');
    
    // Modal should open
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await expect(page.locator('text=Add AI Model')).toBeVisible();
  });

  test('API keys shown as "Configured" badge, not plain text', async ({ page }) => {
    // This test assumes we have at least one config in DB
    await page.goto('/llm-config');
    
    // API key column should show badge, not actual key
    const apiKeyCell = page.locator('td').filter({ hasText: /API Key/ }).first();
    await expect(apiKeyCell.locator('.badge')).toContainText('Configured');
    
    // Should NOT show actual key text like "sk-ant-..."
    await expect(apiKeyCell).not.toContainText('sk-');
  });

  test('default config has checkmark icon', async ({ page }) => {
    await page.goto('/llm-config');
    
    // Find row with default indicator
    const defaultRow = page.locator('tr').filter({ has: page.locator('.bi-check-circle-fill') });
    await expect(defaultRow).toBeVisible();
  });

  test('edit button opens edit modal with existing values', async ({ page }) => {
    await page.goto('/llm-config');
    
    // Click first edit button
    await page.click('button[title="Edit"]:first-child, button:has-text("Edit"):first-child');
    
    // Modal should open with pre-filled values
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    await expect(page.locator('input[name="name"]')).not.toBeEmpty();
  });

  test('delete button removes config after confirmation', async ({ page }) => {
    await page.goto('/llm-config');
    
    // Get initial row count
    const initialCount = await page.locator('tbody tr').count();
    
    // Click first delete button
    await page.click('button[title="Delete"]:first-child, button:has-text("Delete"):first-child');
    
    // Confirm deletion
    await page.click('button:has-text("Confirm"), button:has-text("Delete")');
    
    // Wait for deletion
    await page.waitForTimeout(500);
    
    // Row count should decrease
    const newCount = await page.locator('tbody tr').count();
    expect(newCount).toBe(initialCount - 1);
  });

  test('app preference dropdowns show all user configs', async ({ page }) => {
    await page.goto('/llm-config');
    
    // Find app preferences section
    const prefsSection = page.locator('text=App-Specific Preferences').locator('..');
    await expect(prefsSection).toBeVisible();
    
    // Review app dropdown should exist
    const reviewDropdown = page.locator('select[name="review-preference"]');
    await expect(reviewDropdown).toBeVisible();
    
    // Should have options (at least "Use Default")
    const options = await reviewDropdown.locator('option').count();
    expect(options).toBeGreaterThan(0);
  });

  test('selecting preference updates immediately', async ({ page }) => {
    await page.goto('/llm-config');
    
    // Select a different preference
    await page.selectOption('select[name="review-preference"]', { index: 1 });
    
    // Should show success message
    await expect(page.locator('.alert-success, .toast-success')).toBeVisible();
  });

  test('usage summary displays total tokens', async ({ page }) => {
    await page.goto('/llm-config');
    
    const usageSection = page.locator('text=Usage Summary').locator('..');
    await expect(usageSection).toBeVisible();
    
    // Should show token count or "No usage yet"
    const hasUsage = await page.locator('text=/\\d+.*tokens/i').isVisible();
    const noUsage = await page.locator('text=No usage yet').isVisible();
    
    expect(hasUsage || noUsage).toBeTruthy();
  });
});

test.describe('Add LLM Config Modal', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/llm-config');
    await page.click('button:has-text("Add Model")');
  });

  test('provider dropdown lists all providers', async ({ page }) => {
    const providerSelect = page.locator('select[name="provider"]');
    await expect(providerSelect).toBeVisible();
    
    // Should have options for all supported providers
    await expect(providerSelect.locator('option[value="anthropic"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="openai"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="ollama"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="deepseek"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="mistral"]')).toBeVisible();
  });

  test('model dropdown updates based on provider', async ({ page }) => {
    // Select Anthropic
    await page.selectOption('select[name="provider"]', 'anthropic');
    
    // Model dropdown should show Claude models
    const modelSelect = page.locator('select[name="model"]');
    await expect(modelSelect.locator('option')).toContainText(/claude/i);
    
    // Select OpenAI
    await page.selectOption('select[name="provider"]', 'openai');
    
    // Model dropdown should show GPT models
    await expect(modelSelect.locator('option')).toContainText(/gpt/i);
  });

  test('API key field is password type', async ({ page }) => {
    const apiKeyInput = page.locator('input[name="apiKey"]');
    await expect(apiKeyInput).toHaveAttribute('type', 'password');
  });

  test('API key field hidden for Ollama (local)', async ({ page }) => {
    await page.selectOption('select[name="provider"]', 'ollama');
    
    // API key field should be hidden or disabled
    const apiKeyInput = page.locator('input[name="apiKey"]');
    const isHidden = await apiKeyInput.isHidden();
    const isDisabled = await apiKeyInput.isDisabled();
    
    expect(isHidden || isDisabled).toBeTruthy();
  });

  test('test connection button pings provider', async ({ page }) => {
    await page.selectOption('select[name="provider"]', 'ollama');
    await page.fill('input[name="name"]', 'Test Ollama');
    await page.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    await page.click('button:has-text("Test Connection")');
    
    // Should show loading state then result
    await expect(page.locator('text=/Testing.../i, text=/Connecting.../i')).toBeVisible();
    
    // Wait for result
    await page.waitForTimeout(2000);
    
    // Should show success or failure
    const hasSuccess = await page.locator('text=/Success/i, text=/Connected/i').isVisible();
    const hasFailure = await page.locator('text=/Failed/i, text=/Error/i').isVisible();
    
    expect(hasSuccess || hasFailure).toBeTruthy();
  });

  test('save button disabled until valid config', async ({ page }) => {
    const saveButton = page.locator('button:has-text("Save")');
    
    // Initially disabled (empty form)
    await expect(saveButton).toBeDisabled();
    
    // Fill required fields
    await page.selectOption('select[name="provider"]', 'ollama');
    await page.fill('input[name="name"]', 'My Ollama');
    await page.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    // Should now be enabled
    await expect(saveButton).toBeEnabled();
  });

  test('save button creates config and closes modal', async ({ page }) => {
    // Fill form
    await page.selectOption('select[name="provider"]', 'ollama');
    await page.fill('input[name="name"]', 'Test Config');
    await page.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    // Save
    await page.click('button:has-text("Save")');
    
    // Modal should close
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    
    // Should show success message
    await expect(page.locator('.alert-success, .toast-success')).toBeVisible();
  });

  test('newly created config appears in table', async ({ page }) => {
    const configName = `Test-${Date.now()}`;
    
    // Create config
    await page.selectOption('select[name="provider"]', 'ollama');
    await page.fill('input[name="name"]', configName);
    await page.fill('input[name="endpoint"]', 'http://localhost:11434');
    await page.click('button:has-text("Save")');
    
    // Wait for modal to close
    await page.waitForSelector('[role="dialog"]', { state: 'hidden' });
    
    // Config should appear in table
    await expect(page.locator(`text=${configName}`)).toBeVisible();
  });
});
