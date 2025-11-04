import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test.describe('Review Card Click - Complete Debugging', () => {
  test('capture all network activity when clicking Review card', async ({ page, context }) => {
    const debugLog: any[] = [];
    const networkLog: any[] = [];
    
    // Capture console messages
    page.on('console', msg => {
      debugLog.push({
        type: 'console',
        level: msg.type(),
        text: msg.text(),
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture page errors
    page.on('pageerror', err => {
      debugLog.push({
        type: 'error',
        message: err.message,
        stack: err.stack,
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture all network requests
    page.on('request', req => {
      networkLog.push({
        type: 'request',
        method: req.method(),
        url: req.url(),
        headers: req.headers(),
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture all network responses
    page.on('response', async res => {
      const entry: any = {
        type: 'response',
        status: res.status(),
        statusText: res.statusText(),
        url: res.url(),
        headers: res.headers(),
        timestamp: new Date().toISOString()
      };
      
      // Capture redirect location
      if (res.status() >= 300 && res.status() < 400) {
        entry.redirectLocation = res.headers()['location'];
      }
      
      // Try to capture response body for small responses
      try {
        const contentType = res.headers()['content-type'] || '';
        if (contentType.includes('json') || contentType.includes('text')) {
          entry.body = await res.text();
        }
      } catch (e) {
        // Ignore if we can't read body
      }
      
      networkLog.push(entry);
    });
    
    // Capture navigation events
    page.on('framenavigated', frame => {
      if (frame === page.mainFrame()) {
        debugLog.push({
          type: 'navigation',
          url: frame.url(),
          timestamp: new Date().toISOString()
        });
      }
    });
    
    console.log('\n=== STARTING TEST ===\n');
    
    // Step 1: Go to login page
    console.log('Step 1: Navigate to root...');
    await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
    debugLog.push({ type: 'step', message: 'Navigated to root', url: page.url() });
    
    // Step 2: Simulate authentication by setting a cookie
    // NOTE: We're using the shared secret "your-secret-key" from the codebase
    // to create a valid JWT for testing
    const crypto = require('crypto');
    
    // Create a simple JWT manually (matching the app's secret)
    const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
    const payload = Buffer.from(JSON.stringify({
      user_id: 1,
      username: 'testuser',
      exp: Math.floor(Date.now() / 1000) + 3600 // 1 hour from now
    })).toString('base64url');
    
    const secret = 'your-secret-key'; // From the codebase
    const signature = crypto
      .createHmac('sha256', secret)
      .update(`${header}.${payload}`)
      .digest('base64url');
    
    const validJWT = `${header}.${payload}.${signature}`;
    
    await context.addCookies([{
      name: 'devsmith_token',
      value: validJWT,
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      secure: false,
      sameSite: 'Lax'
    }]);
    
    debugLog.push({ type: 'step', message: 'Added valid JWT cookie' });
    console.log('Step 2: Added authentication cookie');
    
    // Step 3: Navigate to dashboard
    console.log('Step 3: Navigate to dashboard...');
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    debugLog.push({ type: 'step', message: 'Navigated to dashboard', url: page.url() });
    
    await page.waitForTimeout(1000);
    
    // Verify we're on dashboard
    const isDashboard = page.url().includes('/dashboard');
    debugLog.push({ type: 'check', message: 'On dashboard?', value: isDashboard });
    console.log('On dashboard:', isDashboard);
    
    if (!isDashboard) {
      const html = await page.content();
      debugLog.push({ type: 'error', message: 'Not on dashboard', html: html.substring(0, 500) });
      throw new Error('Failed to reach dashboard - check debug output');
    }
    
    // Step 4: Find Review card
    console.log('Step 4: Looking for Review card...');
    const reviewCard = page.locator('a[href="/review"]').first();
    const cardCount = await reviewCard.count();
    debugLog.push({ type: 'check', message: 'Review cards found', count: cardCount });
    console.log('Review cards found:', cardCount);
    
    if (cardCount === 0) {
      const bodyText = await page.textContent('body');
      debugLog.push({ type: 'error', message: 'No Review card', bodyText: bodyText?.substring(0, 500) });
      throw new Error('Review card not found - check debug output');
    }
    
    await reviewCard.scrollIntoViewIfNeeded();
    
    const cardHref = await reviewCard.getAttribute('href');
    const cardVisible = await reviewCard.isVisible();
    debugLog.push({ type: 'check', message: 'Review card details', href: cardHref, visible: cardVisible });
    console.log('Review card href:', cardHref, 'visible:', cardVisible);
    
    // Step 5: Click the Review card and monitor everything
    console.log('\nStep 5: CLICKING REVIEW CARD...\n');
    debugLog.push({ type: 'step', message: 'About to click Review card' });
    
    const urlBeforeClick = page.url();
    
    // Click and wait for navigation or timeout
    try {
      await Promise.race([
        reviewCard.click(),
        page.waitForNavigation({ timeout: 5000 }).catch(() => {})
      ]);
    } catch (e) {
      debugLog.push({ type: 'error', message: 'Click or navigation failed', error: String(e) });
    }
    
    // Wait a moment for any delayed navigation
    await page.waitForTimeout(2000);
    
    const urlAfterClick = page.url();
    debugLog.push({ 
      type: 'result', 
      message: 'After click',
      urlBefore: urlBeforeClick,
      urlAfter: urlAfterClick,
      navigationOccurred: urlBeforeClick !== urlAfterClick
    });
    
    console.log('\n=== CLICK RESULTS ===');
    console.log('URL before:', urlBeforeClick);
    console.log('URL after:', urlAfterClick);
    console.log('Navigation occurred:', urlBeforeClick !== urlAfterClick);
    
    // Save all logs to file
    const outputDir = './test-results';
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const debugFile = path.join(outputDir, `review-click-debug-${timestamp}.json`);
    
    fs.writeFileSync(debugFile, JSON.stringify({
      summary: {
        urlBefore: urlBeforeClick,
        urlAfter: urlAfterClick,
        navigationOccurred: urlBeforeClick !== urlAfterClick
      },
      debugLog,
      networkLog: networkLog.filter(n => 
        n.url.includes('dashboard') || 
        n.url.includes('review') || 
        n.url.includes('auth')
      )
    }, null, 2));
    
    console.log('\n=== DEBUG OUTPUT SAVED ===');
    console.log('File:', debugFile);
    console.log('\nKey network requests:');
    
    // Print relevant network activity
    const relevantRequests = networkLog.filter(n => 
      n.url.includes('dashboard') || 
      n.url.includes('review') || 
      n.url.includes('auth')
    );
    
    relevantRequests.forEach((entry, i) => {
      if (entry.type === 'request') {
        console.log(`\n[${i}] REQUEST: ${entry.method} ${entry.url}`);
      } else if (entry.type === 'response') {
        console.log(`[${i}] RESPONSE: ${entry.status} ${entry.url}`);
        if (entry.redirectLocation) {
          console.log(`    → Redirect to: ${entry.redirectLocation}`);
        }
        if (entry.body && entry.body.length < 200) {
          console.log(`    Body: ${entry.body}`);
        }
      }
    });
    
    // Final assertion
    if (urlBeforeClick === urlAfterClick) {
      console.log('\n❌ ISSUE REPRODUCED: Click did not cause navigation');
      console.log('The page "flashed and stayed on dashboard" - check debug file for details');
    } else {
      console.log('\n✅ Navigation worked - moved from dashboard to:', urlAfterClick);
    }
    
    // Keep test browser open if navigation failed so we can inspect
    if (urlBeforeClick === urlAfterClick) {
      await page.pause(); // This will keep browser open for manual inspection
    }
  });
});
