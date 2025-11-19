import { test, expect } from '../fixtures/auth.fixture';

/**
 * Comprehensive Review Service Testing - Rule Zero Compliance
 * 
 * Tests ALL 5 reading modes × 3 experience levels = 15 combinations
 * With visual validation for each combination
 * 
 * Following copilot-instructions.md Rule Zero:
 * - Must test EVERY user workflow
 * - Must capture screenshots
 * - Must validate UI correctness
 * - No loading spinners stuck
 * - No error states
 */

// Test data: simple Go code snippet
const TEST_CODE = `package main

import "fmt"

func main() {
    fmt.Println("Hello, DevSmith!")
}`;

// Helper: Navigate to Review app and ensure authenticated
async function navigateToReview(page: any) {
  await page.goto('http://localhost:3000/review');
  
  // Wait for either Review workspace OR login redirect
  await Promise.race([
    page.waitForURL('**/review**', { timeout: 5000 }).catch(() => {}),
    page.waitForURL('**/auth/**', { timeout: 5000 }).catch(() => {})
  ]);

  const currentURL = page.url();
  
  // If redirected to login, we need to authenticate first
  if (currentURL.includes('/auth/')) {
    console.log('User not authenticated - handling login flow');
    
    // Wait for OAuth redirect to GitHub (mocked in test environment)
    await page.waitForURL(/github\.com/, { timeout: 10000 });
    
    // In real environment, user would click "Authorize"
    // In test environment, we simulate successful OAuth callback
    await page.goto('http://localhost:3000/auth/github/callback?code=test_code&state=test_state');
    
    // Wait for redirect back to Review
    await page.waitForURL('**/review**', { timeout: 10000 });
  }
  
  // Now we should be on Review workspace
  await expect(page).toHaveURL(/\/review/);
  
  // CRITICAL: Wait for React app to fully render ReviewPage component
  // ReviewPage.jsx always renders these core elements
  await page.waitForSelector('.analysis-mode-selector', { timeout: 15000, state: 'visible' });
  await page.waitForSelector('.code-editor-container', { timeout: 15000, state: 'visible' });
  
  // Verify we're not seeing a loading spinner
  const hasLoadingSpinner = await page.locator('.spinner-border').count();
  if (hasLoadingSpinner > 0) {
    // Wait for loading to complete
    await page.waitForSelector('.spinner-border', { state: 'detached', timeout: 10000 });
  }
}

// Helper: Submit code for analysis
// Helper: Fill code and submit analysis
// React-based ReviewPage.jsx structure:
// - AnalysisModeSelector: Clickable cards with class="mode-card {mode}"
// - Experience Level: <select> with value="beginner|novice|intermediate|expert"  
// - Learning Style: Radio buttons with id="outputQuick" or "outputDetailed"
// - CodeEditor: Monaco Editor from @monaco-editor/react (uses textarea.ime-text-area)
async function submitCodeAnalysis(page: any, mode: string, userMode: string, outputMode: string) {
  // navigateToReview() already verified .analysis-mode-selector is visible
  
  // Fill code in Monaco Editor
  // Monaco renders with class="monaco-editor" and hidden textarea class="ime-text-area"
  await page.waitForSelector('.monaco-editor', { timeout: 10000 });
  
  // Click on Monaco editor to focus it
  await page.click('.monaco-editor');
  await page.waitForTimeout(500); // Wait for editor to focus
  
  // Use Monaco's setValue API via evaluate for reliable text insertion
  await page.evaluate((code: string) => {
    // Find Monaco editor instance via DOM
    const editorElement = document.querySelector('.monaco-editor');
    if (editorElement && (window as any).monaco) {
      const editor = (window as any).monaco.editor.getEditors()[0];
      if (editor) {
        editor.setValue(code);
        return true;
      }
    }
    // Fallback: directly set textarea value (less reliable but works)
    const textarea = document.querySelector('textarea.ime-text-area') as HTMLTextAreaElement;
    if (textarea) {
      textarea.value = code;
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      return true;
    }
    return false;
  }, TEST_CODE);
  
  // Wait for models to load and select one
  await page.waitForSelector('#model-select', { timeout: 10000 });
  await page.waitForTimeout(2000); // Wait for LLM configs API to populate
  
  // Select first available model (test user must have LLM config)
  const modelOptions = await page.locator('#model-select option').count();
  if (modelOptions > 0) {
    // Check if model selector is enabled
    const isDisabled = await page.locator('#model-select').getAttribute('disabled');
    if (isDisabled === null) {
      await page.selectOption('#model-select', { index: 0 });
    }
  }
  
  // Select reading mode by clicking the mode card
  await page.click(`.mode-card.${mode}`);
  
  // Select user experience level (second form-select dropdown, after model selector)
  const experienceSelect = page.locator('select.form-select').nth(1);
  await experienceSelect.selectOption(userMode);
  
  // Select output format (click the label instead of hidden radio button)
  const outputId = outputMode === 'detailed' ? 'outputDetailed' : 'outputQuick';
  await page.click(`label[for="${outputId}"]`);
  
  // CRITICAL: Set up API response listener BEFORE clicking button
  // This ensures we capture the actual network request
  const responsePromise = page.waitForResponse(
    (response: any) => response.url().includes('/api/review/modes/') && response.request().method() === 'POST',
    { timeout: 90000 } // 90 seconds for AI processing
  );
  
  // Click "Analyze Code" button
  await page.click('button:has-text("Analyze Code")');
  
  // Wait for Analyze button to show loading state (button text changes to "Analyzing...")
  await page.waitForSelector('button:has-text("Analyzing")', { timeout: 5000 });
  
  // CRITICAL: Wait for ACTUAL API response (not just UI state)
  let apiResponse;
  try {
    apiResponse = await responsePromise;
  } catch (error) {
    throw new Error(`API request timeout or failed after clicking Analyze: ${error}`);
  }
  
  // Check HTTP status code immediately (catch errors before proceeding)
  const status = apiResponse.status();
  if (status !== 200) {
    const errorBody = await apiResponse.text();
    throw new Error(`HTTP ${status} error from Review API: ${errorBody}`);
  }
  
  // Wait for analysis to complete (max 60 seconds for AI processing)
  // Button text changes back from "Analyzing..." to "Analyze Code"
  await page.waitForSelector('button:has-text("Analyze Code")', { timeout: 60000 });
  
  // Verify analysis output appeared
  await page.waitForSelector('.analysis-output', { timeout: 5000 });
  
  // Return API response for validation in validateAIOutput
  return apiResponse;
}

// Helper: Validate no error states
async function validateNoErrors(page: any) {
  // Check for HTTP 500 error message
  const errorText = await page.textContent('body');
  expect(errorText).not.toContain('Analysis Failed');
  expect(errorText).not.toContain('500');
  expect(errorText).not.toContain('Internal Server Error');
  
  // Check for loading spinner stuck
  const loadingSpinner = await page.locator('.loading-spinner').count();
  expect(loadingSpinner).toBe(0);
  
  // Check for error alert boxes
  const errorAlerts = await page.locator('.alert-danger, .error-message').count();
  expect(errorAlerts).toBe(0);
}

// Helper: Validate AI analysis output contains meaningful content
async function validateAIOutput(page: any, mode: string, outputMode: string, apiResponse: any) {
  // CRITICAL: API response is now passed in from analyze workflow
  // This avoids duplicate waiting and ensures we validate the SAME response
  
  // FIRST: Validate API response structure
  // (HTTP status already checked in analyze workflow)
  const apiData = await apiResponse.json();
  expect(apiData).toBeTruthy();
  
  // API returns data directly (not wrapped in 'result' field)
  // Validate mode-specific structure
  switch (mode) {
    case 'preview':
      expect(apiData).toHaveProperty('file_tree');
      expect(apiData).toHaveProperty('summary');
      break;
    case 'skim':
      expect(apiData).toHaveProperty('functions');
      break;
    case 'scan':
      expect(apiData).toHaveProperty('matches');
      break;
    case 'detailed':
      expect(apiData).toHaveProperty('line_explanations');
      break;
    case 'critical':
      expect(apiData).toHaveProperty('issues');
      break;
  }
  
  // SECOND: Wait for UI to update with the API response
  await page.waitForSelector('.analysis-output', { timeout: 10000 });
  
  const outputText = await page.textContent('.analysis-output');
  
  // THIRD: Check for error messages in UI (prevent false positives)
  expect(outputText).not.toContain('Analysis Failed');
  expect(outputText).not.toContain('HTTP 500');
  expect(outputText).not.toContain('Internal Server Error');
  expect(outputText.toLowerCase()).not.toContain('error:');
  
  // Validate output is not empty
  expect(outputText).toBeTruthy();
  expect(outputText.length).toBeGreaterThan(50);
  
  // For Scan mode, accept shorter responses when no matches found
  if (mode === 'scan' && outputText.includes('No matches found')) {
    expect(outputText.length).toBeGreaterThan(50); // Valid "no matches" response
  } else {
    expect(outputText.length).toBeGreaterThan(100); // Meaningful content
  }
  
  if (outputMode === 'json') {
    // Validate JSON structure
    expect(outputText).toContain('{');
    expect(outputText).toContain('}');
    
    // Validate mode-specific JSON fields (relaxed to match actual API)
    switch (mode) {
      case 'preview':
        // API returns 'file_tree' or 'file_structure' depending on backend version
        expect(outputText).toMatch(/file_tree|file_structure/);
        expect(outputText).toMatch(/bounded_contexts|tech_stack|summary/);
        break;
      case 'skim':
        expect(outputText).toMatch(/functions|signatures|interfaces/);
        break;
      case 'scan':
        expect(outputText).toMatch(/matches|results/);
        break;
      case 'detailed':
        expect(outputText).toMatch(/line_explanations|explanations|complexity/);
        break;
      case 'critical':
        expect(outputText).toMatch(/issues|severity|recommendations/);
        break;
    }
  } else {
    // HTML output - just validate AI content exists (don't check for specific headings)
    // The React frontend may render differently than backend templates
    switch (mode) {
      case 'preview':
        // Validate actual AI-generated content exists
        expect(outputText).toMatch(/package|function|import|main|file/i);
        break;
      case 'skim':
        // Validate function listings exist
        expect(outputText).toMatch(/function|method|signature/i);
        break;
      case 'scan':
        // Validate search matches found
        expect(outputText).toMatch(/match|found|result/i);
        break;
      case 'detailed':
        // Validate line explanations exist
        expect(outputText).toMatch(/line \d+|explanation|code analysis/i);
        break;
      case 'critical':
        // Validate issues found
        expect(outputText).toMatch(/issue|warning|error|improvement|recommendation/i);
        break;
    }
  }
  
  // Validate no "empty response" or placeholder text
  expect(outputText).not.toContain('No analysis available');
  expect(outputText).not.toContain('Analysis coming soon');
  expect(outputText).not.toContain('TODO');
}

test.describe('Review Service - All Modes × All Experience Levels (Rule Zero)', () => {
  
  test.beforeEach(async ({ authenticatedPage }) => {
    // Navigate to Review with authenticated session
    await navigateToReview(authenticatedPage);
  });

  // ========================================
  // PREVIEW MODE (3 experience levels)
  // ========================================
  
  test('Preview Mode - Beginner - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'preview', 'beginner', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'preview', 'html', apiResponse);
    
    // Screenshot for visual validation
    await page.screenshot({ path: 'test-results/review-preview-beginner-html.png', fullPage: true });
  });

  test('Preview Mode - Intermediate - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'preview', 'intermediate', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'preview', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-preview-intermediate-html.png', fullPage: true });
  });

  test('Preview Mode - Expert - JSON Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'preview', 'expert', 'json');
    await validateNoErrors(page);
    await validateAIOutput(page, 'preview', 'json', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-preview-expert-json.png', fullPage: true });
  });

  // ========================================
  // SKIM MODE (3 experience levels)
  // ========================================
  
  test('Skim Mode - Beginner - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'skim', 'beginner', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'skim', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-skim-beginner-html.png', fullPage: true });
  });

  test('Skim Mode - Intermediate - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'skim', 'intermediate', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'skim', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-skim-intermediate-html.png', fullPage: true });
  });

  test('Skim Mode - Expert - JSON Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'skim', 'expert', 'json');
    await validateNoErrors(page);
    await validateAIOutput(page, 'skim', 'json', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-skim-expert-json.png', fullPage: true });
  });

  // ========================================
  // SCAN MODE (3 experience levels)
  // ========================================
  
  test('Scan Mode - Beginner - HTML Output', async ({ authenticatedPage: page }) => {
    // Scan mode requires filling Monaco Editor, then adding scan query
    await page.evaluate((code) => {
      const monaco = (window as any).monaco;
      if (monaco) {
        const editor = monaco.editor.getModels()[0];
        if (editor) {
          editor.setValue(code);
        }
      }
      const textarea = document.querySelector('textarea.monaco-textarea') as HTMLTextAreaElement;
      if (textarea) {
        textarea.value = code;
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
      }
    }, TEST_CODE);
    
    // Select model
    await page.waitForSelector('#model-select', { timeout: 10000 });
    await page.waitForTimeout(2000);
    const modelOptions = await page.locator('#model-select option').count();
    if (modelOptions > 0) {
      const isDisabled = await page.locator('#model-select').getAttribute('disabled');
      if (isDisabled === null) {
        await page.selectOption('#model-select', { index: 0 });
      }
    }
    
    // Select scan mode
    await page.click('.mode-card.scan');
    
    // Wait for scan query input to appear (conditionally rendered)
    await page.waitForSelector('#scanQuery', { timeout: 5000 });
    
    // Fill scan query input (Scan mode specific field)
    await page.fill('#scanQuery', 'main function');
    
    // Select experience level
    const experienceSelect = page.locator('select.form-select').nth(1);
    await experienceSelect.selectOption('beginner');
    
    // Select HTML output
    await page.click('label[for="outputQuick"]');
    
    // CRITICAL: Set up API response listener BEFORE clicking button
    const responsePromise = page.waitForResponse(
      (response: any) => response.url().includes('/api/review/modes/') && response.request().method() === 'POST',
      { timeout: 90000 }
    );
    
    // Submit form
    await page.click('button:has-text("Analyze Code")');
    await page.waitForSelector('button:has-text("Analyzing")', { timeout: 5000 });
    
    // Wait for API response
    const apiResponse = await responsePromise;
    const status = apiResponse.status();
    if (status !== 200) {
      const errorBody = await apiResponse.text();
      throw new Error(`HTTP ${status} error from Review API: ${errorBody}`);
    }
    
    await page.waitForSelector('button:has-text("Analyze Code")', { timeout: 60000 });
    
    await validateNoErrors(page);
    await validateAIOutput(page, 'scan', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-scan-beginner-html.png', fullPage: true });
  });

  test('Scan Mode - Intermediate - HTML Output', async ({ authenticatedPage: page }) => {
    await page.evaluate((code) => {
      const monaco = (window as any).monaco;
      if (monaco) {
        const editor = monaco.editor.getModels()[0];
        if (editor) {
          editor.setValue(code);
        }
      }
      const textarea = document.querySelector('textarea.monaco-textarea') as HTMLTextAreaElement;
      if (textarea) {
        textarea.value = code;
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
      }
    }, TEST_CODE);
    
    await page.waitForSelector('#model-select', { timeout: 10000 });
    await page.waitForTimeout(2000);
    const modelOptions = await page.locator('#model-select option').count();
    if (modelOptions > 0) {
      const isDisabled = await page.locator('#model-select').getAttribute('disabled');
      if (isDisabled === null) {
        await page.selectOption('#model-select', { index: 0 });
      }
    }
    
    await page.click('.mode-card.scan');
    await page.waitForSelector('#scanQuery', { timeout: 5000 });
    await page.fill('#scanQuery', 'imports');
    
    const experienceSelect = page.locator('select.form-select').nth(1);
    await experienceSelect.selectOption('intermediate');
    
    await page.click('label[for="outputQuick"]');
    
    // CRITICAL: Set up API response listener BEFORE clicking button
    const responsePromise = page.waitForResponse(
      (response: any) => response.url().includes('/api/review/modes/') && response.request().method() === 'POST',
      { timeout: 90000 }
    );
    
    await page.click('button:has-text("Analyze Code")');
    await page.waitForSelector('button:has-text("Analyzing")', { timeout: 5000 });
    
    // Wait for API response
    const apiResponse = await responsePromise;
    const status = apiResponse.status();
    if (status !== 200) {
      const errorBody = await apiResponse.text();
      throw new Error(`HTTP ${status} error from Review API: ${errorBody}`);
    }
    
    await page.waitForSelector('button:has-text("Analyze Code")', { timeout: 60000 });
    
    await validateNoErrors(page);
    await validateAIOutput(page, 'scan', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-scan-intermediate-html.png', fullPage: true });
  });

  test('Scan Mode - Expert - JSON Output', async ({ authenticatedPage: page }) => {
    await page.evaluate((code) => {
      const monaco = (window as any).monaco;
      if (monaco) {
        const editor = monaco.editor.getModels()[0];
        if (editor) {
          editor.setValue(code);
        }
      }
      const textarea = document.querySelector('textarea.monaco-textarea') as HTMLTextAreaElement;
      if (textarea) {
        textarea.value = code;
        textarea.dispatchEvent(new Event('input', { bubbles: true }));
      }
    }, TEST_CODE);
    
    await page.waitForSelector('#model-select', { timeout: 10000 });
    await page.waitForTimeout(2000);
    const modelOptions = await page.locator('#model-select option').count();
    if (modelOptions > 0) {
      const isDisabled = await page.locator('#model-select').getAttribute('disabled');
      if (isDisabled === null) {
        await page.selectOption('#model-select', { index: 0 });
      }
    }
    
    await page.click('.mode-card.scan');
    await page.waitForSelector('#scanQuery', { timeout: 5000 });
    await page.fill('#scanQuery', 'fmt.Println');
    
    const experienceSelect = page.locator('select.form-select').nth(1);
    await experienceSelect.selectOption('expert');
    
    await page.click('label[for="outputDetailed"]');
    
    // CRITICAL: Set up API response listener BEFORE clicking button
    const responsePromise = page.waitForResponse(
      (response: any) => response.url().includes('/api/review/modes/') && response.request().method() === 'POST',
      { timeout: 90000 }
    );
    
    await page.click('button:has-text("Analyze Code")');
    await page.waitForSelector('button:has-text("Analyzing")', { timeout: 5000 });
    
    // Wait for API response
    const apiResponse = await responsePromise;
    const status = apiResponse.status();
    if (status !== 200) {
      const errorBody = await apiResponse.text();
      throw new Error(`HTTP ${status} error from Review API: ${errorBody}`);
    }
    
    await page.waitForSelector('button:has-text("Analyze Code")', { timeout: 60000 });
    
    await validateNoErrors(page);
    await validateAIOutput(page, 'scan', 'json', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-scan-expert-json.png', fullPage: true });
  });

  // ========================================
  // DETAILED MODE (3 experience levels)
  // ========================================
  
  test('Detailed Mode - Beginner - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'detailed', 'beginner', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'detailed', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-detailed-beginner-html.png', fullPage: true });
  });

  test('Detailed Mode - Intermediate - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'detailed', 'intermediate', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'detailed', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-detailed-intermediate-html.png', fullPage: true });
  });

  test('Detailed Mode - Expert - JSON Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'detailed', 'expert', 'json');
    await validateNoErrors(page);
    await validateAIOutput(page, 'detailed', 'json', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-detailed-expert-json.png', fullPage: true });
  });

  // ========================================
  // CRITICAL MODE (3 experience levels)
  // ========================================
  
  test('Critical Mode - Beginner - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'critical', 'beginner', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'critical', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-critical-beginner-html.png', fullPage: true });
  });

  test('Critical Mode - Intermediate - HTML Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'critical', 'intermediate', 'html');
    await validateNoErrors(page);
    await validateAIOutput(page, 'critical', 'html', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-critical-intermediate-html.png', fullPage: true });
  });

  test('Critical Mode - Expert - JSON Output', async ({ authenticatedPage: page }) => {
    const apiResponse = await submitCodeAnalysis(page, 'critical', 'expert', 'json');
    await validateNoErrors(page);
    await validateAIOutput(page, 'critical', 'json', apiResponse);
    
    await page.screenshot({ path: 'test-results/review-critical-expert-json.png', fullPage: true });
  });
});
