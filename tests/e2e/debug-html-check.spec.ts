import { test } from '@playwright/test';

test('debug: dump full HTML', async ({ page }) => {
  // Auth
  await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  
  // Go to dashboard
  await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
  
  // Get full HTML
  const html = await page.content();
  
  // Find the nav section
  const navStart = html.indexOf('<nav');
  const navEnd = html.indexOf('</nav>', navStart) + 6;
  
  if (navStart !== -1) {
    const navHTML = html.substring(navStart, navEnd);
    console.log('=== NAV HTML ===');
    console.log(navHTML.substring(0, 1500));
    console.log('...\n');
    console.log('Contains x-data?', navHTML.includes('x-data'));
    console.log('Contains x-init?', navHTML.includes('x-init'));
    console.log('Contains @click?', navHTML.includes('@click'));
  }
});
