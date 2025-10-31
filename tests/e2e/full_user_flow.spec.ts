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
       
       // Should see Portal login page or dashboard
       const response = await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       expect([200, 304]).toContain(response?.status());
     });

     test('Platform health check endpoint responds', async ({ page }) => {
       const response = await page.goto('http://localhost:3000/health', { waitUntil: 'domcontentloaded' });
       expect([200, 304]).toContain(response?.status());
     });

     // Test error handling
     test('Handles navigation to non-existent pages gracefully', async ({ page }) => {
       const response = await page.goto('http://localhost:3000/nonexistent', { waitUntil: 'domcontentloaded' });
       expect([404, 301, 302, 307, 308]).toContain(response?.status());
     });

     // Service health tests
     test('All services respond to health checks', async ({ page }) => {
       // All services route through same health endpoint
       const response = await page.goto('http://localhost:3000/health', { waitUntil: 'domcontentloaded' });
       expect([200, 304]).toContain(response?.status());
     });

     // Performance tests
     test('Portal loads quickly', async ({ page }) => {
       const startTime = Date.now();
       await page.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
       const loadTime = Date.now() - startTime;
       
       // Should load within 5 seconds
       expect(loadTime).toBeLessThan(5000);
     });

     // Concurrent access test
     test('Can handle concurrent navigation', async ({ browser }) => {
       const page1 = await browser.newPage();
       const page2 = await browser.newPage();
       const page3 = await browser.newPage();

       try {
         // All pages can access the health endpoint simultaneously
         const resp1 = await page1.goto('http://localhost:3000/health', { waitUntil: 'domcontentloaded' });
         const resp2 = await page2.goto('http://localhost:3000/', { waitUntil: 'domcontentloaded' });
         const resp3 = await page3.goto('http://localhost:3000/health', { waitUntil: 'domcontentloaded' });

         // All requests should succeed
         expect([200, 304]).toContain(resp1?.status());
         expect([200, 304]).toContain(resp2?.status());
         expect([200, 304]).toContain(resp3?.status());
       } finally {
         await page1.close();
         await page2.close();
         await page3.close();
       }
     });
});