import { test, expect } from '@playwright/test';

test.describe('VERIFY: Review App Actually Works', () => {
  test('Can paste code and run Preview mode analysis', async ({ page }) => {
    // Go to review page
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Paste test code
    const testCode = `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`;
    
    await page.fill('textarea[name="pasted_code"]', testCode);
    
    // Click Preview mode button
    const previewButton = page.locator('button:has-text("Select Preview")').first();
    await previewButton.click();
    
    // Wait for results to appear (HTMX should swap content into #reading-mode-demo)
    await page.waitForTimeout(5000); // Give Ollama time to respond
    
    // Check if results container has content
    const resultsContainer = page.locator('#reading-mode-demo');
    const content = await resultsContainer.textContent();
    
    console.log('Results container content:', content);
    
    // Should contain analysis results
    expect(content).toBeTruthy();
    expect(content!.length).toBeGreaterThan(100); // Should have actual analysis text
  });
  
  test('Dark mode toggle actually works', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Get initial state
    const html = page.locator('html');
    const initialClass = await html.getAttribute('class');
    
    // Click dark mode toggle
    const darkModeButton = page.locator('#dark-mode-toggle');
    await darkModeButton.click();
    
    // Wait for Alpine.js to update
    await page.waitForTimeout(500);
    
    // Check class changed
    const newClass = await html.getAttribute('class');
    
    console.log('Initial class:', initialClass);
    console.log('New class:', newClass);
    
    // Either dark was added or removed
    expect(initialClass).not.toBe(newClass);
  });
  
  test('Navigation links work', async ({ page }) => {
    await page.goto('http://localhost:3000/review', { waitUntil: 'domcontentloaded' });
    
    // Click on Logs link
    const logsLink = page.locator('a:has-text("Logs")').first();
    const href = await logsLink.getAttribute('href');
    
    console.log('Logs link href:', href);
    
    await logsLink.click();
    await page.waitForLoadState('domcontentloaded');
    
    // Should navigate to logs page
    expect(page.url()).toContain('logs');
  });
});


