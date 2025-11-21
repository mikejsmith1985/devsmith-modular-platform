import { test, expect } from '../fixtures/auth.fixture';

const TEST_CODE = `function example() {
  return "test";
}`;

test('Review Page - Form Submission Test', async ({ authenticatedPage: page }) => {
  console.log('=== Starting Review Page Form Submission Test ===');
  
  // Navigate to Review page
  await page.goto('/review', { waitUntil: 'networkidle' });
  await page.waitForSelector('.analysis-mode-selector', { timeout: 15000 });
  await page.waitForSelector('.code-editor-container', { timeout: 15000 });
  console.log('✓ Review page loaded');
  
  // Fill code using Monaco Editor
  await page.waitForSelector('.monaco-editor', { timeout: 10000 });
  await page.click('.monaco-editor');
  await page.waitForTimeout(500);
  
  const codeSet = await page.evaluate((code: string) => {
    const textarea = document.querySelector('textarea.ime-text-area') as HTMLTextAreaElement;
    if (textarea) {
      textarea.value = code;
      textarea.dispatchEvent(new Event('input', { bubbles: true }));
      return true;
    }
    return false;
  }, TEST_CODE);
  
  expect(codeSet).toBe(true);
  console.log('✓ Code entered into Monaco Editor');
  
  // Wait for models to load and select one (ModelSelector component)
  await page.waitForSelector('#model-select', { timeout: 5000 });
  await page.waitForTimeout(1000); // Wait for models API to populate dropdown
  
  // Select first available model
  const modelOptions = await page.locator('#model-select option').count();
  console.log(`Found ${modelOptions} model options`);
  if (modelOptions > 0) {
    await page.selectOption('#model-select', { index: 0 });
    console.log('✓ Model selected');
  } else {
    console.warn('⚠ No models available - button may be disabled');
  }
  
  // Select Preview mode
  await page.click('.mode-card.preview');
  console.log('✓ Preview mode selected');
  
  // Select experience level (second form-select, after model selector)
  // Use label text to find the correct select element
  const experienceSelect = page.locator('select.form-select').nth(1); // Second select element
  await experienceSelect.selectOption('intermediate');
  console.log('✓ Experience level selected');
  
  // Select output format (click the label, not the hidden radio button)
  await page.click('label[for="outputQuick"]');
  console.log('✓ Output format selected');
  
  // Click Analyze button
  const analyzeButton = page.locator('button:has-text("Analyze Code")');
  await expect(analyzeButton).toBeEnabled();
  await analyzeButton.click();
  console.log('✓ Analyze button clicked');
  
  // Wait for loading spinner to appear (proves request was sent)
  await page.waitForSelector('.spinner-border', { timeout: 5000 });
  console.log('✓ Loading spinner appeared - API request sent');
  
  // Take screenshot
  await page.screenshot({ path: '/tmp/review-form-submitted.png', fullPage: true });
  console.log('✓ Screenshot saved to /tmp/review-form-submitted.png');
  
  console.log('=== Test Passed: Form submission working ===');
});
