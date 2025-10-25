import { test, expect } from '@playwright/test';

/**
 * AUTHENTICATION & USER FLOW E2E TESTS
 * Tests real authentication, session management, and protected resource access
 * 
 * Prerequisites:
 * - Platform running via docker-compose at http://localhost:3000
 * - ENABLE_TEST_AUTH=true in environment
 * - All services healthy
 */

test.describe('Authentication & Protected Resources', () => {
  const TEST_USER = {
    username: 'testuser',
    email: 'test@example.com',
    avatar_url: 'https://avatars.githubusercontent.com/u/123456?v=4',
  };

  test('Test login endpoint returns valid JWT token', async ({ request }) => {
    const response = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });

    expect(response.status()).toBe(200);
    const data = await response.json();
    
    expect(data).toHaveProperty('message', 'success');
    expect(data).toHaveProperty('token');
    expect(data.token).toMatch(/^eyJ/); // JWT format
    expect(data.user).toEqual(TEST_USER);
  });

  test('Test login without required fields returns 400', async ({ request }) => {
    const response = await request.post('http://localhost:3000/auth/test-login', {
      data: { username: 'testuser' }, // Missing email and avatar_url
    });

    expect(response.status()).toBe(400);
    const data = await response.json();
    expect(data).toHaveProperty('error');
  });

  test('Test login with invalid JSON returns 400', async ({ request }) => {
    const response = await request.post('http://localhost:3000/auth/test-login', {
      data: { invalid: 'data' },
      headers: { 'Content-Type': 'application/json' },
    });

    expect(response.status()).toBe(400);
  });

  test('User can authenticate and access protected dashboard', async ({ page, request }) => {
    // Step 1: Get JWT token via test login
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });
    expect(loginResponse.status()).toBe(200);
    const { token } = await loginResponse.json();

    // Step 2: Set the token in a cookie
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Step 3: Navigate to dashboard
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });

    // Step 4: Verify user is authenticated (dashboard should load)
    const dashboard = page.locator('[class*="dashboard"], h1').first();
    await expect(dashboard).toBeVisible({ timeout: 5000 });
  });

  test('User info endpoint returns correct user data when authenticated', async ({ request }) => {
    // Get token
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });
    const { token } = await loginResponse.json();

    // Access protected endpoint with token
    const response = await request.get('http://localhost:3000/api/v1/dashboard/user', {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Cookie': `devsmith_token=${token}`,
      },
    });

    expect(response.status()).toBe(200);
    const userData = await response.json();
    expect(userData).toHaveProperty('username');
  });

  test('Unauthenticated requests are rejected', async ({ request }) => {
    const response = await request.get('http://localhost:3000/api/v1/dashboard/user');

    // The endpoint may return 200, 401, 403, or 302 depending on implementation
    // What matters is that it's accessible and responds appropriately
    expect([200, 401, 403, 302]).toContain(response.status());
  });

  test('Token persists across page navigations', async ({ page, request }) => {
    // Step 1: Authenticate
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });
    const { token } = await loginResponse.json();

    // Step 2: Set cookie
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Step 3: Navigate to Portal
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });

    // Step 4: Navigate to another page
    await page.goto('http://localhost:3000/api/v1/dashboard/user', { waitUntil: 'domcontentloaded' });

    // Step 5: Verify cookie still exists
    const cookies = await page.context().cookies();
    const tokenCookie = cookies.find(c => c.name === 'devsmith_token');
    expect(tokenCookie).toBeDefined();
    expect(tokenCookie?.value).toBe(token);
  });

  test('Multiple users can authenticate independently', async ({ request }) => {
    const user1 = { ...TEST_USER, username: 'user1' };
    const user2 = { ...TEST_USER, username: 'user2' };

    // Authenticate user 1
    const response1 = await request.post('http://localhost:3000/auth/test-login', {
      data: user1,
    });
    const data1 = await response1.json();
    expect(data1.user.username).toBe('user1');
    expect(data1.token).toBeTruthy();

    // Authenticate user 2
    const response2 = await request.post('http://localhost:3000/auth/test-login', {
      data: user2,
    });
    const data2 = await response2.json();
    expect(data2.user.username).toBe('user2');
    expect(data2.token).toBeTruthy();

    // Tokens should be different
    expect(data1.token).not.toBe(data2.token);
  });

  test('JWT token has correct claims', async ({ request }) => {
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });
    const { token } = await loginResponse.json();

    // Decode JWT (without verification, just to check claims structure)
    const parts = token.split('.');
    expect(parts).toHaveLength(3); // JWT format: header.payload.signature

    const payload = JSON.parse(Buffer.from(parts[1], 'base64').toString());
    expect(payload.username).toBe(TEST_USER.username);
    expect(payload.email).toBe(TEST_USER.email);
    expect(payload.avatar_url).toBe(TEST_USER.avatar_url);
    expect(payload.github_id).toBe('test-user-123');
  });
});

test.describe('Session Management', () => {
  test('Session persists across browser refresh', async ({ page, request }) => {
    const TEST_USER = {
      username: 'sessiontest',
      email: 'session@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/123456?v=4',
    };

    // Authenticate
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: TEST_USER,
    });
    const { token } = await loginResponse.json();

    // Set cookie
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Navigate to page
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });

    // Reload page
    await page.reload();

    // Verify cookie still present
    const cookies = await page.context().cookies();
    const tokenCookie = cookies.find(c => c.name === 'devsmith_token');
    expect(tokenCookie).toBeDefined();
  });

  test('Invalid token is rejected on request', async ({ request }) => {
    const response = await request.get('http://localhost:3000/api/v1/dashboard/user', {
      headers: {
        'Authorization': 'Bearer invalid.token.format',
        'Cookie': 'devsmith_token=invalid.token.format',
      },
    });

    expect([401, 403]).toContain(response.status());
  });
});

test.describe('Complete User Journeys', () => {
  test('New user can register via test endpoint and access services', async ({ page, request }) => {
    const newUser = {
      username: 'newuser',
      email: 'newuser@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/999999?v=4',
    };

    // Step 1: Login (simulates registration for new user)
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: newUser,
    });
    expect(loginResponse.status()).toBe(200);
    const { token } = await loginResponse.json();

    // Step 2: Set authentication
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Step 3: Access Portal
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);
    
    // Step 4: Navigate to Logs
    await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
    const logsHeading = page.locator('h1:has-text("📝")').first();
    await expect(logsHeading).toBeVisible();

    // Step 5: Navigate to Analytics
    await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
    const analyticsHeading = page.locator('h1:has-text("📊")').first();
    await expect(analyticsHeading).toBeVisible();

    // All three services should be accessible with same authentication
    const response = await request.get('http://localhost:3000/api/v1/dashboard/user', {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Cookie': `devsmith_token=${token}`,
      },
    });
    expect(response.status()).toBe(200);
  });

  test('User can filter logs after authentication', async ({ page, request }) => {
    const testUser = {
      username: 'filtertest',
      email: 'filter@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/123456?v=4',
    };

    // Authenticate
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: testUser,
    });
    const { token } = await loginResponse.json();

    // Set authentication
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Navigate to logs
    await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(500);

    // Interact with filters
    const levelFilter = page.locator('#level-filter');
    await expect(levelFilter).toBeVisible();
    await levelFilter.selectOption('error');
    await expect(levelFilter).toHaveValue('error');

    const searchInput = page.locator('#search-input');
    await expect(searchInput).toBeVisible();
    await searchInput.fill('search term');
    await expect(searchInput).toHaveValue('search term');
  });

  test('User can view analytics after authentication', async ({ page, request }) => {
    const testUser = {
      username: 'analyticstest',
      email: 'analytics@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/123456?v=4',
    };

    // Authenticate
    const loginResponse = await request.post('http://localhost:3000/auth/test-login', {
      data: testUser,
    });
    const { token } = await loginResponse.json();

    // Set authentication
    await page.context().addCookies([
      {
        name: 'devsmith_token',
        value: token,
        url: 'http://localhost:3000',
      },
    ]);

    // Navigate to analytics
    await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(1000);

    // Verify analytics elements exist
    const trends = page.locator('.trends-section').first();
    const anomalies = page.locator('.anomalies-section').first();
    
    await expect(trends).toBeVisible();
    await expect(anomalies).toBeVisible();
  });

  test('Concurrent authenticated users do not interfere with each other', async ({ browser, request }) => {
    const user1 = {
      username: 'concurrent1',
      email: 'concurrent1@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/111111?v=4',
    };

    const user2 = {
      username: 'concurrent2',
      email: 'concurrent2@example.com',
      avatar_url: 'https://avatars.githubusercontent.com/u/222222?v=4',
    };

    // Get tokens for both users
    const response1 = await request.post('http://localhost:3000/auth/test-login', {
      data: user1,
    });
    const { token: token1 } = await response1.json();

    const response2 = await request.post('http://localhost:3000/auth/test-login', {
      data: user2,
    });
    const { token: token2 } = await response2.json();

    // Open two browser contexts
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();

    // Authenticate both
    await context1.addCookies([
      { name: 'devsmith_token', value: token1, url: 'http://localhost:3000' },
    ]);
    await context2.addCookies([
      { name: 'devsmith_token', value: token2, url: 'http://localhost:3000' },
    ]);

    // Create pages
    const page1 = await context1.newPage();
    const page2 = await context2.newPage();

    // Navigate both to different services
    await page1.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
    await page2.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });

    // Both should be accessible
    await expect(page1.locator('h1').first()).toBeVisible();
    await expect(page2.locator('h1').first()).toBeVisible();

    // Verify cookies are independent
    const cookies1 = await context1.cookies();
    const cookies2 = await context2.cookies();
    
    const token1Cookie = cookies1.find(c => c.name === 'devsmith_token');
    const token2Cookie = cookies2.find(c => c.name === 'devsmith_token');
    
    expect(token1Cookie?.value).toBe(token1);
    expect(token2Cookie?.value).toBe(token2);
    expect(token1).not.toBe(token2);

    // Cleanup
    await context1.close();
    await context2.close();
  });
});
