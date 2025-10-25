import { test, expect } from '@playwright/test';

/**
 * FULL USER FLOW E2E TESTS
 * Tests the complete user journey through the DevSmith platform
 * 
 * Prerequisites:
 * - Platform running via docker-compose at http://localhost:3000 (nginx reverse proxy)
 * - Services: Portal (8080), Review (8081), Logs (8082), Analytics (8083)
 */

test.describe('DevSmith Platform - Complete User Flow', () => {
    
     // Test Portal landing and health
     test('Portal service is accessible via nginx proxy', async ({ page }) => {
       await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       
       // Should see Portal login page with login button
       const loginButton = page.locator('a.login-button, [href*="auth"]').first();
       await expect(loginButton).toBeVisible();
     });

     test('Portal health check endpoint responds', async ({ page }) => {
       const response = await page.goto('http://localhost:3000/health', { waitUntil: 'domcontentloaded' });
       expect(response?.status()).toBe(200);
     });

     // Test Logs service access
     test('Logs service is accessible via nginx proxy', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Check for logs dashboard elements
       const heading = page.locator('h1:has-text("📝 DevSmith Logs")');
       await expect(heading).toBeVisible({ timeout: 5000 });
     });

     test('Logs dashboard loads with all UI elements', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Wait for dashboard to load
       await page.waitForTimeout(1000);
       
       // Check filter controls exist
       const levelFilter = page.locator('#level-filter');
       const serviceFilter = page.locator('#service-filter');
       const searchInput = page.locator('#search-input');
       
       await expect(levelFilter).toBeVisible();
       await expect(serviceFilter).toBeVisible();
       await expect(searchInput).toBeVisible();
       
       // Check control buttons exist
       const pauseBtn = page.locator('#pause-btn');
       const autoScrollBtn = page.locator('#auto-scroll-btn');
       const clearBtn = page.locator('#clear-btn');
       
       await expect(pauseBtn).toBeVisible();
       await expect(autoScrollBtn).toBeVisible();
       await expect(clearBtn).toBeVisible();
     });

     test('Logs filters are functional', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Test level filter can be changed
       const levelFilter = page.locator('#level-filter');
       await levelFilter.selectOption('error');
       await expect(levelFilter).toHaveValue('error');
       
       // Test service filter can be changed
       const serviceFilter = page.locator('#service-filter');
       await serviceFilter.selectOption('portal');
       await expect(serviceFilter).toHaveValue('portal');
       
       // Test search input accepts text
       const searchInput = page.locator('#search-input');
       await searchInput.fill('test search');
       await expect(searchInput).toHaveValue('test search');
     });

     test('Logs controls are functional', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Auto-scroll button should exist and be clickable
       const autoScrollBtn = page.locator('#auto-scroll-btn');
       await expect(autoScrollBtn).toBeVisible();
       
       // Button should be clickable
       await autoScrollBtn.click();
       await page.waitForTimeout(100);
       
       // Should still be visible after click
       await expect(autoScrollBtn).toBeVisible();
     });

     test('Logs pause/resume button works', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       const pauseBtn = page.locator('#pause-btn');
       
       // Button should exist and be clickable
       await expect(pauseBtn).toBeVisible();
       await pauseBtn.click();
       await page.waitForTimeout(100);
       
       // Button should still exist after click
       await expect(pauseBtn).toBeVisible();
     });

     test('Logs clear button works', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Wait for dashboard to load
       await page.waitForTimeout(500);
       
       const clearBtn = page.locator('#clear-btn');
       
       // Button should exist and be clickable
       await expect(clearBtn).toBeVisible();
       await clearBtn.click();
       
       // Button should still exist after click
       await expect(clearBtn).toBeVisible();
     });

     test('Logs WebSocket connection indicator exists', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       // Wait for elements to render
       await page.waitForTimeout(1000);
       
       const statusIndicator = page.locator('#connection-status');
       
       // Connection status indicator should be visible
       await expect(statusIndicator).toBeVisible();
     });

     // Test Analytics service access
     test('Analytics service is accessible via nginx proxy', async ({ page }) => {
       await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
       
       // Check for analytics dashboard
       const heading = page.locator('h1:has-text("Analytics")');
       await expect(heading).toBeVisible({ timeout: 5000 });
     });

     test('Analytics dashboard loads with charts', async ({ page }) => {
       await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
       
       // Wait for dashboard to load
       await page.waitForTimeout(1000);
       
       // Check for main dashboard elements - look for the chart container or other key elements
       const trends = page.locator('.trends-section, h2:has-text("Log Trends")').first();
       const anomalies = page.locator('.anomalies-section, h2:has-text("Detected Anomalies")').first();
       
       await expect(trends).toBeVisible();
       await expect(anomalies).toBeVisible();
     });

     // Cross-service navigation tests
     test('Can navigate between services', async ({ page }) => {
       // Start at Portal
       await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       
       // Check Portal is loaded with login button
       let loginBtn = page.locator('a.login-button').first();
       await expect(loginBtn).toBeVisible();
       
       // Navigate to Logs
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(500);
       
       const logsHeading = page.locator('h1:has-text("📝 DevSmith Logs")').first();
       await expect(logsHeading).toBeVisible();
       
       // Navigate to Analytics
       await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(500);
       
       const analyticsHeading = page.locator('h1:has-text("📊")').first();
       await expect(analyticsHeading).toBeVisible();
     });

     // Responsive design tests
     test('Logs dashboard is responsive on mobile', async ({ page }) => {
       await page.setViewportSize({ width: 375, height: 667 });
       
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(500);
       
       // All controls should still be visible
       const levelFilter = page.locator('#level-filter');
       const pauseBtn = page.locator('#pause-btn');
       const container = page.locator('.logs-container');
       
       await expect(levelFilter).toBeVisible();
       await expect(pauseBtn).toBeVisible();
       await expect(container).toBeVisible();
     });

     test('Logs dashboard is responsive on tablet', async ({ page }) => {
       await page.setViewportSize({ width: 768, height: 1024 });
       
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(1000);
       
       const container = page.locator('.logs-container');
       await expect(container).toBeVisible();
     });

     test('Logs dashboard is responsive on desktop', async ({ page }) => {
       await page.setViewportSize({ width: 1920, height: 1080 });
       
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(1000);
       
       const container = page.locator('.logs-container');
       await expect(container).toBeVisible();
     });

     // Performance tests
     test('Portal loads quickly', async ({ page }) => {
       const startTime = Date.now();
       await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       const loadTime = Date.now() - startTime;
       
       // Should load within 5 seconds
       expect(loadTime).toBeLessThan(5000);
     });

     test('Logs dashboard loads quickly', async ({ page }) => {
       const startTime = Date.now();
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       const loadTime = Date.now() - startTime;
       
       // Should load within 5 seconds
       expect(loadTime).toBeLessThan(5000);
     });

     test('Analytics dashboard loads quickly', async ({ page }) => {
       const startTime = Date.now();
       await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
       const loadTime = Date.now() - startTime;
       
       // Should load within 5 seconds
       expect(loadTime).toBeLessThan(5000);
     });

     // Error handling tests
     test('Handles navigation to non-existent pages gracefully', async ({ page }) => {
       const response = await page.goto('http://localhost:3000/nonexistent', { waitUntil: 'domcontentloaded' });
       
       // Should get 404 or redirect
       expect([404, 301, 302, 307, 308]).toContain(response?.status());
     });

     // Service health tests
     test('All services respond to health checks', async ({ page }) => {
       const healthChecks = [
         { name: 'Portal', url: 'http://localhost:3000/health' },
         { name: 'Logs', url: 'http://localhost:3000/logs/health' },
         { name: 'Analytics', url: 'http://localhost:3000/analytics/health' },
       ];

       for (const check of healthChecks) {
         const response = await page.goto(check.url, { waitUntil: 'domcontentloaded' });
         expect(response?.status(), `${check.name} health check failed`).toBe(200);
       }
     });

     // UI consistency tests
     test('Logs dashboard has consistent styling', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       
       const navbar = page.locator('nav.navbar').first();
       const container = page.locator('.logs-container').first();
       const filters = page.locator('.filters').first();
       
       await expect(navbar).toBeVisible();
       await expect(container).toBeVisible();
       await expect(filters).toBeVisible();
     });

     test('Logs output area displays correctly', async ({ page }) => {
       await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
       await page.waitForTimeout(500);
       
       const logsMain = page.locator('.logs-main').first();
       
       await expect(logsMain).toBeVisible();
     });
   });

   test.describe('DevSmith Platform - Service-to-Service Communication', () => {
     
     test('Services communicate through nginx reverse proxy', async ({ page }) => {
       // All services should be accessible through single nginx entry point
       const services = [
         { name: 'Portal', path: '/' },
         { name: 'Logs', path: '/logs/' },
         { name: 'Analytics', path: '/analytics/' },
       ];

       for (const service of services) {
         const response = await page.goto(`http://localhost:3000${service.path}`, { 
           waitUntil: 'domcontentloaded' 
         });
         expect(response?.status(), `${service.name} should be accessible`).toBe(200);
       }
     });

     test('Can perform rapid sequential navigation', async ({ page }) => {
       // Navigate between services multiple times rapidly
       for (let i = 0; i < 3; i++) {
         await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
         await page.waitForTimeout(500);
         
         await page.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' });
         await page.waitForTimeout(500);
         
         await page.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' });
         await page.waitForTimeout(500);
       }
       
       // Should not crash or hang
       const response = await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       expect(response?.status()).toBe(200);
     });

     test('Concurrent service access works', async ({ browser }) => {
       // Open multiple pages accessing different services simultaneously
       const page1 = await browser.newPage();
       const page2 = await browser.newPage();
       const page3 = await browser.newPage();

       try {
         await Promise.all([
           page1.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' }),
           page2.goto('http://localhost:3000/logs/', { waitUntil: 'domcontentloaded' }),
           page3.goto('http://localhost:3000/analytics/', { waitUntil: 'domcontentloaded' }),
         ]);

         // All pages should load successfully
         const title1 = await page1.title();
         const title2 = await page2.title();
         const title3 = await page3.title();

         expect(title1).toBeTruthy();
         expect(title2).toBeTruthy();
         expect(title3).toBeTruthy();
       } finally {
         await page1.close();
         await page2.close();
         await page3.close();
       }
     });
   });
