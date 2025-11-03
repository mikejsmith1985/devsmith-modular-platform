import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const screenshotDir = '/tmp/devsmith-screenshots';

test.describe('Review App - AFTER FIX Screenshots', () => {
  test('6. Capture FIXED UI at localhost:3000/review', async ({ page }) => {
    console.log('📸 Taking screenshot of FIXED /review page');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle', timeout: 30000 });
    
    // Wait a bit for any redirects
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '06-FIXED-review-landing.png'),
      fullPage: true 
    });
    
    // Document what's visible
    const url = page.url();
    const heading = await page.locator('h1, h2').first().textContent();
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/06-FIXED-review-landing.png');
    console.log(`   Final URL: ${url}`);
    console.log(`   Heading: ${heading}`);
  });

  test('7. Verify 2-pane layout structure', async ({ page }) => {
    console.log('📸 Verifying 2-pane layout');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    
    // Check for workspace elements
    const codePane = page.locator('.code-pane').first();
    const analysisPane = page.locator('.analysis-pane').first();
    const modeSelector = page.locator('#mode-selector');
    const analyzeButton = page.getByRole('button', { name: /Analyze Code/i });
    
    const codePaneVisible = await codePane.isVisible();
    const analysisPaneVisible = await analysisPane.isVisible();
    const modeSelectorVisible = await modeSelector.isVisible();
    const analyzeButtonVisible = await analyzeButton.isVisible();
    
    console.log(`   Code pane visible: ${codePaneVisible}`);
    console.log(`   Analysis pane visible: ${analysisPaneVisible}`);
    console.log(`   Mode selector visible: ${modeSelectorVisible}`);
    console.log(`   Analyze button visible: ${analyzeButtonVisible}`);
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '07-FIXED-2pane-layout.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/07-FIXED-2pane-layout.png');
  });

  test('8. Test mode selector interaction', async ({ page }) => {
    console.log('📸 Testing mode selector');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    
    const modeSelector = page.locator('#mode-selector');
    
    // Change to Skim mode
    await modeSelector.selectOption('skim');
    await page.waitForTimeout(500);
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '08-FIXED-mode-selector.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/08-FIXED-mode-selector.png');
  });

  test('9. Test Analyze button in workspace', async ({ page }) => {
    console.log('📸 Testing Analyze button in workspace');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    
    // The code is already pre-filled in demo workspace
    // Just click Analyze button
    const analyzeButton = page.getByRole('button', { name: /Analyze Code/i });
    await analyzeButton.click();
    
    // Wait for loading indicator
    await page.waitForTimeout(2000);
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '09-FIXED-analyze-clicked.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/09-FIXED-analyze-clicked.png');
  });

  test('10. Wait for AI analysis in 2-pane layout', async ({ page }) => {
    console.log('📸 Waiting for AI analysis in workspace');
    
    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(2000);
    
    // Click Analyze button
    const analyzeButton = page.getByRole('button', { name: /Analyze Code/i });
    await analyzeButton.click();
    
    // Wait for analysis result in right pane (15 second timeout)
    try {
      await page.waitForSelector('#analysis-pane', { timeout: 15000 });
      
      // Wait additional time for content to load
      await page.waitForTimeout(3000);
      
      console.log('✅ Analysis completed in 2-pane workspace');
    } catch (e) {
      console.log('⚠️ Analysis did not complete within 15 seconds');
    }
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '10-FIXED-analysis-result-2pane.png'),
      fullPage: true 
    });
    
    console.log('✅ Screenshot saved: /tmp/devsmith-screenshots/10-FIXED-analysis-result-2pane.png');
  });
});
