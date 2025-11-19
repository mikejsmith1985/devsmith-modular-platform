import { test, expect } from '@playwright/test';

test.describe('Review Navigation Flow - With Auth', () => {
  test('complete flow: login → dashboard → click Review card → navigate to Review', async ({ page, context }) => {
    // Enable verbose logging
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    page.on('pageerror', err => console.log('PAGE ERROR:', err));
    page.on('request', req => console.log('REQUEST:', req.method(), req.url()));
    page.on('response', res => console.log('RESPONSE:', res.status(), res.url()));
    
    // Step 1: Navigate to root (should show login)
    console.log('\n=== Step 1: Navigate to root ===');
    await page.goto('/', { waitUntil: 'networkidle' });
    console.log('Current URL:', page.url());
    
    // Step 2: Check if already logged in (by checking for dashboard redirect)
    const currentUrl = page.url();
    const parsedUrl = new URL(currentUrl);
    if (parsedUrl.pathname.includes('/dashboard')) {
      console.log('\n=== Already logged in, on dashboard ===');
    } else if (parsedUrl.pathname.includes('/login') || parsedUrl.pathname === '/') {
      console.log('\n=== Not logged in, need to authenticate ===');
      // For now, we'll set a mock JWT cookie to simulate being logged in
      // In real scenario, you'd go through OAuth flow
      
      const mockJWT = 'test-mock-jwt-token-for-e2e-testing';
      await context.addCookies([{
        name: 'devsmith_token',
        value: mockJWT,
        domain: parsedUrl.hostname,
        path: '/',
        httpOnly: false,
        secure: false,
        sameSite: 'Lax'
      }]);
      
      // Navigate to dashboard
      console.log('\n=== Step 3: Navigate to dashboard with auth cookie ===');
      await page.goto('/dashboard', { waitUntil: 'networkidle' });
      console.log('Dashboard URL:', page.url());
    }
    
    // Step 4: Verify we're on dashboard
    console.log('\n=== Step 4: Verify dashboard loaded ===');
    await page.waitForTimeout(1000); // Give page time to render
    
    const bodyText = await page.textContent('body');
    console.log('Page contains "Code Review"?', bodyText?.includes('Code Review'));
    console.log('Page contains "Welcome"?', bodyText?.includes('Welcome'));
    
    // Try to find the Review card
    const reviewCard = page.locator('a[href="/review"]').first();
    const cardExists = await reviewCard.count();
    console.log('Review card found?', cardExists > 0);
    
    if (cardExists === 0) {
      console.log('ERROR: Review card not found!');
      console.log('Full page HTML (first 500 chars):', (await page.content()).substring(0, 500));
      throw new Error('Review card not found on dashboard');
    }
    
    // Step 5: Click the Review card
    console.log('\n=== Step 5: Click Review card ===');
    await reviewCard.scrollIntoViewIfNeeded();
    
    // Get card details before clicking
    const href = await reviewCard.getAttribute('href');
    console.log('Card href:', href);
    const isVisible = await reviewCard.isVisible();
    console.log('Card visible?', isVisible);
    
    // Click and monitor what happens
    console.log('Clicking card...');
    await reviewCard.click();
    
    // Wait a bit to see what happens
    await page.waitForTimeout(2000);
    
    const newUrl = page.url();
    console.log('URL after click:', newUrl);
    
    // Check if we navigated away from dashboard
    if (newUrl.includes('/dashboard')) {
      console.log('❌ FAIL: Still on dashboard after click');
      console.log('Checking for navigation errors...');
      
      // Check browser console for errors
      const logs = await page.evaluate(() => {
        return {
          // @ts-ignore
          errors: window.__errors || [],
          location: window.location.href
        };
      });
      console.log('Browser state:', logs);
      
      throw new Error('Navigation failed - still on dashboard');
    } else {
      console.log('✅ SUCCESS: Navigated away from dashboard');
      console.log('New URL:', newUrl);
      
      // Verify we're on Review or login
      expect(newUrl).toMatch(/\/(review|auth\/github\/login)/);
    }
  });
});
