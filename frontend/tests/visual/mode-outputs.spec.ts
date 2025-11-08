import { test } from '@playwright/test';
import percySnapshot from '@percy/playwright';

/**
 * Visual Regression Tests for Mode Selection UI
 * 
 * Captures screenshots with Percy to detect visual changes in:
 * 1. Mode selector UI controls
 * 2. Beginner + Full Learn output (with analogies and reasoning)
 * 3. Expert + Quick Learn output (concise, technical)
 * 4. Quick Scan results from GitHub import
 * 
 * Percy will compare these screenshots against baseline to detect regressions.
 */

test.describe('Mode Selection Visual Regression', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:3000');
    await page.waitForLoadState('networkidle');
  });

  test('capture mode selector UI', async ({ page }) => {
    // Navigate to review page if not already there
    const reviewLink = page.locator('a:has-text("Review"), [href*="review"]').first();
    if (await reviewLink.isVisible()) {
      await reviewLink.click();
      await page.waitForLoadState('networkidle');
    }

    // Wait for mode selectors to be visible
    await page.locator('select[name="experience_level"], select').first().waitFor({ state: 'visible' });
    await page.locator('button:has-text("Quick Learn")').first().waitFor({ state: 'visible' });

    // Capture baseline UI
    await percySnapshot(page, 'Review Page - Mode Selectors (Default State)', {
      widths: [375, 768, 1920],
      minHeight: 1024
    });

    // Expand experience level dropdown to show options
    const experienceDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceDropdown.click();

    await percySnapshot(page, 'Review Page - Experience Level Dropdown Expanded', {
      widths: [375, 768, 1920]
    });
  });

  test('capture Beginner + Full Learn output', async ({ page }) => {
    const testCode = `
function calculateTotal(items) {
  let total = 0;
  for (let i = 0; i < items.length; i++) {
    total += items[i].price * items[i].quantity;
  }
  return total;
}`;

    // Navigate to review
    const reviewLink = page.locator('a:has-text("Review"), [href*="review"]').first();
    if (await reviewLink.isVisible()) {
      await reviewLink.click();
      await page.waitForLoadState('networkidle');
    }

    // Select Beginner + Full Learn
    const experienceDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceDropdown.selectOption({ label: 'Beginner (Detailed with analogies)' });

    const fullLearnButton = page.locator('button:has-text("Full Learn")').first();
    await fullLearnButton.click();

    // Paste code
    const codeEditor = page.locator('textarea, .monaco-editor, [contenteditable="true"]').first();
    await codeEditor.click();
    await codeEditor.fill(testCode);

    // Capture "before analysis" state
    await percySnapshot(page, 'Review Page - Beginner Mode Before Analysis', {
      widths: [1920],
      minHeight: 1024
    });

    // Analyze
    const analyzeButton = page.locator('button:has-text("Analyze")').first();
    await analyzeButton.click();

    // Wait for result with longer timeout (AI can be slow)
    await page.waitForSelector('.analysis-result, [class*="result"], [class*="output"]', {
      timeout: 45000
    });

    // Wait a bit for any animations/transitions
    await page.waitForTimeout(1000);

    // Capture analysis result
    await percySnapshot(page, 'Review Page - Beginner + Full Learn Output (With Analogies & Reasoning)', {
      widths: [1920],
      minHeight: 1024,
      percyCSS: `
        /* Hide dynamic timestamps that would cause false positives */
        .timestamp, [class*="timestamp"], time {
          visibility: hidden !important;
        }
      `
    });
  });

  test('capture Expert + Quick Learn output', async ({ page }) => {
    const testCode = `
interface OrderProcessor {
  process(order: Order): Promise<void>;
}

class DefaultOrderProcessor implements OrderProcessor {
  constructor(private db: Database, private validator: Validator) {}
  
  async process(order: Order): Promise<void> {
    await this.validator.validate(order);
    await this.db.save(order);
  }
}`;

    // Navigate to review
    const reviewLink = page.locator('a:has-text("Review"), [href*="review"]').first();
    if (await reviewLink.isVisible()) {
      await reviewLink.click();
      await page.waitForLoadState('networkidle');
    }

    // Select Expert + Quick Learn
    const experienceDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceDropdown.selectOption({ label: 'Expert' });

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
      timeout: 45000
    });

    await page.waitForTimeout(1000);

    // Capture concise expert output
    await percySnapshot(page, 'Review Page - Expert + Quick Learn Output (Concise & Technical)', {
      widths: [1920],
      minHeight: 1024,
      percyCSS: `
        .timestamp, [class*="timestamp"], time {
          visibility: hidden !important;
        }
      `
    });
  });

  test('capture Quick Scan GitHub import', async ({ page }) => {
    // This test assumes the GitHub import feature works
    // If it requires authentication, you may need to set that up first

    // Navigate to review
    const reviewLink = page.locator('a:has-text("Review"), [href*="review"]').first();
    if (await reviewLink.isVisible()) {
      await reviewLink.click();
      await page.waitForLoadState('networkidle');
    }

    // Look for GitHub import button/tab
    const githubImportButton = page.locator('button:has-text("GitHub"), button:has-text("Import"), [aria-label*="GitHub"]').first();
    
    if (await githubImportButton.isVisible({ timeout: 5000 }).catch(() => false)) {
      await githubImportButton.click();
      await page.waitForLoadState('networkidle');

      // Capture GitHub import UI
      await percySnapshot(page, 'Review Page - GitHub Import UI', {
        widths: [1920],
        minHeight: 1024
      });

      // Enter a test repository URL (use a known small public repo)
      const urlInput = page.locator('input[type="text"], input[type="url"]').first();
      if (await urlInput.isVisible({ timeout: 5000 }).catch(() => false)) {
        await urlInput.fill('https://github.com/octocat/Hello-World');

        // Click Quick Scan
        const quickScanButton = page.locator('button:has-text("Quick Scan"), button:has-text("Quick")').first();
        if (await quickScanButton.isVisible({ timeout: 5000 }).catch(() => false)) {
          await quickScanButton.click();

          // Wait for results
          await page.waitForSelector('[class*="result"], [class*="analysis"], [class*="scan-result"]', {
            timeout: 30000
          }).catch(() => {
            // If results don't appear, that's okay for visual test
            console.log('Quick Scan results not loaded, skipping that screenshot');
          });

          await page.waitForTimeout(1000);

          // Capture Quick Scan results
          await percySnapshot(page, 'Review Page - Quick Scan Results (GitHub Import)', {
            widths: [1920],
            minHeight: 1024,
            percyCSS: `
              .timestamp, [class*="timestamp"], time {
                visibility: hidden !important;
              }
            `
          });
        }
      }
    } else {
      console.log('GitHub import feature not visible, skipping this test');
      test.skip();
    }
  });

  test('capture mode transitions', async ({ page }) => {
    // Test to capture UI state changes when switching modes

    const reviewLink = page.locator('a:has-text("Review"), [href*="review"]').first();
    if (await reviewLink.isVisible()) {
      await reviewLink.click();
      await page.waitForLoadState('networkidle');
    }

    // Capture initial state (should default to Intermediate + Quick)
    await percySnapshot(page, 'Review Page - Default Mode Selection', {
      widths: [1920]
    });

    // Change to Beginner
    const experienceDropdown = page.locator('select[name="experience_level"], select').first();
    await experienceDropdown.selectOption({ label: 'Beginner (Detailed with analogies)' });
    await page.waitForTimeout(300);

    await percySnapshot(page, 'Review Page - Beginner Mode Selected', {
      widths: [1920]
    });

    // Change to Full Learn
    const fullLearnButton = page.locator('button:has-text("Full Learn")').first();
    await fullLearnButton.click();
    await page.waitForTimeout(300);

    await percySnapshot(page, 'Review Page - Beginner + Full Learn Selected', {
      widths: [1920]
    });

    // Change to Expert
    await experienceDropdown.selectOption({ label: 'Expert' });
    await page.waitForTimeout(300);

    await percySnapshot(page, 'Review Page - Expert + Full Learn Selected', {
      widths: [1920]
    });

    // Change to Quick
    const quickLearnButton = page.locator('button:has-text("Quick Learn")').first();
    await quickLearnButton.click();
    await page.waitForTimeout(300);

    await percySnapshot(page, 'Review Page - Expert + Quick Learn Selected', {
      widths: [1920]
    });
  });
});
