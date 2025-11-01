import { test, expect } from '@playwright/test';

test('debug: try accessing dashboard directly', async ({ page }) => {
  console.log('TEST START - Dashboard Debug');
  
  // Auth
  const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  console.log('Auth status:', loginResponse.status());
  
  // Try dashboard directly (through nginx)
  console.log('Accessing /dashboard...');
  const dashResp = await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
  console.log('Dashboard status:', dashResp?.status());
  
  const html = await page.content();
  console.log('HTML length:', html.length);
  console.log('Has DevSmith:', html.includes('DevSmith'));
  console.log('HTML first 300 chars:', html.substring(0, 300));
});
