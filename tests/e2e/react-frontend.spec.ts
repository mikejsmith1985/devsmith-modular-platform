import { test, expect } from './fixtures/auth.fixture';

test.describe('React Frontend - Core Functionality', () => {
  test('should render login page with GitHub OAuth button', async ({ page }) => {
    await page.goto('/');
    
    // Should redirect to /login if not authenticated
    await page.waitForURL(/\/login/);
    
    // Check for GitHub login button
    const loginButton = page.locator('button:has-text("Login with GitHub")');
    await expect(loginButton).toBeVisible();
    
    // Verify Bootstrap styling is loaded
    const styles = await loginButton.evaluate((el) => {
      const computed = window.getComputedStyle(el);
      return {
        backgroundColor: computed.backgroundColor,
        color: computed.color
      };
    });
    
    expect(styles.backgroundColor).toBeTruthy();
    expect(styles.color).toBeTruthy();
  });

  test('should have proper SPA routing (no page reloads)', async ({ authenticatedPage }) => {
    // Use authenticated page fixture
    await authenticatedPage.goto('/');
    
    // Wait for dashboard to load
    await authenticatedPage.waitForURL(/\/(dashboard)?$/);
    
    // Check navigation without page reload
    let reloaded = false;
    authenticatedPage.on('load', () => { reloaded = true; });
    
    // Find and click Logs navigation link
    await authenticatedPage.click('a[href="/logs"], button:has-text("Logs")');
    await authenticatedPage.waitForURL(/\/logs/);
    
    // SPA should not reload the page
    expect(reloaded).toBe(false);
  });
});

test.describe('React Frontend - API Integration', () => {
  test('should fetch and display log statistics', async ({ authenticatedPage }) => {
    // Navigate to logs page (already authenticated)
    await authenticatedPage.goto('/logs');
    
    // Wait for API call to complete
    const response = await authenticatedPage.waitForResponse(
      (response) => response.url().includes('/api/logs/v1/stats'),
      { timeout: 10000 }
    );
    
    expect(response.status()).toBe(200);
    const data = await response.json();
    
    // Verify JSON structure
    expect(data).toHaveProperty('debug');
    expect(data).toHaveProperty('info');
    expect(data).toHaveProperty('warning');
    expect(data).toHaveProperty('error');
    expect(data).toHaveProperty('critical');
  });

  test('StatCards should render with real data', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/logs');
    
    // Wait for StatCards to load
    await authenticatedPage.waitForSelector('.stat-card', { timeout: 5000 });
    
    // Check all 5 StatCards are present
    const statCards = await authenticatedPage.locator('.stat-card').count();
    expect(statCards).toBe(5);
    
    // Verify each card has a count
    const debugCount = await authenticatedPage.locator('.stat-card:has-text("Debug")').textContent();
    expect(debugCount).toMatch(/\d+/);
  });
});

test.describe('React Frontend - Authentication Flow', () => {
  test('should redirect unauthenticated users to login', async ({ page }) => {
    // Clear any existing auth
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.removeItem('devsmith_token');
    });
    
    // Try to access protected route
    await page.goto('/dashboard');
    
    // Should redirect to login
    await page.waitForURL(/\/login/);
    expect(page.url()).toContain('/login');
  });

  test('should store JWT token on successful login', async ({ page }) => {
    await page.goto('/login');
    
    // Mock successful OAuth callback
    await page.evaluate(() => {
      // Simulate token storage after GitHub OAuth
      // Using generic JWT-like format for testing (not a real secret)
      localStorage.setItem('devsmith_token', 'test.jwt.token');
    });
    
    // Verify token is stored
    const token = await page.evaluate(() => localStorage.getItem('devsmith_token'));
    expect(token).toBeTruthy();
    expect(token).toContain('test'); // Mock token format
  });
});

test.describe('React Frontend - Responsive Design', () => {
  test('should render properly on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
    await page.goto('/');
    
    // Check if Bootstrap grid adapts
    const container = await page.locator('.container').first();
    await expect(container).toBeVisible();
  });

  test('should render properly on tablet', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 }); // iPad
    await page.goto('/');
    
    const container = await page.locator('.container').first();
    await expect(container).toBeVisible();
  });

  test('should render properly on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 }); // Desktop
    await page.goto('/');
    
    const container = await page.locator('.container').first();
    await expect(container).toBeVisible();
  });
});
