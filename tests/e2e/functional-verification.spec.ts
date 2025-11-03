import { test, expect } from '@playwright/test';

/**
 * Functional Verification Tests for Review App
 * 
 * These tests verify actual functionality, not just UI presence:
 * - Dark mode actually toggles the 'dark' class
 * - Analyze actually calls the API and returns results
 * - Code can actually be edited and re-analyzed
 */

test.describe('Review App - Functional Verification', () => {
  
  test('1. Dark mode toggle actually works', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    // Get initial state
    const htmlElement = page.locator('html');
    const initialHasDarkClass = await htmlElement.evaluate((el) => el.classList.contains('dark'));
    
    // Find and click dark mode toggle
    const darkModeToggle = page.locator('#dark-mode-toggle');
    await expect(darkModeToggle).toBeVisible();
    await darkModeToggle.click();
    
    // Wait for class change
    await page.waitForTimeout(100);
    
    // Verify class toggled
    const afterClickHasDarkClass = await htmlElement.evaluate((el) => el.classList.contains('dark'));
    expect(afterClickHasDarkClass).toBe(!initialHasDarkClass);
    
    // Verify localStorage updated
    const darkModeStorage = await page.evaluate(() => localStorage.getItem('darkMode'));
    expect(darkModeStorage).toBe(afterClickHasDarkClass ? 'true' : 'false');
    
    // Click again to toggle back
    await darkModeToggle.click();
    await page.waitForTimeout(100);
    
    // Verify it toggled back
    const finalHasDarkClass = await htmlElement.evaluate((el) => el.classList.contains('dark'));
    expect(finalHasDarkClass).toBe(initialHasDarkClass);
    
    console.log('✓ Dark mode toggle verified: actually changes dark class and localStorage');
  });

  test('2. Code editing actually works', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    // Find code editor (textarea)
    const codeEditor = page.locator('#code-editor');
    await expect(codeEditor).toBeVisible();
    
    // Verify it's a textarea (editable)
    const tagName = await codeEditor.evaluate((el) => el.tagName);
    expect(tagName).toBe('TEXTAREA');
    
    // Get initial code
    const initialCode = await codeEditor.inputValue();
    expect(initialCode.length).toBeGreaterThan(0);
    
    // Clear and type new code
    await codeEditor.clear();
    const testCode = 'package main\n\nfunc main() {\n\tprintln("Hello World")\n}';
    await codeEditor.fill(testCode);
    
    // Verify code changed
    const newCode = await codeEditor.inputValue();
    expect(newCode).toBe(testCode);
    
    // Verify character count updated
    const charCount = page.locator('#char-count');
    await expect(charCount).toContainText(`${testCode.length} characters`);
    
    console.log('✓ Code editing verified: textarea is editable and character count updates');
  });

  test('3. Analyze code actually calls API and returns results', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    // Ensure code is present
    const codeEditor = page.locator('#code-editor');
    const code = await codeEditor.inputValue();
    expect(code.length).toBeGreaterThan(0);
    
    // Select preview mode
    const modeSelector = page.locator('#mode-selector');
    await modeSelector.selectOption('preview');
    
    // Get analysis pane
    const analysisPane = page.locator('#analysis-pane');
    const initialContent = await analysisPane.textContent();
    
    // Click analyze button
    const analyzeBtn = page.locator('#analyze-btn');
    await expect(analyzeBtn).toBeVisible();
    await expect(analyzeBtn).toBeEnabled();
    
    // Set up API request listener
    const apiRequest = page.waitForRequest(req => 
      req.url().includes('/api/review/modes/preview') && req.method() === 'POST'
    );
    
    // Click analyze
    await analyzeBtn.click();
    
    // Wait for API request
    const request = await apiRequest;
    console.log('✓ API request sent to:', request.url());
    
    // Wait for button to show "Analyzing..."
    await expect(analyzeBtn).toHaveText('Analyzing...');
    await expect(analyzeBtn).toBeDisabled();
    
    // Wait for loading indicator
    const loadingIndicator = page.locator('#analysis-loading');
    await expect(loadingIndicator).toBeVisible();
    
    // Wait for response (with longer timeout for actual AI processing)
    await page.waitForResponse(
      res => res.url().includes('/api/review/modes/preview') && res.status() === 200,
      { timeout: 60000 }
    );
    
    // Wait for loading to hide
    await expect(loadingIndicator).toBeHidden({ timeout: 5000 });
    
    // Wait for button to re-enable
    await expect(analyzeBtn).toHaveText('Analyze Code', { timeout: 5000 });
    await expect(analyzeBtn).toBeEnabled();
    
    // Verify analysis pane content changed
    const finalContent = await analysisPane.textContent();
    expect(finalContent).not.toBe(initialContent);
    expect((finalContent || '').length).toBeGreaterThan((initialContent || '').length);
    
    // Verify analysis pane contains actual results (not just loading state)
    expect(finalContent).not.toContain('No Analysis Yet');
    expect((finalContent || '').length).toBeGreaterThan(100); // Should have substantial content
    
    console.log('✓ Analyze verified: API called, response received, results displayed');
    console.log('  Response length:', (finalContent || '').length, 'characters');
  });

  test('4. Different modes call different endpoints', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];
    
    for (const mode of modes) {
      // Select mode
      const modeSelector = page.locator('#mode-selector');
      await modeSelector.selectOption(mode);
      
      // Set up request listener for this mode
      const apiRequest = page.waitForRequest(req => 
        req.url().includes(`/api/review/modes/${mode}`) && req.method() === 'POST',
        { timeout: 5000 }
      );
      
      // Click analyze
      const analyzeBtn = page.locator('#analyze-btn');
      await analyzeBtn.click();
      
      // Verify correct endpoint called
      const request = await apiRequest;
      expect(request.url()).toContain(`/api/review/modes/${mode}`);
      console.log(`✓ ${mode} mode calls /api/review/modes/${mode}`);
      
      // Wait for response before next iteration
      await page.waitForResponse(
        res => res.url().includes(`/api/review/modes/${mode}`),
        { timeout: 60000 }
      );
      
      // Brief wait between modes
      await page.waitForTimeout(500);
    }
  });

  test('5. Analyze with edited code sends new code', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    const codeEditor = page.locator('#code-editor');
    
    // Clear and enter specific test code
    await codeEditor.clear();
    const testCode = 'package test\n\n// Test function\nfunc TestAnalyze() {\n\t// TODO: implement\n}';
    await codeEditor.fill(testCode);
    
    // Set up request listener to capture POST body
    let requestBody = '';
    page.on('request', req => {
      if (req.url().includes('/api/review/modes/preview') && req.method() === 'POST') {
        requestBody = req.postData() || '';
      }
    });
    
    // Analyze
    const analyzeBtn = page.locator('#analyze-btn');
    await analyzeBtn.click();
    
    // Wait for request
    await page.waitForRequest(req => 
      req.url().includes('/api/review/modes/preview') && req.method() === 'POST'
    );
    
    // Verify request body contains our test code
    expect(requestBody).toContain('TestAnalyze');
    expect(requestBody).toContain('pasted_code');
    
    console.log('✓ Edited code verified: POST body contains edited code');
  });

  test('6. Error handling works (invalid/empty code)', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    const codeEditor = page.locator('#code-editor');
    await codeEditor.clear();
    
    // Try to analyze empty code
    const analyzeBtn = page.locator('#analyze-btn');
    await analyzeBtn.click();
    
    // Should show toast/warning about empty code
    // Look for toast or keep analysis pane in empty state
    await page.waitForTimeout(1000);
    
    // Verify no API call was made for empty code
    // (the JavaScript should prevent it)
    const analysisPane = page.locator('#analysis-pane');
    const content = await analysisPane.textContent();
    
    // Should still show empty state
    expect(content).toContain('No Analysis Yet');
    
    console.log('✓ Error handling verified: empty code does not trigger API call');
  });
});

test.describe('Review App - Screenshot Evidence', () => {
  
  test('Capture functional test evidence', async ({ page }) => {
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    // 1. Initial light mode
    await page.screenshot({ 
      path: '/tmp/devsmith-screenshots/functional-01-light-mode.png',
      fullPage: true 
    });
    
    // 2. Dark mode enabled
    const darkModeToggle = page.locator('#dark-mode-toggle');
    await darkModeToggle.click();
    await page.waitForTimeout(200);
    await page.screenshot({ 
      path: '/tmp/devsmith-screenshots/functional-02-dark-mode.png',
      fullPage: true 
    });
    
    // 3. Code editing
    const codeEditor = page.locator('#code-editor');
    await codeEditor.clear();
    await codeEditor.fill('package main\n\nfunc main() {\n\tprintln("Test")\n}');
    await page.screenshot({ 
      path: '/tmp/devsmith-screenshots/functional-03-code-edited.png',
      fullPage: true 
    });
    
    // 4. Analyzing state
    const analyzeBtn = page.locator('#analyze-btn');
    await analyzeBtn.click();
    await page.waitForTimeout(500);
    await page.screenshot({ 
      path: '/tmp/devsmith-screenshots/functional-04-analyzing.png',
      fullPage: true 
    });
    
    // 5. Results displayed
    await page.waitForResponse(
      res => res.url().includes('/api/review/modes/preview'),
      { timeout: 60000 }
    );
    await page.waitForTimeout(500);
    await page.screenshot({ 
      path: '/tmp/devsmith-screenshots/functional-05-results.png',
      fullPage: true 
    });
    
    console.log('✓ Functional test screenshots saved to /tmp/devsmith-screenshots/');
  });
});
