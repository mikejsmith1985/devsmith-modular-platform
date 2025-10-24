import { test, expect } from '@playwright/test';

// E2E test: Portal login and dashboard user flow using test authentication endpoint
// This bypasses GitHub OAuth while still testing the complete user workflow

test('User can login and see dashboard with test auth', async ({ page }) => {
  // Use the test authentication endpoint to create a valid session
  const testUser = {
    username: 'test-user',
    email: 'test@example.com',
    avatar_url: 'https://avatars.githubusercontent.com/u/12345?v=4'
  };

  // Navigate to a dummy page first to establish domain context
  await page.goto('http://localhost:3000/');

  // Call the test login endpoint using page.evaluate to make the request from the page context
  // This ensures cookies are properly set in the browser
  const responseData = await page.evaluate(async (userData) => {
    const response = await fetch('http://localhost:3000/auth/test-login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
      credentials: 'include' // Important: include cookies in the request
    });
    return response.json();
  }, testUser);

  console.log('Test login response:', responseData);
  expect(responseData.success).toBe(true);

  // Navigate to dashboard - should be authenticated now with cookie set
  await page.goto('http://localhost:3000/dashboard');

  // Validate dashboard page renders correctly
  await expect(page.locator('.dashboard-container')).toBeVisible({ timeout: 10000 });
  await expect(page.locator('.avatar')).toBeVisible();

  // Check that username appears in the dashboard
  const dashboardContent = await page.textContent('body');
  expect(dashboardContent).toContain(testUser.username);
});

// Optional: Test with custom user data
test('Dashboard renders with custom user data', async ({ page }) => {
  const customUser = {
    username: 'custom-tester',
    email: 'custom@test.com',
    avatar_url: 'https://example.com/custom-avatar.png'
  };

  // Navigate to establish context
  await page.goto('http://localhost:3000/');

  // Authenticate with custom user using page context
  const responseData = await page.evaluate(async (userData) => {
    const response = await fetch('http://localhost:3000/auth/test-login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(userData),
      credentials: 'include'
    });
    return response.json();
  }, customUser);

  expect(responseData.success).toBe(true);

  // Navigate to dashboard with increased timeout
  await page.goto('http://localhost:3000/dashboard', { waitUntil: 'domcontentloaded', timeout: 10000 });

  // Verify custom user data is displayed
  await expect(page.locator('.dashboard-container')).toBeVisible({ timeout: 10000 });
  const dashboardContent = await page.textContent('body');
  expect(dashboardContent).toContain(customUser.username);
});
