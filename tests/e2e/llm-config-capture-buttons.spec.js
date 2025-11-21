const { test, expect } = require('@playwright/test');

test.describe('LLM Config - Button Capture', () => {
  test.beforeEach(async ({ page }) => {
    // Authenticate by calling test-login endpoint
    const authResponse = await page.request.post('/auth/test-login', {
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
    
    // Extract JWT token from cookie
    const cookies = await authResponse.headersArray();
    const setCookieHeader = cookies.find(h => h.name.toLowerCase() === 'set-cookie');
    let jwtToken = null;
    
    if (setCookieHeader) {
      const cookieMatch = setCookieHeader.value.match(/devsmith_token=([^;]+)/);
      if (cookieMatch) {
        jwtToken = cookieMatch[1];
        
        // Set cookie in browser
        await page.context().addCookies([{
          name: 'devsmith_token',
          value: jwtToken,
          domain: 'localhost',
          path: '/'
        }]);
      }
    }
    
    // Navigate to a page to set localStorage (React app needs token in localStorage)
    await page.goto('/portal');
    
    // Inject JWT token into localStorage for React auth context
    if (jwtToken) {
      await page.evaluate((token) => {
        localStorage.setItem('devsmith_token', token);
      }, jwtToken);
    }
  });

  test('should capture all buttons on LLM config page', async ({ page }) => {
    // Navigate to LLM config page
    await page.goto('/llm-config');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // Get all visible text
    const pageText = await page.evaluate(() => document.body.innerText);
    console.log('\n=== VISIBLE TEXT ===');
    console.log(pageText);
    
    // Get all button elements
    const buttons = await page.evaluate(() => {
      const btns = Array.from(document.querySelectorAll('button'));
      return btns.map(btn => ({
        text: btn.innerText.trim(),
        className: btn.className,
        disabled: btn.disabled,
        type: btn.type,
        visible: btn.offsetParent !== null
      }));
    });
    
    console.log('\n=== ALL BUTTONS ===');
    console.log(JSON.stringify(buttons, null, 2));
    
    // Check for specific text
    const hasAIFactory = await page.locator('text=AI Factory').count();
    const hasAddButton = await page.locator('button:has-text("Add AI Model")').count();
    const hasAddIcon = await page.locator('button:has-text("+ Add")').count();
    
    console.log('\n=== TEXT CHECKS ===');
    console.log(`"AI Factory" found: ${hasAIFactory} times`);
    console.log(`"Add AI Model" button found: ${hasAddButton} times`);
    console.log(`"+ Add" button found: ${hasAddIcon} times`);
    
    // Take screenshot
    await page.screenshot({ path: 'test-results/llm-config-buttons-debug.png', fullPage: true });
    
    // Get HTML of main content area
    const mainHTML = await page.evaluate(() => {
      const main = document.querySelector('main') || document.querySelector('[role="main"]') || document.querySelector('.container');
      return main ? main.innerHTML : document.body.innerHTML;
    });
    
    console.log('\n=== MAIN HTML (first 1000 chars) ===');
    console.log(mainHTML.substring(0, 1000));
  });
});
