import { test, expect } from '@playwright/test';

test('debug: nav on authenticated dashboard', async ({ page }) => {
  // Auth
  await page.request.post('http://localhost:3000/auth/test-login', {
    data: {
      username: 'testuser',
      email: 'test@example.com',
      avatar_url: 'http://example.com/avatar.png'
    }
  });
  
  // Go to dashboard (authenticated route)
  await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
  
  // Check for nav
  const nav = page.locator('nav');
  console.log('Nav elements found:', await nav.count());
  
  // Check for DevSmith text
  const devsmithText = page.locator('text=DevSmith');
  console.log('DevSmith text found:', await devsmithText.count());
  
  // Check for dark mode button
  const darkButton = page.locator('button[type="button"]').filter({ has: page.locator('svg') });
  console.log('Dark mode buttons found:', await darkButton.count());
  
  // Check Alpine.js attributes
  const alpine = page.locator('[x-data*="dark"]');
  console.log('Alpine dark containers:', await alpine.count());
  
  // Get nav HTML
  if (await nav.count() > 0) {
    const navHTML = await nav.first().innerHTML();
    console.log('Nav HTML (first 300 chars):', navHTML.substring(0, 300));
  }
  
  expect(await nav.count()).toBeGreaterThan(0);
  expect(await alpine.count()).toBeGreaterThan(0);
});
