const { test, expect } = require('@playwright/test');

test.describe('LLM Configuration UI Debug', () => {
  test('debug page load with console capture', async ({ page }) => {
    // Capture console messages
    const consoleMessages = [];
    page.on('console', msg => {
      consoleMessages.push(`${msg.type()}: ${msg.text()}`);
      console.log(`BROWSER ${msg.type()}: ${msg.text()}`);
    });

    // Capture page errors
    page.on('pageerror', error => {
      console.log(`PAGE ERROR: ${error.message}`);
      consoleMessages.push(`ERROR: ${error.message}`);
    });

    // Authenticate first
    console.log('Step 1: Authenticating...');
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
    console.log('Auth response:', authData);
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
        console.log('Cookie set successfully');
      }
    }

    // Navigate to page
    console.log('Step 2: Navigating to /llm-config...');
    const response = await page.goto('http://localhost:3000/llm-config', {
      waitUntil: 'networkidle',
      timeout: 30000
    });
    console.log('Page loaded, status:', response.status());

    // Wait for any of these selectors
    console.log('Step 3: Waiting for page content...');
    try {
      await Promise.race([
        page.waitForSelector('text=AI Factory', { timeout: 5000 }),
        page.waitForSelector('text=DevSmith Platform', { timeout: 5000 }),
        page.waitForSelector('.spinner-border', { timeout: 5000 }),
        page.waitForSelector('.alert-danger', { timeout: 5000 })
      ]);
      console.log('Some content appeared');
    } catch (e) {
      console.log('Timeout waiting for content');
    }

    // Take screenshot
    await page.screenshot({ path: 'test-results/llm-config-debug.png', fullPage: true });

    // Get page HTML
    const html = await page.content();
    console.log('Page HTML length:', html.length);
    console.log('Contains "AI Factory":', html.includes('AI Factory'));
    console.log('Contains "DevSmith":', html.includes('DevSmith'));

    // Get all visible text
    const bodyText = await page.locator('body').textContent();
    console.log('Visible text length:', bodyText.length);
    console.log('First 500 chars:', bodyText.substring(0, 500));

    // Check if JavaScript loaded
    const scripts = await page.locator('script[src]').count();
    console.log('Number of external scripts:', scripts);

    // Print all console messages
    console.log('\n=== All Console Messages ===');
    consoleMessages.forEach(msg => console.log(msg));
    console.log('=== End Console Messages ===\n');

    // Check for specific elements
    const hasAIFactory = await page.locator('text=AI Factory').count();
    const hasNavbar = await page.locator('nav.navbar').count();
    const hasCard = await page.locator('.card').count();

    console.log('Elements found:');
    console.log('- AI Factory text:', hasAIFactory);
    console.log('- Navbar:', hasNavbar);
    console.log('- Cards:', hasCard);

    // Fail if AI Factory not found
    expect(hasAIFactory).toBeGreaterThan(0);
  });
});
