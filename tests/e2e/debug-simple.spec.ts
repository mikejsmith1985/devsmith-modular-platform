import { test, expect } from '@playwright/test';

test('debug: basic load', async ({ page }) => {
  console.log('TEST START');
  
  console.log('Navigating...');
  const response = await page.goto('http://localhost:3000', { 
    waitUntil: 'domcontentloaded', 
    timeout: 15000 
  });
  console.log('Navigation complete, status:', response?.status());
  
  const title = await page.title();
  console.log('Page title:', title);
  
  expect(title.length).toBeGreaterThan(0);
  console.log('TEST PASS');
});
