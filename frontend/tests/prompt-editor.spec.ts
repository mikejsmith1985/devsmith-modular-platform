import { test, expect } from '@playwright/test';

/**
 * Phase 4, Task 4.1: Prompt Editor Modal Component - E2E Tests
 * 
 * TDD RED Phase: These tests define expected behavior before implementation
 * 
 * Test Coverage:
 * - Modal open/close functionality
 * - Display current prompt (default or custom)
 * - Custom prompt badge visibility
 * - Variable reference panel
 * - Character counter
 * - Save custom prompt functionality
 * - Factory reset functionality
 * - Persistence after page refresh
 */

test.describe('Prompt Editor Modal', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to Review page
    await page.goto('/review');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
  });

  test('should display Details button on mode cards', async ({ page }) => {
    // Each mode card should have a Details button
    const previewDetailsBtn = page.locator('.mode-card.preview .btn-details');
    const skimDetailsBtn = page.locator('.mode-card.skim .btn-details');
    const scanDetailsBtn = page.locator('.mode-card.scan .btn-details');
    const detailedDetailsBtn = page.locator('.mode-card.detailed .btn-details');
    const criticalDetailsBtn = page.locator('.mode-card.critical .btn-details');
    
    await expect(previewDetailsBtn).toBeVisible();
    await expect(skimDetailsBtn).toBeVisible();
    await expect(scanDetailsBtn).toBeVisible();
    await expect(detailedDetailsBtn).toBeVisible();
    await expect(criticalDetailsBtn).toBeVisible();
  });

  test('should open modal when clicking Details button', async ({ page }) => {
    // Click Details button on Preview mode
    await page.click('.mode-card.preview .btn-details');
    
    // Modal should be visible
    const modal = page.locator('.prompt-editor-modal');
    await expect(modal).toBeVisible();
    
    // Modal should have title
    await expect(modal.locator('.modal-title')).toContainText('Edit Prompt');
  });

  test('should display current prompt text in modal', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    
    // Wait for modal to load
    await page.waitForSelector('.prompt-editor-modal');
    
    // Prompt textarea should contain text
    const promptTextarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await expect(promptTextarea).not.toBeEmpty();
    
    // Should contain common prompt variables
    const promptText = await promptTextarea.inputValue();
    expect(promptText).toContain('{{code}}');
  });

  test('should show system default badge initially', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    
    // Wait for modal
    await page.waitForSelector('.prompt-editor-modal');
    
    // Should show "System Default" badge
    const badge = page.locator('.prompt-editor-modal .badge-default');
    await expect(badge).toBeVisible();
    await expect(badge).toContainText('System Default');
    
    // Should NOT show "Custom" badge initially
    const customBadge = page.locator('.prompt-editor-modal .badge-custom');
    await expect(customBadge).not.toBeVisible();
  });

  test('should display variable reference panel', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    
    // Wait for modal
    await page.waitForSelector('.prompt-editor-modal');
    
    // Variable reference panel should be visible
    const varPanel = page.locator('.variable-reference-panel');
    await expect(varPanel).toBeVisible();
    
    // Should list available variables
    await expect(varPanel).toContainText('{{code}}');
    await expect(varPanel).toContainText('Code to analyze');
  });

  test('should update character count when editing prompt', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    
    // Wait for modal
    await page.waitForSelector('.prompt-editor-modal');
    
    // Get initial character count
    const charCount = page.locator('.character-count');
    const initialCount = await charCount.textContent();
    
    // Edit prompt
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('Short prompt with {{code}}');
    
    // Character count should update
    const newCount = await charCount.textContent();
    expect(newCount).not.toBe(initialCount);
    expect(newCount).toContain('28'); // Length of "Short prompt with {{code}}"
  });

  test('should save custom prompt and show Custom badge', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    
    // Wait for modal
    await page.waitForSelector('.prompt-editor-modal');
    
    // Edit prompt
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('My custom prompt for {{code}} analysis');
    
    // Click Save button
    await page.click('.prompt-editor-modal .btn-save');
    
    // Wait for save to complete
    await page.waitForTimeout(500);
    
    // Modal should close
    await expect(page.locator('.prompt-editor-modal')).not.toBeVisible();
    
    // Re-open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Should now show Custom badge
    const customBadge = page.locator('.prompt-editor-modal .badge-custom');
    await expect(customBadge).toBeVisible();
    await expect(customBadge).toContainText('Custom');
    
    // Should NOT show System Default badge
    const defaultBadge = page.locator('.prompt-editor-modal .badge-default');
    await expect(defaultBadge).not.toBeVisible();
    
    // Prompt should be saved
    const savedPrompt = await textarea.inputValue();
    expect(savedPrompt).toBe('My custom prompt for {{code}} analysis');
  });

  test('should show Factory Reset button after saving custom prompt', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Initially, Factory Reset button should NOT be visible
    const resetBtn = page.locator('.prompt-editor-modal .btn-factory-reset');
    await expect(resetBtn).not.toBeVisible();
    
    // Edit and save custom prompt
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('Custom prompt with {{code}}');
    await page.click('.prompt-editor-modal .btn-save');
    await page.waitForTimeout(500);
    
    // Re-open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Factory Reset button should NOW be visible
    await expect(resetBtn).toBeVisible();
  });

  test('should reset to system default when clicking Factory Reset', async ({ page }) => {
    // First, create a custom prompt
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    
    // Save original default prompt
    const originalPrompt = await textarea.inputValue();
    
    // Save custom prompt
    await textarea.fill('Custom prompt with {{code}}');
    await page.click('.prompt-editor-modal .btn-save');
    await page.waitForTimeout(500);
    
    // Re-open and click Factory Reset
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    await page.click('.prompt-editor-modal .btn-factory-reset');
    
    // Confirm reset
    await page.click('.confirm-reset-btn'); // Confirmation modal button
    await page.waitForTimeout(500);
    
    // Modal should reload with system default
    const resetPrompt = await textarea.inputValue();
    expect(resetPrompt).toBe(originalPrompt);
    
    // Should show System Default badge again
    const defaultBadge = page.locator('.prompt-editor-modal .badge-default');
    await expect(defaultBadge).toBeVisible();
    
    // Factory Reset button should be hidden
    const resetBtn = page.locator('.prompt-editor-modal .btn-factory-reset');
    await expect(resetBtn).not.toBeVisible();
  });

  test('should close modal without saving when clicking Cancel', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    const originalPrompt = await textarea.inputValue();
    
    // Edit prompt
    await textarea.fill('Unsaved changes');
    
    // Click Cancel
    await page.click('.prompt-editor-modal .btn-cancel');
    
    // Modal should close
    await expect(page.locator('.prompt-editor-modal')).not.toBeVisible();
    
    // Re-open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Prompt should be unchanged
    const currentPrompt = await textarea.inputValue();
    expect(currentPrompt).toBe(originalPrompt);
  });

  test('should persist custom prompt after page refresh', async ({ page }) => {
    // Save custom prompt
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('Persistent custom prompt with {{code}}');
    await page.click('.prompt-editor-modal .btn-save');
    await page.waitForTimeout(500);
    
    // Refresh page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // Re-open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Custom prompt should still be there
    const savedPrompt = await textarea.inputValue();
    expect(savedPrompt).toBe('Persistent custom prompt with {{code}}');
    
    // Should show Custom badge
    const customBadge = page.locator('.prompt-editor-modal .badge-custom');
    await expect(customBadge).toBeVisible();
  });

  test('should load different prompts for different modes', async ({ page }) => {
    // Get prompt for Preview mode
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    const previewPrompt = await page.locator('.prompt-editor-modal textarea[name="prompt"]').inputValue();
    await page.click('.prompt-editor-modal .btn-cancel');
    
    // Get prompt for Skim mode
    await page.click('.mode-card.skim .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    const skimPrompt = await page.locator('.prompt-editor-modal textarea[name="prompt"]').inputValue();
    await page.click('.prompt-editor-modal .btn-cancel');
    
    // Prompts should be different
    expect(previewPrompt).not.toBe(skimPrompt);
  });

  test('should validate required variables are present', async ({ page }) => {
    // Click Details button
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Try to save prompt without required {{code}} variable
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('Invalid prompt without code variable');
    
    // Click Save
    await page.click('.prompt-editor-modal .btn-save');
    
    // Should show validation error
    const errorMsg = page.locator('.prompt-editor-modal .validation-error');
    await expect(errorMsg).toBeVisible();
    await expect(errorMsg).toContainText('{{code}}');
    await expect(errorMsg).toContainText('required');
    
    // Modal should remain open
    await expect(page.locator('.prompt-editor-modal')).toBeVisible();
  });

  test('should validate scan mode requires {{query}} variable', async ({ page }) => {
    // Click Details for Scan mode
    await page.click('.mode-card.scan .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Try to save without {{query}}
    const textarea = page.locator('.prompt-editor-modal textarea[name="prompt"]');
    await textarea.fill('Scan prompt with {{code}} but no query');
    
    // Click Save
    await page.click('.prompt-editor-modal .btn-save');
    
    // Should show validation error for missing {{query}}
    const errorMsg = page.locator('.prompt-editor-modal .validation-error');
    await expect(errorMsg).toBeVisible();
    await expect(errorMsg).toContainText('{{query}}');
  });
});

test.describe('Prompt Editor Visual Tests', () => {
  test('should match visual snapshot - default state', async ({ page }) => {
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Take snapshot
    await expect(page.locator('.prompt-editor-modal')).toHaveScreenshot('prompt-editor-default.png');
  });

  test('should match visual snapshot - custom prompt', async ({ page }) => {
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Save custom prompt
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    await page.fill('.prompt-editor-modal textarea[name="prompt"]', 'Custom prompt with {{code}}');
    await page.click('.prompt-editor-modal .btn-save');
    await page.waitForTimeout(500);
    
    // Re-open and take snapshot
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    await expect(page.locator('.prompt-editor-modal')).toHaveScreenshot('prompt-editor-custom.png');
  });

  test('should match visual snapshot - variable reference expanded', async ({ page }) => {
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Expand variable reference panel
    await page.click('.variable-reference-panel .expand-btn');
    
    // Take snapshot
    await expect(page.locator('.prompt-editor-modal')).toHaveScreenshot('prompt-editor-variables-expanded.png');
  });

  test('should match visual snapshot - long prompt with scroll', async ({ page }) => {
    await page.goto('/review');
    await page.waitForLoadState('networkidle');
    
    // Open modal
    await page.click('.mode-card.preview .btn-details');
    await page.waitForSelector('.prompt-editor-modal');
    
    // Fill with very long prompt
    const longPrompt = 'Long prompt text. '.repeat(100) + '{{code}}';
    await page.fill('.prompt-editor-modal textarea[name="prompt"]', longPrompt);
    
    // Take snapshot showing scroll
    await expect(page.locator('.prompt-editor-modal')).toHaveScreenshot('prompt-editor-long-prompt.png');
  });
});
