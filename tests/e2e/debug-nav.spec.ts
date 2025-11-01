import { test, expect } from '@playwright/test';

test('debug: check for nav element', async ({ page }) => {
  console.log('TEST START - Nav Debug');
  
  // Auth first
  const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  console.log('Auth:', loginResponse.ok());
  
  // Navigate
  await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
  console.log('Navigated');
  
  // Get HTML
  const html = await page.content();
  console.log('HTML length:', html.length);
  console.log('Contains <nav>?', html.includes('<nav'));
  console.log('Contains navbar?', html.includes('navbar'));
  console.log('Contains DevSmith?', html.includes('DevSmith'));
  
  // Try to find nav
  const nav = page.locator('nav');
  console.log('Nav count:', await nav.count());
  
  // Check what we got
  const body = await page.locator('body').innerHTML();
  console.log('Body HTML (first 500 chars):', body.substring(0, 500));
});
