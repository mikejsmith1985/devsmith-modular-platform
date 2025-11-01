import { test, expect } from '@playwright/test';

test('debug: test auth endpoint', async ({ page }) => {
  console.log('TEST START - Auth Debug');
  
  console.log('POST to test-login...');
  const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  console.log('Login response status:', loginResponse.status());
  console.log('Login OK?', loginResponse.ok());
  
  const json = await loginResponse.json();
  console.log('Login response:', json.message);
  
  console.log('Navigating to localhost:3000...');
  const pageResp = await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
  console.log('Page response:', pageResp?.status());
  
  const title = await page.title();
  console.log('Page title:', title);
  
  console.log('TEST PASS');
});
