import { test, expect } from '@playwright/test';

test('debug: check cookie persistence', async ({ page }) => {
  console.log('TEST START - Cookie Debug');
  
  // Auth
  const loginResponse = await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  console.log('Auth status:', loginResponse.status());
  
  // Check for set-cookie header
  const headers = loginResponse.headers();
  console.log('Set-Cookie header:', headers['set-cookie']);
  
  // Check cookies
  const cookies1 = await page.context().cookies();
  console.log('Cookies after login POST:', cookies1.length);
  cookies1.forEach(c => console.log(`  ${c.name}=${c.value.substring(0,20)}...`));
  
  // Navigate
  await page.goto('http://localhost:3000', { waitUntil: 'domcontentloaded' });
  
  // Check cookies after navigation
  const cookies2 = await page.context().cookies();
  console.log('Cookies after navigate:', cookies2.length);
  cookies2.forEach(c => console.log(`  ${c.name}=${c.value.substring(0,20)}...`));
  
  // Get response headers
  const req = await page.evaluate(() => {
    return { cookies: document.cookie };
  });
  console.log('Document.cookie:', req.cookies);
});
