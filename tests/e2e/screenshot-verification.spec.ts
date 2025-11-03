import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Create screenshots directory
const screenshotDir = '/tmp/devsmith-screenshots';
if (!fs.existsSync(screenshotDir)) {
  fs.mkdirSync(screenshotDir, { recursive: true });
}

test.describe('Review App - Screenshot Verification', () => {
  test('1. Capture current UI at localhost:3000', async ({ page }) => {
    console.log('📸 Taking screenshot of localhost:3000 root');
    
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
    await page.screenshot({ 
      path: path.join(screenshotDir, '01-root-redirect.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/01-root-redirect.png');
  });

  test('2. Capture Review landing page', async ({ page }) => {
    console.log('📸 Taking screenshot of /review page');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    await page.screenshot({ 
      path: path.join(screenshotDir, '02-review-landing.png'),
      fullPage: true 
    });
    
    // Document what's visible
    const heading = await page.locator('h1, h2').first().textContent();
    const buttonCount = await page.locator('button').count();
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/02-review-landing.png');
    console.log(`   Heading: ${heading}`);
    console.log(`   Button count: ${buttonCount}`);
  });

  test('3. Test code input interaction', async ({ page }) => {
    console.log('📸 Testing code input interaction');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    
    // Find and fill code input
    const codeInput = page.locator('textarea').first();
    await codeInput.fill('package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello")\n}');
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '03-code-input-filled.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/03-code-input-filled.png');
  });

  test('4. Test Preview Mode button click', async ({ page }) => {
    console.log('📸 Testing Preview Mode button click');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    
    // Fill code
    const codeInput = page.locator('textarea').first();
    await codeInput.fill('package main\n\nfunc main() {}');
    
    // Click Preview Mode button
    const previewButton = page.getByRole('button', { name: /Preview Mode/i }).first();
    await previewButton.click();
    
    // Wait 2 seconds for any visual feedback
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '04-preview-button-clicked.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/04-preview-button-clicked.png');
  });

  test('5. Wait for analysis result', async ({ page }) => {
    console.log('📸 Waiting for AI analysis result');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    
    // Fill code
    const codeInput = page.locator('textarea').first();
    await codeInput.fill('package main\n\nfunc main() {}');
    
    // Click Preview Mode
    const previewButton = page.getByRole('button', { name: /Preview Mode/i }).first();
    await previewButton.click();
    
    // Wait for result or timeout (10 seconds)
    try {
      await page.waitForSelector('text=/summary|file_tree|Preview Mode Analysis/', { timeout: 10000 });
      console.log('✅ Analysis result appeared');
    } catch (e) {
      console.log('❌ Analysis result did not appear within 10 seconds');
    }
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '05-analysis-result.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/05-analysis-result.png');
  });
});
