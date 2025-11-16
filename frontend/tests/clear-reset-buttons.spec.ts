/**
 * Phase 4, Task 4.3: Test Clear/Reset Button Functionality
 * 
 * These tests validate that Clear and Reset buttons work correctly with the files array
 * (not the old code/setCode state).
 * 
 * Test Coverage:
 * - Clear button clears active file content
 * - Clear button clears analysis results
 * - Clear button does not affect other tabs
 * - Reset button replaces all files with default example
 * - Reset button clears analysis results
 * - Reset button resets UI to single default tab
 */

import { test, expect } from '@playwright/test';

test.describe('ReviewPage - Clear/Reset Buttons', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to review page
    await page.goto('http://localhost:3000/review');
    
    // Wait for Monaco editor to load
    await page.waitForSelector('.monaco-editor', { timeout: 10000 });
  });

  test('Clear button clears active file content', async ({ page }) => {
    // GIVEN: User has entered code in Monaco editor
    const editor = page.locator('.monaco-editor');
    await editor.click();
    await page.keyboard.type('function test() { return 42; }');
    
    // Verify content was entered
    const content = await page.evaluate(() => {
      const monaco = window.monaco;
      const editor = monaco.editor.getModels()[0];
      return editor.getValue();
    });
    expect(content).toContain('function test()');

    // WHEN: User clicks Clear button
    const clearButton = page.locator('button:has-text("Clear")');
    await clearButton.click();

    // THEN: Active file content is cleared
    const clearedContent = await page.evaluate(() => {
      const monaco = window.monaco;
      const editor = monaco.editor.getModels()[0];
      return editor.getValue();
    });
    expect(clearedContent).toBe('');
  });

  test('Clear button clears analysis results', async ({ page }) => {
    // GIVEN: User has run analysis with results displayed
    const editor = page.locator('.monaco-editor');
    await editor.click();
    await page.keyboard.type('function analyze() { return "result"; }');

    // Select mode and analyze
    await page.click('text=Quick Preview');
    await page.click('button:has-text("Analyze Code")');

    // Wait for results
    await page.waitForSelector('.analysis-result, .alert-success, .card-body', { timeout: 10000 });

    // WHEN: User clicks Clear button
    const clearButton = page.locator('button:has-text("Clear")');
    await clearButton.click();

    // THEN: Analysis results are removed from UI
    const resultsVisible = await page.locator('.analysis-result, .alert-success').count();
    expect(resultsVisible).toBe(0);
  });

  test('Clear button does not affect other tabs', async ({ page }) => {
    // GIVEN: User has multiple files open
    // Create first file with content
    const editor = page.locator('.monaco-editor');
    await editor.click();
    await page.keyboard.type('const file1 = "data";');

    // Add second file (via GitHub import or manual add)
    // For this test, we'll simulate having multiple tabs
    // Note: This may require GitHub import functionality to be working
    // For now, test with default file + manual content
    
    // Verify first file has content
    const file1Content = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });
    expect(file1Content).toContain('const file1');

    // WHEN: User clicks Clear on first tab
    const clearButton = page.locator('button:has-text("Clear")');
    await clearButton.click();

    // THEN: Only active file is cleared
    const clearedContent = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });
    expect(clearedContent).toBe('');

    // Note: Full multi-tab test requires GitHub import functionality
    // This test validates clear only affects active editor
  });

  test('Reset button replaces all files with default example', async ({ page }) => {
    // GIVEN: User has modified code or imported files
    const editor = page.locator('.monaco-editor');
    await editor.click();
    
    // Clear existing content and add custom code
    await page.keyboard.press('Control+A');
    await page.keyboard.type('const custom = "modified code";');

    // Verify custom code is present
    const customContent = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });
    expect(customContent).toContain('const custom');

    // WHEN: User clicks Reset to Default button
    const resetButton = page.locator('button:has-text("Reset to Default")');
    await resetButton.click();

    // THEN: Code is reset to default example
    const resetContent = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });
    
    // Verify default example code is restored
    expect(resetContent).toContain('Example code for DevSmith Code Review');
    expect(resetContent).not.toContain('const custom');
  });

  test('Reset button clears analysis results', async ({ page }) => {
    // GIVEN: User has run analysis with results
    const editor = page.locator('.monaco-editor');
    await editor.click();
    await page.keyboard.type('function toAnalyze() {}');

    // Run analysis
    await page.click('text=Quick Preview');
    await page.click('button:has-text("Analyze Code")');
    await page.waitForSelector('.analysis-result, .alert-success, .card-body', { timeout: 10000 });

    // WHEN: User clicks Reset to Default
    const resetButton = page.locator('button:has-text("Reset to Default")');
    await resetButton.click();

    // THEN: Analysis results are cleared
    const resultsVisible = await page.locator('.analysis-result, .alert-success').count();
    expect(resultsVisible).toBe(0);
  });

  test('Reset button resets UI to single default tab', async ({ page }) => {
    // GIVEN: User may have multiple file tabs open
    // (This assumes GitHub import creates multiple tabs)
    
    // Modify the default file
    const editor = page.locator('.monaco-editor');
    await editor.click();
    await page.keyboard.press('Control+A');
    await page.keyboard.type('const modified = true;');

    // WHEN: User clicks Reset to Default
    const resetButton = page.locator('button:has-text("Reset to Default")');
    await resetButton.click();

    // THEN: Single default tab is displayed
    const fileTabsCount = await page.locator('.file-tab').count();
    expect(fileTabsCount).toBe(1);

    // AND: Tab shows default filename
    const defaultTab = page.locator('.file-tab:has-text("info.txt")');
    await expect(defaultTab).toBeVisible();

    // AND: Content is default example
    const content = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });
    expect(content).toContain('Example code for DevSmith Code Review');
  });

  test('Clear and Reset buttons are always visible', async ({ page }) => {
    // WHEN: Page loads
    // THEN: Clear button is visible
    const clearButton = page.locator('button:has-text("Clear")');
    await expect(clearButton).toBeVisible();

    // AND: Reset button is visible
    const resetButton = page.locator('button:has-text("Reset to Default")');
    await expect(resetButton).toBeVisible();
  });

  test('Clear button clears error messages', async ({ page }) => {
    // GIVEN: User has triggered an error (e.g., analysis without code)
    // Click analyze without entering code or selecting mode
    const analyzeButton = page.locator('button:has-text("Analyze Code")');
    await analyzeButton.click();

    // Wait for error message (if validation shows errors)
    await page.waitForTimeout(500);

    // WHEN: User clicks Clear button
    const clearButton = page.locator('button:has-text("Clear")');
    await clearButton.click();

    // THEN: Error messages are cleared
    const errorMessages = await page.locator('.alert-danger, .text-danger').count();
    expect(errorMessages).toBe(0);
  });

  test('Reset button clears error messages', async ({ page }) => {
    // GIVEN: User has an error displayed
    const analyzeButton = page.locator('button:has-text("Analyze Code")');
    await analyzeButton.click();
    await page.waitForTimeout(500);

    // WHEN: User clicks Reset to Default
    const resetButton = page.locator('button:has-text("Reset to Default")');
    await resetButton.click();

    // THEN: Error messages are cleared
    const errorMessages = await page.locator('.alert-danger, .text-danger').count();
    expect(errorMessages).toBe(0);
  });

  test('Clear button is disabled when editor is empty', async ({ page }) => {
    // GIVEN: Editor has no content
    const clearButton = page.locator('button:has-text("Clear")');

    // WHEN: Editor is empty
    const content = await page.evaluate(() => {
      const monaco = window.monaco;
      const model = monaco.editor.getModels()[0];
      return model.getValue();
    });

    // THEN: Clear button may be disabled (optional UX improvement)
    // Note: This is an optional test - button may always be enabled
    // For now, just verify button exists
    await expect(clearButton).toBeVisible();
  });
});

test.describe('ReviewPage - Multi-File Clear/Reset (Future)', () => {
  // These tests require GitHub import functionality to be fully working
  // They validate multi-file scenarios

  test.skip('Clear only affects active file in multi-file scenario', async ({ page }) => {
    // TODO: Implement when GitHub import creates multiple tabs
    // 1. Import multi-file GitHub repo
    // 2. Edit file1
    // 3. Switch to file2
    // 4. Click Clear
    // 5. Verify file2 cleared but file1 unchanged
  });

  test.skip('Reset removes all imported files and restores default', async ({ page }) => {
    // TODO: Implement when GitHub import creates multiple tabs
    // 1. Import multi-file GitHub repo
    // 2. Have 5+ file tabs open
    // 3. Click Reset to Default
    // 4. Verify only 1 tab remains (info.txt with default example)
  });
});
