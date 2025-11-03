import { test, expect } from '@playwright/test';

test.describe('Review App - Basic Smoke Test', () => {
  test('should load the Review app homepage', async ({ page }) => {
    // Navigate to root (should redirect to /review)
    await page.goto('http://localhost:3000');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Verify we're on the Review page
    await expect(page).toHaveTitle(/DevSmith Review/);
    
    // Verify the main heading is present
    const heading = page.locator('h1, h2').filter({ hasText: /DevSmith Review|Code Analysis/ });
    await expect(heading).toBeVisible();
    
    // Verify all 5 reading mode buttons are present
    await expect(page.getByRole('button', { name: /Preview Mode/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Skim Mode/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Scan Mode/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Detailed Mode/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Critical Mode/i })).toBeVisible();
    
    // Verify code input textarea exists
    const codeInput = page.locator('textarea[name="pasted_code"], textarea[placeholder*="code"]');
    await expect(codeInput).toBeVisible();
  });

  test('should successfully analyze code in Preview Mode', async ({ page }) => {
    // Navigate to Review app
    await page.goto('http://localhost:3000/review');
    await page.waitForLoadState('networkidle');
    
    // Find and fill the code input
    const codeInput = page.locator('textarea').first();
    await codeInput.fill('package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello World")\n}');
    
    // Click Preview Mode button
    await page.getByRole('button', { name: /Preview Mode|preview/i }).first().click();
    
    // Wait for analysis to complete (up to 15 seconds)
    await page.waitForSelector('text=/summary|file_tree|tech_stack/', { timeout: 15000 });
    
    // Verify analysis results are displayed
    const analysisResult = page.locator('text=/Preview Mode Analysis|summary|file_tree/');
    await expect(analysisResult).toBeVisible();
  });
});
