import { test, expect } from '@playwright/test';

test.describe('Console Verification', () => {
  test('should have clean console on homepage', async ({ page }) => {
    const consoleMessages: any[] = [];
    const consoleErrors: any[] = [];
    
    // Capture console messages
    page.on('console', msg => {
      const type = msg.type();
      const text = msg.text();
      
      consoleMessages.push({ type, text });
      
      if (type === 'error' || type === 'warning') {
        consoleErrors.push({ type, text });
      }
    });

    // Navigate to homepage
    await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
    
    // Wait a bit for any deferred console messages
    await page.waitForTimeout(2000);
    
    // Log all messages for debugging
    console.log('\n=== Console Messages ===');
    consoleMessages.forEach(msg => {
      console.log(`[${msg.type}] ${msg.text}`);
    });
    console.log('========================\n');
    
    // Check for errors
    const errorCount = consoleErrors.length;
    if (errorCount > 0) {
      console.log('\n❌ Console Errors Found:');
      consoleErrors.forEach(err => {
        console.log(`  [${err.type}] ${err.text}`);
      });
    } else {
      console.log('✅ Console is clean - no errors or warnings');
    }
    
    // Assert no errors
    expect(errorCount, `Found ${errorCount} console errors/warnings`).toBe(0);
  });

  test('should have clean console on dashboard', async ({ page }) => {
    const consoleErrors: any[] = [];
    
    page.on('console', msg => {
      const type = msg.type();
      if (type === 'error' || type === 'warning') {
        consoleErrors.push({ type: type, text: msg.text() });
      }
    });

    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    
    if (consoleErrors.length > 0) {
      console.log('\n❌ Dashboard Console Errors:');
      consoleErrors.forEach(err => {
        console.log(`  [${err.type}] ${err.text}`);
      });
    } else {
      console.log('✅ Dashboard console is clean');
    }
    
    expect(consoleErrors.length, `Found ${consoleErrors.length} console errors/warnings on dashboard`).toBe(0);
  });

  test('should have clean console on review page', async ({ page }) => {
    const consoleErrors: any[] = [];
    
    page.on('console', msg => {
      const type = msg.type();
      if (type === 'error' || type === 'warning') {
        consoleErrors.push({ type: type, text: msg.text() });
      }
    });

    await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    
    if (consoleErrors.length > 0) {
      console.log('\n❌ Review Page Console Errors:');
      consoleErrors.forEach(err => {
        console.log(`  [${err.type}] ${err.text}`);
      });
    } else {
      console.log('✅ Review page console is clean');
    }
    
    expect(consoleErrors.length, `Found ${consoleErrors.length} console errors/warnings on review page`).toBe(0);
  });
});
