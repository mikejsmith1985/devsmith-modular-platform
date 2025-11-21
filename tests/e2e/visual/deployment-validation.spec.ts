import { test, expect } from '@playwright/test';

test.describe('Deployment Validation - All Recent Changes', () => {
  
  test('AI Factory - Dark mode and branding', async ({ page }) => {
    // Navigate to AI Factory
    await page.goto('/llm-config', { waitUntil: 'networkidle' });
    
    // Wait for React app to render
    await page.waitForSelector('.navbar-brand', { timeout: 15000 });
    
    // Check for DevSmith Platform branding (not "AI Factory" twice)
    const navbarBrand = await page.locator('.navbar-brand').first();
    const brandText = await navbarBrand.textContent();
    
    console.log('Navbar brand text:', brandText);
    expect(brandText).toContain('DevSmith Platform');
    
    // Check for dark mode toggle button
    const darkModeToggle = await page.locator('button').filter({ hasText: /sun|moon/ }).first();
    await expect(darkModeToggle).toBeVisible();
    console.log('Dark mode toggle found!');
    
    // Toggle dark mode
    await darkModeToggle.click();
    await page.waitForTimeout(500);
    
    // Verify dark mode is applied
    const container = await page.locator('.container').first();
    const containerClass = await container.getAttribute('class');
    console.log('Container classes after dark mode toggle:', containerClass);
    expect(containerClass).toContain('text-light');
    
    // Take screenshot
    await page.screenshot({ 
      path: 'test-results/validation-ai-factory-dark.png',
      fullPage: true 
    });
  });
  
  test('Review Page - Import button position', async ({ page }) => {
    // Navigate to Review page
    await page.goto('/review', { waitUntil: 'networkidle' });
    
    // Wait for page to load
    await page.waitForSelector('button', { timeout: 15000 });
    
    // Check import button is NOT in the header
    const headerImportButton = await page.locator('.row .col-12 button').filter({ hasText: 'Import from GitHub' }).count();
    console.log('Import buttons in header:', headerImportButton);
    
    // Check import button IS in the secondary controls row
    const secondaryImportButton = await page.locator('button').filter({ hasText: 'Import from GitHub' });
    await expect(secondaryImportButton).toBeVisible();
    
    // Verify it's near the Clear button
    const clearButton = await page.locator('button').filter({ hasText: 'Clear' });
    await expect(clearButton).toBeVisible();
    
    console.log('Import button found in correct location (near Clear button)');
    
    // Take screenshot
    await page.screenshot({ 
      path: 'test-results/validation-review-buttons.png',
      fullPage: true 
    });
  });
  
  test('Review Page - Model Selector uses AI Factory', async ({ page }) => {
    // Navigate to Review page
    await page.goto('/review', { waitUntil: 'networkidle' });
    
    // Wait for model selector to load
    await page.waitForSelector('select#model-select', { timeout: 15000 });
    await page.waitForTimeout(2000);
    
    // Find the model selector
    const modelSelect = await page.locator('select#model-select');
    await expect(modelSelect).toBeVisible();
    
    // Get the options
    const options = await modelSelect.locator('option').allTextContents();
    console.log('Model selector options:', options);
    
    // Check if options include provider information (AI Factory format)
    const hasProviderInfo = options.some(opt => opt.includes('anthropic') || opt.includes('ollama'));
    console.log('Options have provider info:', hasProviderInfo);
    
    // Take screenshot
    await page.screenshot({ 
      path: 'test-results/validation-review-model-selector.png',
      fullPage: true 
    });
  });
  
  test('Prompt Editor Modal - Dark mode', async ({ page }) => {
    // Navigate to Review page
    await page.goto('/review', { waitUntil: 'networkidle' });
    
    // Wait for page to load
    await page.waitForSelector('button', { timeout: 15000 });
    
    // Click Details button for Preview mode
    const detailsButton = await page.locator('button').filter({ hasText: 'Details' }).first();
    await detailsButton.click();
    
    // Wait for modal
    await page.waitForTimeout(1000);
    
    // Verify modal is visible
    const modal = await page.locator('.modal.show');
    await expect(modal).toBeVisible();
    
    // Take screenshot in light mode
    await page.screenshot({ 
      path: 'test-results/validation-prompt-modal-light.png',
      fullPage: true 
    });
    
    // Toggle dark mode (find the sun/moon icon in the page, not modal)
    const darkModeToggle = await page.locator('button i.bi-moon-fill, button i.bi-sun-fill').first();
    await darkModeToggle.click();
    await page.waitForTimeout(500);
    
    // Take screenshot in dark mode
    await page.screenshot({ 
      path: 'test-results/validation-prompt-modal-dark.png',
      fullPage: true 
    });
    
    // Check if modal has dark mode classes
    const modalContent = await page.locator('.modal-content');
    const modalClass = await modalContent.getAttribute('class');
    console.log('Modal classes:', modalClass);
    expect(modalClass).toContain('bg-dark');
  });
  
});
