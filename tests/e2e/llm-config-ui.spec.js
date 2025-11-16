const { test, expect } = require('@playwright/test');

test.describe('LLM Configuration UI', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate by calling test-login endpoint
    const authResponse = await page.request.post('http://localhost:3000/auth/test-login', {
      data: {
        username: 'playwright-test',
        email: 'playwright@devsmith.local',
        avatar_url: 'https://example.com/avatar.png',
        github_id: '99999'
      }
    });
    
    expect(authResponse.ok()).toBeTruthy();
    const authData = await authResponse.json();
    expect(authData.message).toBe('success');
    
    // Extract and set the authentication cookie
    const cookies = await authResponse.headersArray();
    const setCookieHeader = cookies.find(h => h.name.toLowerCase() === 'set-cookie');
    if (setCookieHeader) {
      const cookieMatch = setCookieHeader.value.match(/devsmith_token=([^;]+)/);
      if (cookieMatch) {
        await page.context().addCookies([{
          name: 'devsmith_token',
          value: cookieMatch[1],
          domain: 'localhost',
          path: '/'
        }]);
      }
    }
  });

  test('should load LLM config page and display existing configs', async ({ page }) => {
    // Navigate to LLM config page
    await page.goto('http://localhost:3000/llm-config');
    
    // Wait for page to load
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Take screenshot of initial state
    await page.screenshot({ path: 'test-results/llm-config-page-load.png', fullPage: true });
    
    // Verify page title
    await expect(page.locator('text=AI Model Management')).toBeVisible();
    
    console.log('✓ LLM Config page loaded successfully');
  });

  test('should open Add AI Model modal', async ({ page }) => {
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Click Add AI Model button
    await page.click('button:has-text("Add AI Model"), button:has-text("+ Add")');
    
    // Wait for modal to appear
    await page.waitForSelector('text=Add AI Model Configuration', { timeout: 5000 });
    
    // Take screenshot of modal
    await page.screenshot({ path: 'test-results/llm-config-modal-open.png', fullPage: true });
    
    // Verify modal fields
    await expect(page.locator('label:has-text("Configuration Name")')).toBeVisible();
    await expect(page.locator('label:has-text("Provider")')).toBeVisible();
    await expect(page.locator('label:has-text("Model")')).toBeVisible();
    
    console.log('✓ Add AI Model modal opened successfully');
  });

  test('should create Ollama configuration successfully', async ({ page }) => {
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Open modal
    await page.click('button:has-text("Add AI Model"), button:has-text("+ Add")');
    await page.waitForSelector('text=Add AI Model Configuration');
    
    // Fill in form for Ollama
    await page.fill('input[name="name"]', 'Playwright Test Ollama');
    await page.selectOption('select[name="provider"]', 'ollama');
    
    // Wait for model dropdown to populate
    await page.waitForTimeout(500);
    await page.selectOption('select[name="model"]', 'deepseek-coder-v2:16b');
    
    // Take screenshot before saving
    await page.screenshot({ path: 'test-results/llm-config-ollama-filled.png', fullPage: true });
    
    // Click Save button
    await page.click('button:has-text("Save")');
    
    // Wait for modal to close and config to appear in list
    await page.waitForSelector('text=Add AI Model Configuration', { state: 'hidden', timeout: 5000 });
    
    // Wait a bit for the list to refresh
    await page.waitForTimeout(1000);
    
    // Take screenshot of result
    await page.screenshot({ path: 'test-results/llm-config-ollama-saved.png', fullPage: true });
    
    // Verify the config appears in the list
    await expect(page.locator('text=deepseek-coder-v2:16b')).toBeVisible();
    
    console.log('✓ Ollama configuration created successfully');
  });

  test('should create Claude configuration with API key', async ({ page }) => {
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Open modal
    await page.click('button:has-text("Add AI Model"), button:has-text("+ Add")');
    await page.waitForSelector('text=Add AI Model Configuration');
    
    // Fill in form for Claude
    await page.fill('input[name="name"]', 'Playwright Test Claude');
    await page.selectOption('select[name="provider"]', 'anthropic');
    
    // Wait for model dropdown to populate
    await page.waitForTimeout(500);
    await page.selectOption('select[name="model"]', 'claude-3-5-sonnet-20241022');
    
    // Fill in API key
    await page.fill('input[name="api_key"]', 'sk-test-playwright-fake-key-12345678901234567890');
    
    // Take screenshot before saving
    await page.screenshot({ path: 'test-results/llm-config-claude-filled.png', fullPage: true });
    
    // Click Save button
    await page.click('button:has-text("Save")');
    
    // Wait for modal to close
    await page.waitForSelector('text=Add AI Model Configuration', { state: 'hidden', timeout: 5000 });
    
    // Wait for list to refresh
    await page.waitForTimeout(1000);
    
    // Take screenshot of result
    await page.screenshot({ path: 'test-results/llm-config-claude-saved.png', fullPage: true });
    
    // Verify the config appears in the list
    await expect(page.locator('text=claude-3-5-sonnet-20241022')).toBeVisible();
    
    console.log('✓ Claude configuration created successfully');
  });

  test('should expand Advanced Settings dropdown', async ({ page }) => {
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Open modal
    await page.click('button:has-text("Add AI Model"), button:has-text("+ Add")');
    await page.waitForSelector('text=Add AI Model Configuration');
    
    // Click Advanced Settings
    await page.click('button:has-text("Advanced Settings")');
    
    // Wait for settings to expand
    await page.waitForTimeout(300);
    
    // Take screenshot
    await page.screenshot({ path: 'test-results/llm-config-advanced-settings.png', fullPage: true });
    
    // Verify "Set as Default" checkbox is visible
    await expect(page.locator('label:has-text("Set as Default")')).toBeVisible();
    
    console.log('✓ Advanced Settings dropdown works');
  });

  test('should show validation errors for missing required fields', async ({ page }) => {
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForSelector('text=AI Model Management', { timeout: 10000 });
    
    // Open modal
    await page.click('button:has-text("Add AI Model"), button:has-text("+ Add")');
    await page.waitForSelector('text=Add AI Model Configuration');
    
    // Try to save without filling anything
    const saveButton = page.locator('button:has-text("Save")');
    
    // Check if save button is disabled (it should be due to isFormValid check)
    const isDisabled = await saveButton.isDisabled();
    expect(isDisabled).toBeTruthy();
    
    // Take screenshot
    await page.screenshot({ path: 'test-results/llm-config-validation.png', fullPage: true });
    
    console.log('✓ Form validation prevents saving incomplete forms');
  });
});
