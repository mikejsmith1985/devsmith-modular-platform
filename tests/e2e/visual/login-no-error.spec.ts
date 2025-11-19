import { test, expect } from '@playwright/test';
import percySnapshot from '@percy/playwright';

test.describe('Login Page - Visual Verification', () => {
  test('Login page loads without errors - VISUAL', async ({ page }) => {
    console.log('🧪 Testing login page loads cleanly...');
    
    // STEP 1: Navigate to login page
    await page.goto('/login');
    console.log('📍 Navigated to login page');
    
    // STEP 2: Wait for React to load
    await page.waitForSelector('.card', { timeout: 5000 });
    console.log('✅ Login card rendered');
    
    // STEP 3: Verify NO error messages visible
    const errorAlert = await page.locator('.alert-danger').count();
    expect(errorAlert).toBe(0);
    console.log('✅ No error alerts shown');
    
    // STEP 4: Verify essential elements exist
    await expect(page.locator('h2:has-text("DevSmith Platform")')).toBeVisible();
    console.log('✅ Title visible');
    
    await expect(page.locator('button:has-text("Login with GitHub")')).toBeVisible();
    console.log('✅ GitHub login button visible');
    
    await expect(page.locator('input[type="email"]')).toBeVisible();
    console.log('✅ Email input visible');
    
    await expect(page.locator('input[type="password"]')).toBeVisible();
    console.log('✅ Password input visible');
    
    // STEP 5: Check console for errors
    const logs: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      logs.push(text);
      if (msg.type() === 'error') {
        console.log(`❌ Console error: ${text}`);
      }
    });
    
    // Wait a moment for any async errors
    await page.waitForTimeout(1000);
    
    // Should NOT have "Failed to fetch user" error
    const hasAuthError = logs.some(log => log.includes('Failed to fetch user'));
    expect(hasAuthError).toBe(false);
    console.log('✅ No "Failed to fetch user" error in console');
    
    // STEP 6: PERCY SNAPSHOT - Visual proof
    await percySnapshot(page, 'Login Page - No Errors');
    console.log('📸 Percy snapshot captured');
    
    console.log('🎉 Login page verification complete - ALL CHECKS PASSED');
  });
  
  test('Login page shows proper error ONLY when login fails', async ({ page }) => {
    console.log('🧪 Testing login page shows errors appropriately...');
    
    await page.goto('/login');
    await page.waitForSelector('.card');
    
    // Initially NO errors
    let errorAlert = await page.locator('.alert-danger').count();
    expect(errorAlert).toBe(0);
    console.log('✅ No errors on initial load');
    
    // Try to submit without credentials
    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    // NOW should show error
    await page.waitForSelector('.alert-danger', { timeout: 3000 });
    errorAlert = await page.locator('.alert-danger').count();
    expect(errorAlert).toBe(1);
    console.log('✅ Error shown AFTER failed login attempt');
    
    await percySnapshot(page, 'Login Page - After Failed Login');
    console.log('📸 Percy snapshot captured');
  });
  
  test('OAuth flow initiates without showing auth errors', async ({ page }) => {
    console.log('🧪 Testing OAuth initiation...');
    
    await page.goto('/login');
    await page.waitForSelector('.card');
    
    // No errors before clicking GitHub button
    const errorBeforeClick = await page.locator('.alert-danger').count();
    expect(errorBeforeClick).toBe(0);
    console.log('✅ No errors before OAuth');
    
    // Click GitHub login button
    const githubButton = page.locator('button:has-text("Login with GitHub")');
    await githubButton.click();
    
    // Should redirect to GitHub
    await page.waitForURL(/github\.com\/login/, { timeout: 5000 });
    console.log('✅ Redirected to GitHub OAuth');
    
    // Verify PKCE parameters in URL
    const url = page.url();
    const decodedUrl = decodeURIComponent(url);
    expect(decodedUrl).toContain('code_challenge=');
    expect(decodedUrl).toContain('code_challenge_method=S256');
    console.log('✅ PKCE parameters present');
    
    await percySnapshot(page, 'OAuth Redirect - GitHub Login Page');
    console.log('📸 Percy snapshot captured');
  });
});
