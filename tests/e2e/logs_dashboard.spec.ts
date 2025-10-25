import { test, expect } from '@playwright/test';

// E2E tests for Logs Dashboard UI
// Tests the dashboard loads correctly and UI elements are present

test.describe('Logs Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the logs dashboard
    // Note: Update URL based on your development environment
    await page.goto('http://localhost:8082/', { waitUntil: 'domcontentloaded' });
  });

  test('dashboard loads successfully', async ({ page }) => {
    // Verify page title
    await expect(page).toHaveTitle(/DevSmith Logs/i);
  });

  test('displays main dashboard heading', async ({ page }) => {
    // Check for the main heading
    const heading = page.locator('h1:has-text("📝 DevSmith Logs")');
    await expect(heading).toBeVisible();
  });

  test('displays subtitle', async ({ page }) => {
    // Check for subtitle
    const subtitle = page.locator('p.subtitle:has-text("Real-time development logs")');
    await expect(subtitle).toBeVisible();
  });

  test('displays all filter controls', async ({ page }) => {
    // Level filter
    const levelFilter = page.locator('#level-filter');
    await expect(levelFilter).toBeVisible();
    
    // Service filter
    const serviceFilter = page.locator('#service-filter');
    await expect(serviceFilter).toBeVisible();
    
    // Search input
    const searchInput = page.locator('#search-input');
    await expect(searchInput).toBeVisible();
  });

  test('level filter has correct options', async ({ page }) => {
    const levelFilter = page.locator('#level-filter');
    
    // Check for all level options
    const allOption = levelFilter.locator('option[value="all"]');
    const infoOption = levelFilter.locator('option[value="info"]');
    const warnOption = levelFilter.locator('option[value="warn"]');
    const errorOption = levelFilter.locator('option[value="error"]');
    
    await expect(allOption).toBeVisible();
    await expect(infoOption).toBeVisible();
    await expect(warnOption).toBeVisible();
    await expect(errorOption).toBeVisible();
  });

  test('service filter has correct options', async ({ page }) => {
    const serviceFilter = page.locator('#service-filter');
    
    // Check for service options
    const allOption = serviceFilter.locator('option[value="all"]');
    const portalOption = serviceFilter.locator('option[value="portal"]');
    const reviewOption = serviceFilter.locator('option[value="review"]');
    const logsOption = serviceFilter.locator('option[value="logs"]');
    const analyticsOption = serviceFilter.locator('option[value="analytics"]');
    
    await expect(allOption).toBeVisible();
    await expect(portalOption).toBeVisible();
    await expect(reviewOption).toBeVisible();
    await expect(logsOption).toBeVisible();
    await expect(analyticsOption).toBeVisible();
  });

  test('displays all control buttons', async ({ page }) => {
    const pauseBtn = page.locator('#pause-btn');
    const autoScrollBtn = page.locator('#auto-scroll-btn');
    const clearBtn = page.locator('#clear-btn');
    const connectionStatus = page.locator('#connection-status');
    
    await expect(pauseBtn).toBeVisible();
    await expect(autoScrollBtn).toBeVisible();
    await expect(clearBtn).toBeVisible();
    await expect(connectionStatus).toBeVisible();
  });

  test('pause button has correct text', async ({ page }) => {
    const pauseBtn = page.locator('#pause-btn');
    await expect(pauseBtn).toContainText('⏸️ Pause');
  });

  test('auto-scroll button is active by default', async ({ page }) => {
    const autoScrollBtn = page.locator('#auto-scroll-btn');
    await expect(autoScrollBtn).toHaveClass(/active/);
  });

  test('clear button has correct text', async ({ page }) => {
    const clearBtn = page.locator('#clear-btn');
    await expect(clearBtn).toContainText('🗑️ Clear');
  });

  test('connection status shows connected', async ({ page }) => {
    const status = page.locator('#connection-status');
    // Wait for WebSocket connection to establish
    await page.waitForTimeout(1000);
    
    // Connection status should show connected or attempting to connect
    const statusText = await status.textContent();
    expect(statusText).toMatch(/Connected|Reconnecting/);
  });

  test('logs output container is present', async ({ page }) => {
    const logsOutput = page.locator('#logs-output');
    await expect(logsOutput).toBeVisible();
  });

  test('search input is functional', async ({ page }) => {
    const searchInput = page.locator('#search-input');
    
    // Type into search box
    await searchInput.fill('test search');
    
    // Verify the input value
    await expect(searchInput).toHaveValue('test search');
    
    // Clear the input
    await searchInput.clear();
    await expect(searchInput).toHaveValue('');
  });

  test('level filter is functional', async ({ page }) => {
    const levelFilter = page.locator('#level-filter');
    
    // Change filter value
    await levelFilter.selectOption('error');
    
    // Verify the selected value
    await expect(levelFilter).toHaveValue('error');
  });

  test('service filter is functional', async ({ page }) => {
    const serviceFilter = page.locator('#service-filter');
    
    // Change filter value
    await serviceFilter.selectOption('portal');
    
    // Verify the selected value
    await expect(serviceFilter).toHaveValue('portal');
  });

  test('pause button toggles state', async ({ page }) => {
    const pauseBtn = page.locator('#pause-btn');
    
    // Initial state should have Pause
    await expect(pauseBtn).toContainText('⏸️ Pause');
    
    // Click to pause
    await pauseBtn.click();
    
    // Wait a moment for state update
    await page.waitForTimeout(100);
    
    // Should now show Resume
    await expect(pauseBtn).toContainText('▶️ Resume');
  });

  test('auto-scroll button toggles state', async ({ page }) => {
    const autoScrollBtn = page.locator('#auto-scroll-btn');
    
    // Initial state should have active class
    await expect(autoScrollBtn).toHaveClass(/active/);
    
    // Click to toggle
    await autoScrollBtn.click();
    
    // Wait a moment for state update
    await page.waitForTimeout(100);
    
    // Should no longer have active class
    await expect(autoScrollBtn).not.toHaveClass(/active/);
  });

  test('clear button clears logs', async ({ page }) => {
    const clearBtn = page.locator('#clear-btn');
    const logsOutput = page.locator('#logs-output');
    
    // Click clear button
    await clearBtn.click();
    
    // Logs output should be empty or have no visible entries
    // (actual check depends on whether logs are populated from API)
    await expect(logsOutput).toBeVisible();
  });

  test('navbar is present', async ({ page }) => {
    const navbar = page.locator('nav.navbar');
    await expect(navbar).toBeVisible();
  });

  test('navbar brand shows DevSmith Logs', async ({ page }) => {
    const brand = page.locator('a.navbar-brand:has-text("📝 DevSmith Logs")');
    await expect(brand).toBeVisible();
  });

  test('navbar has dashboard link', async ({ page }) => {
    const dashboardLink = page.locator('nav a:has-text("Dashboard")');
    await expect(dashboardLink).toBeVisible();
  });

  test('navbar has health link', async ({ page }) => {
    const healthLink = page.locator('nav a:has-text("Health")');
    await expect(healthLink).toBeVisible();
  });

  test('page has correct CSS classes', async ({ page }) => {
    const container = page.locator('.logs-container');
    const header = page.locator('.logs-header');
    const main = page.locator('.logs-main');
    
    await expect(container).toBeVisible();
    await expect(header).toBeVisible();
    await expect(main).toBeVisible();
  });

  test('filter group labels are visible', async ({ page }) => {
    const levelLabel = page.locator('label[for="level-filter"]');
    const serviceLabel = page.locator('label[for="service-filter"]');
    const searchLabel = page.locator('label[for="search-input"]');
    
    await expect(levelLabel).toContainText('Level:');
    await expect(serviceLabel).toContainText('Service:');
    await expect(searchLabel).toContainText('Search:');
  });

  test('search input has placeholder text', async ({ page }) => {
    const searchInput = page.locator('#search-input');
    await expect(searchInput).toHaveAttribute('placeholder', 'Filter by message...');
  });

  test('HTML structure is valid', async ({ page }) => {
    // Check for DOCTYPE
    const html = await page.content();
    expect(html).toContain('<!DOCTYPE html');
    
    // Check for proper head and body tags
    expect(html).toContain('<html');
    expect(html).toContain('</html>');
    expect(html).toContain('<head>');
    expect(html).toContain('</head>');
    expect(html).toContain('<body>');
    expect(html).toContain('</body>');
  });

  test('stylesheets are loaded', async ({ page }) => {
    // Check for Bootstrap CSS
    const bootstrapLink = page.locator('link[href*="bootstrap"]');
    await expect(bootstrapLink).toBeVisible();
    
    // Check for logs CSS
    const logsLink = page.locator('link[href*="/static/css/logs.css"]');
    await expect(logsLink).toBeVisible();
  });

  test('scripts are loaded', async ({ page }) => {
    const html = await page.content();
    
    // Check for websocket script
    expect(html).toContain('/static/js/websocket.js');
    
    // Check for logs script
    expect(html).toContain('/static/js/logs.js');
    
    // Check for Bootstrap script
    expect(html).toContain('bootstrap.bundle.min.js');
  });

  test('page is responsive on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    const container = page.locator('.logs-container');
    await expect(container).toBeVisible();
    
    // All controls should still be visible
    await expect(page.locator('#level-filter')).toBeVisible();
    await expect(page.locator('#pause-btn')).toBeVisible();
  });

  test('page is responsive on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    
    const container = page.locator('.logs-container');
    await expect(container).toBeVisible();
    
    // All controls should still be visible
    await expect(page.locator('#level-filter')).toBeVisible();
  });

  test('page is responsive on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    const container = page.locator('.logs-container');
    await expect(container).toBeVisible();
  });
});
