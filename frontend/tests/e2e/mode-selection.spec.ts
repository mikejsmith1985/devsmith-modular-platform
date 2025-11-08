import { test, expect } from '@playwright/test';

/**
 * E2E Test: Mode Selection User Flow
 * 
 * Tests that user experience level and learning style selections
 * properly affect AI analysis output.
 * 
 * Test Scenarios:
 * 1. Beginner + Full Learn → Should show simple language with analogies AND reasoning trace
 * 2. Expert + Quick Learn → Should show technical terms, concise, NO reasoning trace
 * 3. Intermediate + Quick (defaults) → Standard terminology, concise
 */

test.describe('Review Mode Selection', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to review page
    await page.goto('http://localhost:3000');
    
    // Wait for page to load
    await expect(page).toHaveTitle(/DevSmith/);
  });

  test('should display mode selection controls', async ({ page }) => {
    // Verify Experience Level dropdown exists
    const experienceLevelDropdown = page.locator('select[name="experience_level"], select:has-text("Experience Level")').first();
    await expect(experienceLevelDropdown).toBeVisible();
    
    // Verify Learning Style toggle/buttons exist
    const quickLearnButton = page.locator('button:has-text("Quick Learn")').first();
    const fullLearnButton = page.locator('button:has-text("Full Learn")').first();
    await expect(quickLearnButton).toBeVisible();
    await expect(fullLearnButton).toBeVisible();
    
    // Verify all experience levels are available
    const experienceOptions = await experienceLevelDropdown.locator('option').allTextContents();
    expect(experienceOptions).toContain('Beginner (Detailed with analogies)');
    expect(experienceOptions).toContain('Expert');
  });

  test('Beginner + Full Learn produces simple output with reasoning', async ({ page }) => {
    // Sample code for analysis
    const testCode = `
function processOrder(order) {
  if (!order.id) {
    throw new Error('Order ID required');
  }
  return saveToDatabase(order);
}`;

    // Select Beginner experience level
    const experienceLevelDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceLevelDropdown.selectOption({ label: 'Beginner (Detailed with analogies)' });
    
    // Select Full Learn
    const fullLearnButton = page.locator('button:has-text("Full Learn")').first();
    await fullLearnButton.click();
    
    // Paste code into editor
    const codeEditor = page.locator('textarea, .monaco-editor, [contenteditable="true"]').first();
    await codeEditor.click();
    await codeEditor.fill(testCode);
    
    // Click Analyze button
    const analyzeButton = page.locator('button:has-text("Analyze")').first();
    await analyzeButton.click();
    
    // Wait for analysis result
    await page.waitForSelector('.analysis-result, [class*="result"], [class*="output"]', { 
      timeout: 30000 // AI analysis can take time
    });
    
    // Verify output contains beginner-friendly language
    const resultContent = await page.locator('.analysis-result, [class*="result"], [class*="output"]').first().textContent();
    
    // Null check
    expect(resultContent).not.toBeNull();
    const resultText = resultContent || '';
    
    // Should contain simple analogies or explanations
    const hasSimpleLanguage = 
      resultText.toLowerCase().includes('like') ||
      resultText.toLowerCase().includes('similar to') ||
      resultText.toLowerCase().includes('think of') ||
      resultText.toLowerCase().includes('for example');
    
    expect(hasSimpleLanguage).toBeTruthy();
    
    // Should contain reasoning trace (Full Learn mode)
    const hasReasoningTrace = 
      resultText.toLowerCase().includes('reasoning') ||
      resultText.toLowerCase().includes('analysis_approach') ||
      resultText.toLowerCase().includes('key_observations');
    
    expect(hasReasoningTrace).toBeTruthy();
  });

  test('Expert + Quick Learn produces concise technical output', async ({ page }) => {
    const testCode = `
class OrderProcessor {
  constructor(private db: Database) {}
  
  async process(order: Order): Promise<void> {
    await this.validate(order);
    await this.db.save(order);
  }
}`;

    // Select Expert experience level
    const experienceLevelDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceLevelDropdown.selectOption({ label: 'Expert' });
    
    // Select Quick Learn (may already be selected by default)
    const quickLearnButton = page.locator('button:has-text("Quick Learn")').first();
    await quickLearnButton.click();
    
    // Paste code
    const codeEditor = page.locator('textarea, .monaco-editor, [contenteditable="true"]').first();
    await codeEditor.click();
    await codeEditor.fill(testCode);
    
    // Analyze
    const analyzeButton = page.locator('button:has-text("Analyze")').first();
    await analyzeButton.click();
    
    // Wait for result
    await page.waitForSelector('.analysis-result, [class*="result"], [class*="output"]', { 
      timeout: 30000
    });
    
    const resultContent = await page.locator('.analysis-result, [class*="result"], [class*="output"]').first().textContent();
    expect(resultContent).not.toBeNull();
    const resultText = resultContent || '';
    
    // Should NOT contain reasoning trace (Quick mode)
    const hasReasoningTrace = 
      resultText.toLowerCase().includes('reasoning_trace') ||
      resultText.toLowerCase().includes('analysis_approach');
    
    expect(hasReasoningTrace).toBeFalsy();
    
    // Output should be concise (arbitrary threshold: less than previous test)
    // Expert mode should produce shorter, more focused output
    expect(resultText.length).toBeLessThan(5000); // Arbitrary reasonable limit
  });

  test('mode changes persist across interactions', async ({ page }) => {
    // Set Expert + Full
    const experienceLevelDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceLevelDropdown.selectOption({ label: 'Expert' });
    
    const fullLearnButton = page.locator('button:has-text("Full Learn")').first();
    await fullLearnButton.click();
    
    // Paste code
    const testCode = 'function test() { return 42; }';
    const codeEditor = page.locator('textarea, .monaco-editor, [contenteditable="true"]').first();
    await codeEditor.click();
    await codeEditor.fill(testCode);
    
    // Analyze
    const analyzeButton = page.locator('button:has-text("Analyze")').first();
    await analyzeButton.click();
    
    await page.waitForSelector('.analysis-result, [class*="result"], [class*="output"]', { timeout: 30000 });
    
    // Clear and enter new code
    await codeEditor.click();
    await codeEditor.fill('const x = 100;');
    
    // Analyze again - modes should still be Expert + Full
    await analyzeButton.click();
    await page.waitForSelector('.analysis-result, [class*="result"], [class*="output"]', { timeout: 30000 });
    
    // Verify Expert is still selected
    const selectedExperience = await experienceLevelDropdown.inputValue();
    expect(selectedExperience).toContain('expert');
    
    // Verify Full Learn is still active
    const fullLearnActive = await fullLearnButton.evaluate((el) => 
      el.classList.contains('active') || 
      el.classList.contains('selected') ||
      el.getAttribute('aria-pressed') === 'true'
    );
    expect(fullLearnActive).toBeTruthy();
  });
});
