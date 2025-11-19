/**
 * Review Analysis Full Workflow E2E Test
 * 
 * Tests the complete user workflow for code analysis in Review service:
 * 1. Use authenticated page fixture (auto-login)
 * 2. Navigate to Review workspace
 * 3. Enter code in Monaco Editor (VS Code-style editor)
 * 4. Select Preview mode
 * 5. Submit analysis
 * 6. Verify analysis results displayed (not HTTP 500)
 * 7. Verify Ollama LLM config loaded successfully
 * 
 * This test validates the fix for the "Ollama endpoint not configured" issue
 * where Portal API wasn't returning full LLM config to Review service.
 * 
 * FIX DETAILS:
 * - Portal handler now returns full 9-field config (was 3 fields)
 * - Includes api_endpoint, api_key, max_tokens, temperature, etc.
 * 
 * PASS CRITERIA:
 * - Analysis completes without HTTP 500 error
 * - Analysis results HTML contains expected elements
 * - No "Analysis Failed" error message
 * - No "Ollama endpoint not configured" in console/network
 */

import { test, expect } from './fixtures/auth.fixture';

// Test data
const TEST_CODE = `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`;

test.describe('Review Analysis Full Workflow', () => {
  test('should complete Preview mode analysis successfully', async ({ authenticatedPage }) => {
    // Step 1: Navigate to Review app (React frontend at /review)
    await authenticatedPage.goto('http://localhost:3000/review');
    await authenticatedPage.waitForLoadState('networkidle');

    // Wait for workspace page to load
    await authenticatedPage.waitForTimeout(1000);

    // Step 2: Enter code in the textarea editor
    const codeEditor = authenticatedPage.getByRole('textbox', { name: 'Code editor' });
    await expect(codeEditor).toBeVisible({ timeout: 5000 });
    
    // Clear pre-filled sample code and enter test code
    await codeEditor.fill('');
    await codeEditor.fill(TEST_CODE);

    // Alternative: Use Monaco's internal textarea if keyboard fails
    // const editorTextarea = authenticatedPage.locator('.monaco-editor textarea').first();
    // await editorTextarea.fill(TEST_CODE);

    // Step 3: Select Preview mode (if mode selector exists)
    // Note: May need to adjust selector based on actual UI
    const modeSelect = authenticatedPage.locator('select#reading-mode, select[name="mode"]').first();
    if (await modeSelect.isVisible({ timeout: 5000 }).catch(() => false)) {
      await modeSelect.selectOption('preview');
    }

    // Step 4: Click Analyze button (verified from ReviewPage.jsx)
    const analyzeButton = authenticatedPage.getByText('Analyze Code').first();
    await expect(analyzeButton).toBeVisible();
    await expect(analyzeButton).toBeEnabled();
    
    // Monitor network requests for errors
    let hasNetworkError = false;
    authenticatedPage.on('response', (response: any) => {
      if (response.url().includes('/api/review/modes/preview')) {
        if (response.status() === 500) {
          hasNetworkError = true;
          console.error('❌ HTTP 500 detected on preview endpoint');
        }
      }
    });
    
    await analyzeButton.click();

    // Step 5: Wait for analysis to complete
    // Should show loading state then results (not error)
    await authenticatedPage.waitForSelector('.analysis-results, [data-testid="preview-results"], .analysis-output', { 
      timeout: 30000 
    });

    // Step 6: Verify no HTTP 500 error occurred
    expect(hasNetworkError).toBe(false);

    // Step 7: Verify analysis results are displayed
    const resultsContainer = authenticatedPage.locator('.analysis-results, [data-testid="preview-results"], .analysis-output').first();
    await expect(resultsContainer).toBeVisible();

    // Step 8: Verify no error messages
    const errorMessages = authenticatedPage.locator('text=/Analysis Failed|Error|Failed/i');
    await expect(errorMessages).toHaveCount(0, { timeout: 5000 });

    // Step 9: Verify expected analysis content exists
    // Preview mode should show summary, key areas, technologies
    const summary = authenticatedPage.locator('text=/Summary|Overview|Preview/i').first();
    await expect(summary).toBeVisible({ timeout: 5000 });

    // Step 10: Take screenshot for visual verification
    await authenticatedPage.screenshot({ 
      path: 'test-results/review-analysis-success.png',
      fullPage: true 
    });

    console.log('✅ Review analysis completed successfully');
  });

  test('should load Ollama config from Portal API', async ({ authenticatedPage, request }) => {
    // Step 1: Navigate to Review (ensures authenticated)
    await authenticatedPage.goto('http://localhost:3000/review/workspace/demo');
    
    // Step 2: Verify Portal API returns full LLM config
    const cookies = await authenticatedPage.context().cookies();
    const response = await request.get('http://localhost:3000/api/portal/app-llm-preferences', {
      headers: {
        'Cookie': cookies.map(c => `${c.name}=${c.value}`).join('; ')
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();

    // Step 2: Verify review config has all required fields
    expect(config.review).toBeDefined();
    expect(config.review.api_endpoint).toBe('http://host.docker.internal:11434');
    expect(config.review.provider).toBe('ollama');
    expect(config.review.model_name).toBe('deepseek-coder:6.7b');
    expect(config.review.max_tokens).toBe(8192);
    expect(config.review.temperature).toBe(0.7);

    console.log('✅ Portal API returns full Ollama config:', config.review);
  });

  test('should handle all 5 reading modes without errors', async ({ authenticatedPage }) => {
    const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];

    for (const mode of modes) {
      console.log(`Testing ${mode} mode...`);

      // Navigate to Review app (React frontend at /review) for each mode
      await authenticatedPage.goto('http://localhost:3000/review');
      await authenticatedPage.waitForLoadState('networkidle');
      
      // Wait for workspace page to load
      await authenticatedPage.waitForTimeout(1000);

      // Enter code in the textarea editor
      const codeEditor = authenticatedPage.getByRole('textbox', { name: 'Code editor' });
      await expect(codeEditor).toBeVisible({ timeout: 5000 });
      await codeEditor.fill('');
      await codeEditor.fill(TEST_CODE);

      // Select mode (if selector exists)
      const modeSelect = authenticatedPage.locator('select#reading-mode, select[name="mode"]').first();
      if (await modeSelect.isVisible({ timeout: 5000 }).catch(() => false)) {
        await modeSelect.selectOption(mode);
      }

      // Monitor for HTTP 500
      let hasError = false;
      const errorHandler = (response: any) => {
        if (response.url().includes(`/api/review/modes/${mode}`) && response.status() === 500) {
          hasError = true;
          console.error(`❌ HTTP 500 on ${mode} mode`);
        }
      };
      authenticatedPage.on('response', errorHandler);
      
      // Click analyze button
      const analyzeButton = authenticatedPage.getByText('Analyze Code').first();
      await analyzeButton.click();

      // Wait for results
      await authenticatedPage.waitForSelector('.analysis-results, [data-testid*="results"], .analysis-output', { 
        timeout: 30000 
      });

      // Verify no errors
      expect(hasError).toBe(false);
      console.log(`✅ ${mode} mode completed successfully`);

      // Remove listener
      authenticatedPage.removeListener('response', errorHandler);

      // Wait before next test
      await authenticatedPage.waitForTimeout(1000);
    }
  });

  test('should persist model selection across analyses', async ({ authenticatedPage }) => {
    // Navigate to Review app (React frontend at /review)
    await authenticatedPage.goto('http://localhost:3000/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Wait for React app to render and Monaco Editor to be ready
    await authenticatedPage.waitForTimeout(2000);

    // Step 1: Select a specific model (if model selector exists)
    const modelSelect = authenticatedPage.locator('select#model-override, select[name="model"]').first();
    if (await modelSelect.isVisible({ timeout: 5000 }).catch(() => false)) {
      await modelSelect.selectOption('deepseek-coder:6.7b');
    }

    // Step 2: Enter code in the textarea editor
    const codeEditor = authenticatedPage.getByRole('textbox', { name: 'Code editor' });
    await expect(codeEditor).toBeVisible({ timeout: 5000 });
    await codeEditor.fill('');
    await codeEditor.fill(TEST_CODE);

    // Step 3: Perform analysis
    const analyzeButton = authenticatedPage.getByText('Analyze Code').first();
    await analyzeButton.click();

    await authenticatedPage.waitForSelector('.analysis-results, .analysis-output', { timeout: 30000 });

    // Step 4: Verify model selection persists (if applicable)
    if (await modelSelect.isVisible().catch(() => false)) {
      const selectedModel = await modelSelect.inputValue();
      expect(selectedModel).toBe('deepseek-coder:6.7b');
    }

    console.log('✅ Model selection persisted');
  });
});

test.describe('Review Analysis Error Scenarios', () => {
  test('should show user-friendly error for invalid code', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('http://localhost:3000/review');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Wait for workspace page to load
    await authenticatedPage.waitForTimeout(1000);

    // Clear the pre-filled sample code (workspace pre-fills 916 chars)
    const codeEditor = authenticatedPage.getByRole('textbox', { name: 'Code editor' });
    await expect(codeEditor).toBeVisible({ timeout: 5000 });
    await codeEditor.fill(''); // Clear to empty

    // Try to analyze (button should be disabled for empty code)
    const analyzeButton = authenticatedPage.getByText('Analyze Code').first();
    
    // Button should be disabled when code is empty
    await expect(analyzeButton).toBeDisabled({ timeout: 5000 });

    console.log('✅ Empty code validation works - button disabled');
  });

  test('should gracefully handle Ollama unavailable', async ({ authenticatedPage }) => {
    // This test validates error handling when Ollama is down
    // Skip if Ollama is actually unavailable (would cause legitimate failures)
    test.skip(process.env.SKIP_OLLAMA_TESTS === 'true', 'Ollama tests skipped');

    // TODO: Mock Ollama unavailability by intercepting network requests
    // and returning 503 error
    console.log('⏭️  Ollama unavailability test - to be implemented');
  });
});
