import { test, expect } from '../fixtures/auth.fixture';

const TEST_CODE = `function calculateSum(arr) {
  let sum = 0;
  for (let i = 0; i < arr.length; i++) {
    sum += arr[i];
  }
  return sum;
}`;

// Test matrix: 5 modes × 3 experience levels = 15 tests
const modes = ['preview', 'skim', 'scan', 'detailed', 'critical'];
const experienceLevels = ['beginner', 'intermediate', 'expert'];

test.describe('Review Service - Form Submission (All Combinations)', () => {
  test.beforeEach(async ({ authenticatedPage: page }) => {
    // Navigate and wait for page to load
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    await page.waitForSelector('.analysis-mode-selector', { timeout: 15000 });
    await page.waitForSelector('.code-editor-container', { timeout: 15000 });
  });

  for (const mode of modes) {
    for (const experience of experienceLevels) {
      test(`${mode.charAt(0).toUpperCase() + mode.slice(1)} Mode - ${experience.charAt(0).toUpperCase() + experience.slice(1)}`, async ({ authenticatedPage: page }) => {
        // Fill code
        await page.click('.monaco-editor');
        await page.waitForTimeout(300);
        await page.evaluate((code: string) => {
          const textarea = document.querySelector('textarea.ime-text-area') as HTMLTextAreaElement;
          if (textarea) {
            textarea.value = code;
            textarea.dispatchEvent(new Event('input', { bubbles: true }));
          }
        }, TEST_CODE);

        // Wait for model to load
        await page.waitForSelector('#model-select', { timeout: 10000 });
        await page.waitForTimeout(2000);
        
        // Select model if available
        const modelDisabled = await page.locator('#model-select').getAttribute('disabled');
        if (modelDisabled === null) {
          await page.selectOption('#model-select', { index: 0 });
        }

        // Select mode
        await page.click(`.mode-card.${mode}`);

        // Select experience level
        const experienceSelect = page.locator('select.form-select').nth(1);
        await experienceSelect.selectOption(experience);

        // Select output format (quick)
        await page.click('label[for="outputQuick"]');

        // Click Analyze
        const analyzeButton = page.locator('button:has-text("Analyze Code")');
        await expect(analyzeButton).toBeEnabled({ timeout: 5000 });
        await analyzeButton.click();

        // Verify loading state appears
        await page.waitForSelector('button:has-text("Analyzing")', { timeout: 5000 });

        // Take screenshot of submitted state
        await page.screenshot({ 
          path: `/tmp/review-${mode}-${experience}-submitted.png`,
          fullPage: false
        });

        console.log(`✓ ${mode} mode with ${experience} experience level - form submitted successfully`);
      });
    }
  }
});
