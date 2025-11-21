import { test, expect } from '@playwright/test';

test.describe('Review Authentication Flow', () => {
  test('clicking Review card on dashboard navigates to Review app', async ({ page }) => {
    // GIVEN: User visits Dashboard (unauthenticated is fine - we just want to test navigation)
    await page.goto('/dashboard', { waitUntil: 'networkidle' });
    
    // THEN: Dashboard loads
    await expect(page.locator('text=Code Review')).toBeVisible({ timeout: 10000 });
    
    console.log('✅ Dashboard loaded');
    
    // WHEN: User clicks the Review card
    const reviewCard = page.locator('a[href="/review"]').first();
    await expect(reviewCard).toBeVisible();
    
    console.log('✅ Review card found, clicking...');
    
    // Click and wait for navigation
    const [response] = await Promise.all([
      page.waitForResponse(resp => resp.url().includes('/review'), { timeout: 10000 }),
      reviewCard.click()
    ]);
    
    console.log(`✅ Navigation response: ${response.status()} ${response.url()}`);
    
    // THEN: Should navigate away from dashboard
    // Either to Review app OR to login redirect
    await page.waitForURL(url => !url.toString().includes('/dashboard'), { timeout: 10000 });
    
    const finalUrl = page.url();
    console.log(`✅ Final URL: ${finalUrl}`);
    
    // Should be on either /review or /auth/github/login
    expect(finalUrl).toMatch(/\/(review|auth\/github\/login)/);
    
    console.log('✅ PASS: Navigation works - user left dashboard');
  });

  test('Review service returns redirect for unauthenticated requests', async ({ page }) => {
    // GIVEN: No authentication
    
    // WHEN: Directly accessing Review endpoint (follow redirects: false)
    const response = await page.request.get('/review', {
      maxRedirects: 0
    });
    
    // THEN: Response is a redirect (302), not unauthorized (401)
    expect(response.status()).toBe(302);
    expect(response.headers()['location']).toContain('/auth/github/login');
    
    console.log('✅ PASS: Review returns 302 redirect (not 401)');
    console.log(`   Location: ${response.headers()['location']}`);
  });

  test('Review service does NOT return 401 Unauthorized', async ({ page }) => {
    // GIVEN: No authentication
    
    // WHEN: Directly accessing Review endpoint
    const response = await page.request.get('/review', {
      maxRedirects: 0
    });
    
    // THEN: Response is NOT 401 (this was the bug we fixed)
    expect(response.status()).not.toBe(401);
    
    console.log('✅ PASS: Review does not return 401 (bug fixed!)');
    console.log(`   Actual status: ${response.status()}`);
  });
});
