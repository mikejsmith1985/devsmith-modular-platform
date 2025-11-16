import { test, expect } from '@playwright/test';

test.describe('LLM Config - Visual Verification', () => {
  test('authenticate and view LLM config page', async ({ page, context }) => {
    // Step 1: Get auth token via API request
    console.log('Step 1: Getting auth token...');
    const authResponse = await context.request.post('http://localhost:3000/auth/test-login', {
      data: {
        username: 'visual-test',
        email: 'visual@devsmith.local',
        avatar_url: 'https://example.com/avatar.png',
        github_id: '77777'
      }
    });
    
    expect(authResponse.ok()).toBeTruthy();
    const headers = authResponse.headers();
    const setCookie = headers['set-cookie'];
    console.log('Set-Cookie header:', setCookie);
    
    // Extract token from Set-Cookie header
    const tokenMatch = setCookie.match(/devsmith_token=([^;]+)/);
    if (!tokenMatch) {
      throw new Error('No token found in Set-Cookie header');
    }
    
    // Set the cookie in the browser context
    await context.addCookies([{
      name: 'devsmith_token',
      value: tokenMatch[1],
      domain: 'localhost',
      path: '/',
      httpOnly: true
    }]);
    
    console.log('Cookie set successfully');
    
    // Step 2: Navigate to LLM config page
    console.log('Step 2: Navigating to /llm-config...');
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForTimeout(2000); // Wait for page to fully load
    
    // Take screenshot
    await page.screenshot({ path: 'test-results/llm-config-visual-verification.png', fullPage: true });
    console.log('Screenshot saved to: test-results/llm-config-visual-verification.png');
    
    // Check what's actually on the page
    const pageContent = await page.content();
    console.log('Page URL:', page.url());
    console.log('Page title:', await page.title());
    
    // Check if we're redirected to login
    if (page.url().includes('/login')) {
      console.error('❌ STILL REDIRECTED TO LOGIN - Authentication failed!');
      throw new Error('Authentication did not work - redirected to login');
    }
    
    // Check for key elements
    const hasH2 = await page.locator('h2').count();
    console.log('Number of <h2> elements:', hasH2);
    
    const h2Text = hasH2 > 0 ? await page.locator('h2').first().textContent() : 'none';
    console.log('First <h2> text:', h2Text);
    
    // Check for error messages
    const errorAlert = await page.locator('.alert-danger').count();
    if (errorAlert > 0) {
      const errorText = await page.locator('.alert-danger').textContent();
      console.log('❌ Error on page:', errorText);
    }
    
    // Verify we're on the right page
    expect(page.url()).toContain('/llm-config');
    console.log('✓ Successfully loaded /llm-config page');
  });

  test('create Ollama config via UI', async ({ page, context }) => {
    // Get auth token
    const authResponse = await context.request.post('http://localhost:3000/auth/test-login', {
      data: {
        username: 'ollama-test',
        email: 'ollama@devsmith.local',
        avatar_url: 'https://example.com/avatar.png',
        github_id: '66666'
      }
    });
    
    const headers = authResponse.headers();
    const setCookie = headers['set-cookie'];
    const tokenMatch = setCookie.match(/devsmith_token=([^;]+)/);
    
    await context.addCookies([{
      name: 'devsmith_token',
      value: tokenMatch![1],
      domain: 'localhost',
      path: '/',
      httpOnly: true
    }]);
    
    // Navigate to page
    await page.goto('http://localhost:3000/llm-config');
    await page.waitForTimeout(1000);
    
    console.log('Looking for Add button...');
    // Find and click the Add button (could be various texts)
    const addButtonSelectors = [
      'button:has-text("Add AI Model")',
      'button:has-text("+ Add")',
      'button:has-text("Add Model")',
      'button:has-text("Add")'
    ];
    
    let clicked = false;
    for (const selector of addButtonSelectors) {
      const count = await page.locator(selector).count();
      if (count > 0) {
        console.log(`Found button with selector: ${selector}`);
        await page.click(selector);
        clicked = true;
        break;
      }
    }
    
    if (!clicked) {
      console.error('Could not find Add button');
      await page.screenshot({ path: 'test-results/no-add-button.png', fullPage: true });
      throw new Error('Add button not found');
    }
    
    // Wait for modal
    console.log('Waiting for modal...');
    await page.waitForSelector('text=Add AI Model Configuration', { timeout: 5000 });
    await page.screenshot({ path: 'test-results/modal-opened.png', fullPage: true });
    
    // Fill form
    console.log('Filling form...');
    await page.fill('input[name="name"]', 'Visual Test Ollama');
    await page.selectOption('select[name="provider"]', 'ollama');
    await page.waitForTimeout(500);
    await page.selectOption('select[name="model"]', 'deepseek-coder-v2:16b');
    
    await page.screenshot({ path: 'test-results/form-filled.png', fullPage: true });
    
    // Click save
    console.log('Clicking Save...');
    await page.click('button:has-text("Save")');
    
    // Wait for modal to close
    await page.waitForSelector('text=Add AI Model Configuration', { state: 'hidden', timeout: 5000 });
    await page.waitForTimeout(1000);
    
    await page.screenshot({ path: 'test-results/config-saved.png', fullPage: true });
    
    // Verify config appears
    await expect(page.locator('text=deepseek-coder-v2:16b')).toBeVisible();
    console.log('✓ Ollama config created successfully!');
  });
});
