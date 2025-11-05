import { test as base, Page } from '@playwright/test';

/**
 * Test user data structure
 */
export interface TestUser {
  username: string;
  email: string;
  avatar_url: string;
  github_id: string;
}

/**
 * Auth fixtures for E2E tests
 * Provides authenticated page contexts for testing authenticated flows
 */
type AuthFixtures = {
  /**
   * Test user data (consistent across all tests)
   */
  testUser: TestUser;

  /**
   * Pre-authenticated page with valid session
   * Uses /auth/test-login endpoint to create session
   * Automatically sets JWT cookie
   */
  authenticatedPage: Page;
};

/**
 * Default test user
 * Consistent user data for all auth tests
 */
const DEFAULT_TEST_USER: TestUser = {
  username: 'testuser',
  email: 'test@example.com',
  avatar_url: 'https://example.com/avatar.png',
  github_id: '123456'
};

/**
 * Extended Playwright test with auth fixtures
 * 
 * Usage:
 * ```typescript
 * import { test, expect } from './fixtures/auth.fixture';
 * 
 * test('authenticated flow', async ({ authenticatedPage, testUser }) => {
 *   await authenticatedPage.goto('/dashboard');
 *   // User is already authenticated
 * });
 * ```
 */
export const test = base.extend<AuthFixtures>({
  /**
   * Provide consistent test user data
   */
  testUser: async ({}, use) => {
    await use(DEFAULT_TEST_USER);
  },

  /**
   * Create authenticated page context
   * 
   * Process:
   * 1. Call /auth/test-login to create Redis session
   * 2. Extract JWT token from Set-Cookie header
   * 3. Set cookie on page context
   * 4. Provide authenticated page to test
   * 
   * Requirements:
   * - ENABLE_TEST_AUTH=true in docker-compose.yml
   * - Portal service running with test auth endpoint
   */
  authenticatedPage: async ({ page, testUser }, use) => {
    // Call test auth endpoint to create session
    const response = await page.request.post('http://localhost:3000/auth/test-login', {
      data: testUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    // Verify session created successfully
    if (response.status() !== 200) {
      throw new Error(`Failed to create test session: ${response.status()} ${await response.text()}`);
    }

    const data = await response.json();
    
    // Verify we got a token
    if (!data.token || !data.token.Claims || !data.token.Claims.session_id) {
      throw new Error('Test auth endpoint did not return valid token with session_id');
    }

    // Extract JWT from Set-Cookie header
    const setCookieHeader = response.headers()['set-cookie'];
    if (!setCookieHeader) {
      throw new Error('No Set-Cookie header in test auth response');
    }

    // Parse cookie value
    const cookieMatch = setCookieHeader.match(/devsmith_token=([^;]+)/);
    if (!cookieMatch) {
      throw new Error('Could not extract devsmith_token from Set-Cookie header');
    }

    const tokenValue = cookieMatch[1];

    // Set cookie on page context for all subsequent requests
    await page.context().addCookies([{
      name: 'devsmith_token',
      value: tokenValue,
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      sameSite: 'Strict',
      expires: Math.floor(Date.now() / 1000) + (7 * 24 * 60 * 60) // 7 days
    }]);

    // Provide authenticated page to test
    await use(page);

    // Cleanup: Logout after test completes
    try {
      await page.goto('http://localhost:3000/auth/logout');
    } catch (error) {
      // Ignore logout errors (session might already be expired)
      console.log('Logout cleanup error (ignored):', error);
    }
  }
});

/**
 * Export expect from Playwright for convenience
 */
export { expect } from '@playwright/test';
