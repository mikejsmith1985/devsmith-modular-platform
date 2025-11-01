import { test, expect } from '@playwright/test';

test('debug: dark mode toggle works with vanilla JS', async ({ page }) => {
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
  
  // Check for dark mode button
  const darkButton = page.locator('#dark-mode-toggle');
  console.log('Dark button found:', await darkButton.count());
  await expect(darkButton).toBeVisible();
  
  // Check for icons
  const sunIcon = page.locator('#sun-icon');
  const moonIcon = page.locator('#moon-icon');
  console.log('Sun icon found:', await sunIcon.count());
  console.log('Moon icon found:', await moonIcon.count());
  
  // Get initial dark class
  const html = page.locator('html');
  const initialClass = await html.getAttribute('class');
  console.log('Initial HTML class:', initialClass);
  
  // Click toggle
  await darkButton.click();
  await page.waitForTimeout(300);
  
  // Check dark class changed
  const updatedClass = await html.getAttribute('class');
  console.log('Updated HTML class:', updatedClass);
  
  const wasDark = initialClass?.includes('dark');
  const isDark = updatedClass?.includes('dark');
  console.log('Was dark?', wasDark, 'Is dark?', isDark);
  
  expect(wasDark).not.toBe(isDark);
});
