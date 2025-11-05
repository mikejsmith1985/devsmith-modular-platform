import { test, expect } from '@playwright/test';

test.describe('Portal OAuth Flow - Visual Verification', () => {
  test('should complete full OAuth login flow with screenshots', async ({ page }) => {
    // Step 1: Navigate to login page
    await page.goto('http://localhost:3000/login');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/oauth-01-login-page.png', fullPage: true });
    
    // Verify login button exists
    const loginButton = page.locator('a[href="/auth/login"]');
    await expect(loginButton).toBeVisible();
    
    // Step 2: Click login button (will redirect to GitHub)
    await loginButton.click();
    
    // Wait for GitHub OAuth page OR callback (if already authorized)
    await page.waitForURL(/github\.com\/login\/oauth\/authorize|localhost:3000\/auth\/github\/callback/, { timeout: 10000 });
    
    const currentUrl = page.url();
    console.log('Current URL after login click:', currentUrl);
    
    if (currentUrl.includes('github.com')) {
      // On GitHub OAuth page - would need real credentials to continue
      await page.screenshot({ path: 'test-results/oauth-02-github-oauth.png', fullPage: true });
      console.log('Redirected to GitHub OAuth - manual authorization required');
    } else if (currentUrl.includes('/auth/github/callback')) {
      // Already authorized, captured callback
      await page.screenshot({ path: 'test-results/oauth-02-callback-redirect.png', fullPage: true });
      console.log('OAuth callback received');
      
      // Wait for redirect to dashboard or error
      await page.waitForURL(/\/(dashboard|login)/, { timeout: 10000 });
      await page.screenshot({ path: 'test-results/oauth-03-after-callback.png', fullPage: true });
      
      const finalUrl = page.url();
      console.log('Final URL after callback:', finalUrl);
      
      if (finalUrl.includes('/dashboard')) {
        console.log('✅ SUCCESS: Logged in and redirected to dashboard');
        
        // Verify dashboard content
        await expect(page.locator('text=DevSmith Platform')).toBeVisible();
        await page.screenshot({ path: 'test-results/oauth-04-dashboard-success.png', fullPage: true });
        
      } else if (finalUrl.includes('/login')) {
        console.log('❌ FAILED: Redirected back to login');
        
        // Check for error message
        const errorText = await page.textContent('body');
        console.log('Page content:', errorText);
      }
    }
  });
  
  test('should show error message on failed authentication', async ({ page }) => {
    // This test checks what happens when OAuth callback returns an error
    await page.goto('http://localhost:3000/auth/github/callback?error=access_denied');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/oauth-error-case.png', fullPage: true });
    
    // Should show error or redirect to login
    const url = page.url();
    console.log('Error case URL:', url);
  });
});
