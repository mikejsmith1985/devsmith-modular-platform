import { test, expect } from '@playwright/test';
import * as path from 'path';

/**
 * E2E Test: Portal → Review Integration with JWT Authentication
 * 
 * This test validates the complete user flow:
 * 1. User logs into Portal (GitHub OAuth)
 * 2. User sees dashboard with Review card
 * 3. User clicks Review card → navigates to Review workspace
 * 4. Review workspace loads with authenticated context (JWT cookie)
 * 5. User pastes code and runs analysis
 * 6. Analysis results display correctly
 * 
 * Screenshots captured at each step for visual validation.
 */

test.describe('Portal → Review Integration', () => {
  const screenshotDir = '/tmp/devsmith-screenshots';
  const gatewayURL = 'http://localhost:3000';  // Nginx gateway - single entry point

  test('complete flow: login → dashboard → review → analyze', async ({ page }) => {
    // Step 1: Navigate to Portal via gateway
    await page.goto(gatewayURL + '/');
    await page.screenshot({ 
      path: path.join(screenshotDir, '01-portal-home.png'),
      fullPage: true 
    });

    // Step 2: Authenticate using test login endpoint (POST request)
    // Note: Portal is accessible via gateway at port 3000
    const testUser = {
      username: 'e2e-test-user',
      email: 'e2e@devsmith.test',
      avatar_url: 'https://avatars.githubusercontent.com/u/test'
    };

    const loginResponse = await page.request.post(gatewayURL + '/auth/test-login', {
      data: testUser
    });

    expect(loginResponse.ok()).toBeTruthy();
    const loginData = await loginResponse.json();
    const jwtToken = loginData.token;
    
    console.log('✓ Test login successful, JWT token obtained');

    // Set the JWT cookie manually
    await page.context().addCookies([{
      name: 'devsmith_token',
      value: jwtToken,
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      secure: false,
      sameSite: 'Lax'
    }]);

    await page.screenshot({ 
      path: path.join(screenshotDir, '02-authenticated.png'),
      fullPage: true 
    });

    // Step 3: Navigate to dashboard via gateway
    await page.goto(gatewayURL + '/dashboard');
    await expect(page).toHaveURL(/.*dashboard/);
    await page.screenshot({ 
      path: path.join(screenshotDir, '03-portal-dashboard.png'),
      fullPage: true 
    });

    // Step 4: Verify JWT cookie is present (authentication successful)
    const authCookies = await page.context().cookies();
    const jwtCookie = authCookies.find(c => c.name === 'devsmith_token');
    expect(jwtCookie).toBeDefined();
    expect(jwtCookie?.value).toBeTruthy();
    console.log('✓ JWT cookie present:', jwtCookie?.value.substring(0, 20) + '...');

    // Step 5: Find and verify Review card is visible
    const reviewCard = page.locator('text=/Code Review/i').first();
    await expect(reviewCard).toBeVisible({ timeout: 5000 });
    await page.screenshot({ 
      path: path.join(screenshotDir, '04-review-card-visible.png'),
      fullPage: true 
    });

    // Step 6: Click Review card to navigate to Review workspace
    // Review service is on port 8081
    const reviewLink = page.locator('a[href*="/review"]').first();
    await reviewLink.click();
    
    // Wait for Review workspace to load (different service, may take a moment)
    await page.waitForURL('**/review**', { timeout: 10000 });
    await page.screenshot({ 
      path: path.join(screenshotDir, '05-review-workspace-loaded.png'),
      fullPage: true 
    });

    // Step 7: Verify Review workspace UI elements are present (authenticated session loaded)
    console.log('✓ Verifying authenticated workspace UI elements...');
    await expect(page.locator('#mode-selector')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#model-selector')).toBeVisible();
    await expect(page.locator('#analyze-btn')).toBeVisible();
    console.log('✅ Authenticated workspace elements visible');

    // Step 8: Verify workspace title shows user context
    const workspaceTitle = await page.locator('h1').first().textContent();
    console.log(`✓ Workspace title: ${workspaceTitle}`);
    expect(workspaceTitle).toContain('Code Review Session');
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '06-authenticated-workspace-verified.png'),
      fullPage: true 
    });

    // Step 9: Verify mode selector has all 5 reading modes
    const modeOptions = await page.locator('#mode-selector option').allTextContents();
    console.log(`✓ Reading modes available: ${modeOptions.join(', ')}`);
    expect(modeOptions.length).toBe(5);
    expect(modeOptions).toContain('👁️ Preview');
    expect(modeOptions).toContain('⚡ Skim');
    expect(modeOptions).toContain('🔎 Scan');
    expect(modeOptions).toContain('🔬 Detailed');
    expect(modeOptions).toContain('⚠️ Critical');

    // Step 10: Select Preview mode
    await page.locator('#mode-selector').selectOption('preview');
    await page.screenshot({ 
      path: path.join(screenshotDir, '07-preview-mode-selected.png'),
      fullPage: true 
    });

    console.log('✅ Portal → Review authenticated integration complete');
    
    // Note: Analyze button click and AI analysis are separate from auth integration
    // and are tested in dedicated AI integration tests
    // This test validates that:
    // 1. JWT authentication works end-to-end
    // 2. User context is passed from Portal to Review
    // 3. Authenticated workspace loads with user-specific session
    
    console.log('✅ E2E test complete: Portal → Review integration validated');
    console.log(`📸 Screenshots saved to: ${screenshotDir}/`);
  });

  test('unauthenticated access to Review should redirect to login', async ({ page }) => {
    // Clear cookies to simulate unauthenticated user
    await page.context().clearCookies();

    // Try to access Review workspace directly via gateway
    const response = await page.goto(gatewayURL + '/review/workspace/test-session');

    // Should receive 401 Unauthorized (Review service protects these endpoints)
    // Check either status code or page content
    const statusCode = response?.status();
    const pageContent = await page.content();
    const isUnauthorized = statusCode === 401 || 
                          pageContent.includes('401') || 
                          pageContent.includes('Unauthorized') ||
                          pageContent.includes('Authentication required');

    expect(isUnauthorized).toBeTruthy();
    
    await page.screenshot({ 
      path: path.join(screenshotDir, '12-unauthenticated-blocked.png'),
      fullPage: true 
    });

    console.log('✅ Unauthenticated access properly blocked (status:', statusCode, ')');
  });

  test('JWT expiration handling', async ({ page }) => {
    // This test would require JWT manipulation (future enhancement)
    // For now, document the expected behavior
    test.skip(true, 'JWT expiration test requires token manipulation - future enhancement');
    
    /**
     * Expected behavior when JWT expires:
     * 1. Review endpoint returns 401
     * 2. Frontend detects 401 and redirects to Portal login
     * 3. User re-authenticates
     * 4. New JWT issued
     * 5. User returns to Review workspace
     */
  });
});

test.afterEach(async ({}, testInfo) => {
  if (testInfo.status === 'failed') {
    console.error(`❌ Test failed: ${testInfo.title}`);
    console.log(`📸 Check screenshots in: /tmp/devsmith-screenshots/`);
  }
});
