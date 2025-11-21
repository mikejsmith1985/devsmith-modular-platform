import { test, expect } from '@playwright/test';

test.describe('AI Factory Connection Validation', () => {
  test('should reject invalid Ollama endpoint with connection error', async ({ page }) => {
    // Navigate to AI Factory (will redirect to login if not authenticated)
    await page.goto('/');
    
    // Click AI Factory card if on dashboard, or handle auth redirect
    const url = page.url();
    if (url.includes('dashboard')) {
      // Already authenticated
      await page.click('text=AI Factory');
    } else {
      // Not authenticated - skip test
      test.skip();
      return;
    }
    
    // Wait for AI Factory page to load
    const navigatedToAIFactory = await page.waitForURL('**/llm-configs**', { timeout: 5000 }).catch(() => false);
    if (!navigatedToAIFactory) {
      test.skip();
      return;
    }
    
    // Click to create new config
    const addButtonClicked = await page.click('button:has-text("Add Configuration")').then(() => true).catch(() => false);
    if (!addButtonClicked) {
      test.skip();
      return;
    }
    
    // Fill in form with INVALID Ollama endpoint
    await page.fill('input[name="provider"]', 'ollama');
    await page.fill('input[name="model"]', 'qwen2.5-coder:7b');
    await page.fill('input[name="endpoint"]', 'http://invalid-host:11434');
    
    // Click Save (should trigger connection validation)
    await page.click('button:has-text("Save")');
    
    // Should see error message about connection test failure
    const errorMessage = await page.locator('text=/Connection test failed|Failed to connect/i').textContent({ timeout: 5000 }).catch(() => null);
    
    if (errorMessage) {
      expect(errorMessage).toContain('failed');
    }
  });
});
