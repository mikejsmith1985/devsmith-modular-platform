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

import { test, expect } from '../fixtures/auth.fixture';

test.describe('LLM Configuration Page', () => {
  test('page loads at /llm-config', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    await expect(authenticatedPage).toHaveURL('/llm-config');
    await expect(authenticatedPage.locator('h2')).toContainText('AI Model Configuration');
  });

  test('"Your AI Models" section displays empty state', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    const modelsSection = authenticatedPage.locator('text=Your AI Models').locator('..');
    await expect(modelsSection).toBeVisible();
    
    // Should show empty state message or table
    await expect(authenticatedPage.locator('text=No AI models configured')).toBeVisible();
  });

  test('"Add Model" button opens modal', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    await authenticatedPage.click('button:has-text("Add Model")');
    
    // Modal should open
    await expect(authenticatedPage.locator('[role="dialog"]')).toBeVisible();
    await expect(authenticatedPage.locator('text=Add AI Model')).toBeVisible();
  });

  test('API keys shown as "Configured" badge, not plain text', async ({ authenticatedPage }) => {
    // This test assumes we have at least one config in DB
    await authenticatedPage.goto('/llm-config');
    
    // API key column should show badge, not actual key
    const apiKeyCell = authenticatedPage.locator('td').filter({ hasText: /API Key/ }).first();
    await expect(apiKeyCell.locator('.badge')).toContainText('Configured');
    
    // Should NOT show actual key text like "sk-ant-..."
    await expect(apiKeyCell).not.toContainText('sk-');
  });

  test('default config has checkmark icon', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    // Find row with default indicator
    const defaultRow = authenticatedPage.locator('tr').filter({ has: authenticatedPage.locator('.bi-check-circle-fill') });
    await expect(defaultRow).toBeVisible();
  });

  test('edit button opens edit modal with existing values', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    // Click first edit button
    await authenticatedPage.click('button[title="Edit"]:first-child, button:has-text("Edit"):first-child');
    
    // Modal should open with pre-filled values
    await expect(authenticatedPage.locator('[role="dialog"]')).toBeVisible();
    await expect(authenticatedPage.locator('input[name="name"]')).not.toBeEmpty();
  });

  test('delete button removes config after confirmation', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    // Get initial row count
    const initialCount = await authenticatedPage.locator('tbody tr').count();
    
    // Click first delete button
    await authenticatedPage.click('button[title="Delete"]:first-child, button:has-text("Delete"):first-child');
    
    // Confirm deletion
    await authenticatedPage.click('button:has-text("Confirm"), button:has-text("Delete")');
    
    // Wait for deletion
    await authenticatedPage.waitForTimeout(500);
    
    // Row count should decrease
    const newCount = await authenticatedPage.locator('tbody tr').count();
    expect(newCount).toBe(initialCount - 1);
  });

  test('app preference dropdowns show all user configs', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    // Find app preferences section
    const prefsSection = authenticatedPage.locator('text=App-Specific Preferences').locator('..');
    await expect(prefsSection).toBeVisible();
    
    // Review app dropdown should exist
    const reviewDropdown = authenticatedPage.locator('select[name="review-preference"]');
    await expect(reviewDropdown).toBeVisible();
    
    // Should have options (at least "Use Default")
    const options = await reviewDropdown.locator('option').count();
    expect(options).toBeGreaterThan(0);
  });

  test('selecting preference updates immediately', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    // Select a different preference
    await authenticatedPage.selectOption('select[name="review-preference"]', { index: 1 });
    
    // Should show success message
    await expect(authenticatedPage.locator('.alert-success, .toast-success')).toBeVisible();
  });

  test('usage summary displays total tokens', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    
    const usageSection = authenticatedPage.locator('text=Usage Summary').locator('..');
    await expect(usageSection).toBeVisible();
    
    // Should show token count or "No usage yet"
    const hasUsage = await authenticatedPage.locator('text=/\\d+.*tokens/i').isVisible();
    const noUsage = await authenticatedPage.locator('text=No usage yet').isVisible();
    
    expect(hasUsage || noUsage).toBeTruthy();
  });
});

test.describe('Add LLM Config Modal', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/llm-config');
    await authenticatedPage.click('button:has-text("Add Model")');
  });

  test('provider dropdown lists all providers', async ({ authenticatedPage }) => {
    const providerSelect = authenticatedPage.locator('select[name="provider"]');
    await expect(providerSelect).toBeVisible();
    
    // Should have options for all supported providers
    await expect(providerSelect.locator('option[value="anthropic"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="openai"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="ollama"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="deepseek"]')).toBeVisible();
    await expect(providerSelect.locator('option[value="mistral"]')).toBeVisible();
  });

  test('model dropdown updates based on provider', async ({ authenticatedPage }) => {
    // Select Anthropic
    await authenticatedPage.selectOption('select[name="provider"]', 'anthropic');
    
    // Model dropdown should show Claude models
    const modelSelect = authenticatedPage.locator('select[name="model"]');
    await expect(modelSelect.locator('option')).toContainText(/claude/i);
    
    // Select OpenAI
    await authenticatedPage.selectOption('select[name="provider"]', 'openai');
    
    // Model dropdown should show GPT models
    await expect(modelSelect.locator('option')).toContainText(/gpt/i);
  });

  test('API key field is password type', async ({ authenticatedPage }) => {
    const apiKeyInput = authenticatedPage.locator('input[name="apiKey"]');
    await expect(apiKeyInput).toHaveAttribute('type', 'password');
  });

  test('API key field hidden for Ollama (local)', async ({ authenticatedPage }) => {
    await authenticatedPage.selectOption('select[name="provider"]', 'ollama');
    
    // API key field should be hidden or disabled
    const apiKeyInput = authenticatedPage.locator('input[name="apiKey"]');
    const isHidden = await apiKeyInput.isHidden();
    const isDisabled = await apiKeyInput.isDisabled();
    
    expect(isHidden || isDisabled).toBeTruthy();
  });

  test('test connection button pings provider', async ({ authenticatedPage }) => {
    await authenticatedPage.selectOption('select[name="provider"]', 'ollama');
    await authenticatedPage.fill('input[name="name"]', 'Test Ollama');
    await authenticatedPage.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    await authenticatedPage.click('button:has-text("Test Connection")');
    
    // Should show loading state then result
    await expect(authenticatedPage.locator('text=/Testing.../i, text=/Connecting.../i')).toBeVisible();
    
    // Wait for result
    await authenticatedPage.waitForTimeout(2000);
    
    // Should show success or failure
    const hasSuccess = await authenticatedPage.locator('text=/Success/i, text=/Connected/i').isVisible();
    const hasFailure = await authenticatedPage.locator('text=/Failed/i, text=/Error/i').isVisible();
    
    expect(hasSuccess || hasFailure).toBeTruthy();
  });

  test('save button disabled until valid config', async ({ authenticatedPage }) => {
    const saveButton = authenticatedPage.locator('button:has-text("Save")');
    
    // Initially disabled (empty form)
    await expect(saveButton).toBeDisabled();
    
    // Fill required fields
    await authenticatedPage.selectOption('select[name="provider"]', 'ollama');
    await authenticatedPage.fill('input[name="name"]', 'My Ollama');
    await authenticatedPage.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    // Should now be enabled
    await expect(saveButton).toBeEnabled();
  });

  test('save button creates config and closes modal', async ({ authenticatedPage }) => {
    // Fill form
    await authenticatedPage.selectOption('select[name="provider"]', 'ollama');
    await authenticatedPage.fill('input[name="name"]', 'Test Config');
    await authenticatedPage.fill('input[name="endpoint"]', 'http://localhost:11434');
    
    // Save
    await authenticatedPage.click('button:has-text("Save")');
    
    // Modal should close
    await expect(authenticatedPage.locator('[role="dialog"]')).not.toBeVisible();
    
    // Should show success message
    await expect(authenticatedPage.locator('.alert-success, .toast-success')).toBeVisible();
  });

  test('newly created config appears in table', async ({ authenticatedPage }) => {
    const configName = `Test-${Date.now()}`;
    
    // Create config
    await authenticatedPage.selectOption('select[name="provider"]', 'ollama');
    await authenticatedPage.fill('input[name="name"]', configName);
    await authenticatedPage.fill('input[name="endpoint"]', 'http://localhost:11434');
    await authenticatedPage.click('button:has-text("Save")');
    
    // Wait for modal to close
    await authenticatedPage.waitForSelector('[role="dialog"]', { state: 'hidden' });
    
    // Config should appear in table
    await expect(authenticatedPage.locator(`text=${configName}`)).toBeVisible();
  });
});
