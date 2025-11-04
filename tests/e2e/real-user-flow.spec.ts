import { test, expect, Page } from '@playwright/test';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load test environment variables
dotenv.config({ path: path.join(__dirname, '../../.env.playwright') });

interface NetworkLog {
  type: 'request' | 'response';
  method?: string;
  url: string;
  status?: number;
  headers?: Record<string, string>;
  timestamp: string;
  redirectLocation?: string;
}

interface ConsoleLog {
  type: 'console' | 'error' | 'navigation';
  level?: string;
  text?: string;
  message?: string;
  url?: string;
  timestamp: string;
}

interface DebugSession {
  user_action: string;
  network_log: NetworkLog[];
  console_log: ConsoleLog[];
  page_errors: any[];
  navigation_events: any[];
}

/**
 * Capture all browser activity and optionally send to Logging service
 */
class BrowserDebugCapture {
  private networkLog: NetworkLog[] = [];
  private consoleLog: ConsoleLog[] = [];
  private pageErrors: any[] = [];
  private navigationEvents: any[] = [];
  
  constructor(private page: Page, private userAction: string) {}
  
  async start() {
    // Capture console messages
    this.page.on('console', msg => {
      this.consoleLog.push({
        type: 'console',
        level: msg.type(),
        text: msg.text(),
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture page errors
    this.page.on('pageerror', err => {
      const errorData = {
        message: err.message,
        stack: err.stack,
        timestamp: new Date().toISOString()
      };
      this.pageErrors.push(errorData);
      this.consoleLog.push({
        type: 'error',
        message: err.message,
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture network requests
    this.page.on('request', req => {
      this.networkLog.push({
        type: 'request',
        method: req.method(),
        url: req.url(),
        headers: req.headers(),
        timestamp: new Date().toISOString()
      });
    });
    
    // Capture network responses
    this.page.on('response', async res => {
      const entry: NetworkLog = {
        type: 'response',
        status: res.status(),
        url: res.url(),
        headers: res.headers(),
        timestamp: new Date().toISOString()
      };
      
      if (res.status() >= 300 && res.status() < 400) {
        entry.redirectLocation = res.headers()['location'];
      }
      
      this.networkLog.push(entry);
    });
    
    // Capture navigation
    this.page.on('framenavigated', frame => {
      if (frame === this.page.mainFrame()) {
        const navEvent = {
          url: frame.url(),
          timestamp: new Date().toISOString()
        };
        this.navigationEvents.push(navEvent);
        this.consoleLog.push({
          type: 'navigation',
          url: frame.url(),
          timestamp: new Date().toISOString()
        });
      }
    });
  }
  
  getDebugSession(): DebugSession {
    return {
      user_action: this.userAction,
      network_log: this.networkLog,
      console_log: this.consoleLog,
      page_errors: this.pageErrors,
      navigation_events: this.navigationEvents
    };
  }
  
  /**
   * Send debug session to Logging service
   */
  async sendToLoggingService(userId: number = 1) {
    const loggingApi = process.env.LOGGING_API || 'http://localhost:3000/api/logs';
    
    try {
      const response = await fetch(`${loggingApi}/browser-debug`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: userId,
          session_name: `Playwright: ${this.userAction}`,
          ...this.getDebugSession()
        })
      });
      
      if (!response.ok) {
        console.warn('Failed to send debug session to Logging service:', response.statusText);
      } else {
        console.log('✅ Debug session sent to Logging service');
      }
    } catch (error) {
      console.warn('Could not send debug session to Logging service:', error);
      // Don't fail the test if logging fails
    }
  }
  
  printSummary() {
    console.log('\n=== BROWSER DEBUG SUMMARY ===');
    console.log(`User Action: ${this.userAction}`);
    console.log(`Network Requests: ${this.networkLog.filter(l => l.type === 'request').length}`);
    console.log(`Network Responses: ${this.networkLog.filter(l => l.type === 'response').length}`);
    console.log(`Console Messages: ${this.consoleLog.filter(l => l.type === 'console').length}`);
    console.log(`Page Errors: ${this.pageErrors.length}`);
    console.log(`Navigation Events: ${this.navigationEvents.length}`);
    
    if (this.pageErrors.length > 0) {
      console.log('\n⚠️  PAGE ERRORS:');
      this.pageErrors.forEach(err => {
        console.log(`  - ${err.message}`);
      });
    }
    
    const reviewRequests = this.networkLog.filter(l => 
      l.type === 'request' && l.url.includes('/review')
    );
    if (reviewRequests.length > 0) {
      console.log('\n🔍 Review-related requests:');
      reviewRequests.forEach(req => {
        console.log(`  ${req.method} ${req.url}`);
      });
    }
    
    const redirects = this.networkLog.filter(l => 
      l.type === 'response' && l.status && l.status >= 300 && l.status < 400
    );
    if (redirects.length > 0) {
      console.log('\n↪️  Redirects:');
      redirects.forEach(res => {
        console.log(`  ${res.status} ${res.url} → ${res.redirectLocation}`);
      });
    }
    
    console.log('============================\n');
  }
}

test.describe('Portal to Review Navigation - Real User Flow', () => {
  test('User logs in with GitHub and navigates to Review app', async ({ page, context }) => {
    const baseUrl = process.env.BASE_URL || 'http://localhost:3000';
    
    // Start debug capture
    const debugCapture = new BrowserDebugCapture(page, 'Login and click Review card');
    await debugCapture.start();
    
    console.log('\n=== STEP 1: Navigate to home page ===');
    await page.goto(baseUrl, { waitUntil: 'networkidle' });
    
    // Check if already logged in (cookie exists)
    const cookies = await context.cookies();
    const hasAuthCookie = cookies.some(c => c.name === 'devsmith_token');
    
    if (!hasAuthCookie) {
      console.log('Not logged in - starting OAuth flow...');
      
  // Click "Login with GitHub" button (match a few possible hrefs)
  console.log('\n=== STEP 2: Click "Login with GitHub" ===');
  // Some deployments use /auth/login, others use /auth/github/login — match both.
  const loginButton = page.locator('a[href*="/auth/github/login"], a[href="/auth/login"], button:has-text("Login with GitHub")');
      await expect(loginButton).toBeVisible({ timeout: 5000 });
      await loginButton.click();
      
      // Wait for GitHub OAuth page or redirect back
      console.log('Waiting for GitHub OAuth page...');
      
      // Check if we're on GitHub's login page
      await page.waitForTimeout(2000);
      const currentUrl = page.url();
      
      if (currentUrl.includes('github.com/login')) {
        console.log('\n=== GitHub Login Required ===');
        console.log('To complete this test, you need to:');
        console.log('1. Log in to GitHub in this browser');
        console.log('2. Authorize the DevSmith app');
        console.log('\nAlternatively, set GITHUB_TEST_USERNAME and GITHUB_TEST_PASSWORD in .env.playwright');
        console.log('for automated login.\n');
        
        // Check if test credentials are available
        const username = process.env.GITHUB_TEST_USERNAME;
        const password = process.env.GITHUB_TEST_PASSWORD;
        
        if (username && password) {
          console.log('Using test credentials for automated login...');
          
          // Fill in GitHub login form
          await page.fill('input[name="login"]', username);
          await page.fill('input[name="password"]', password);

          // Try a few common submit selectors (GitHub varies)
          const submitSelectors = [
            'input[type="submit"][value="Sign in"]',
            'button[name="commit"]',
            'button:has-text("Sign in")',
            'input[type="submit"]'
          ];

          let submitted = false;
          for (const sel of submitSelectors) {
            const el = page.locator(sel);
            if (await el.count() > 0) {
              try {
                await el.first().click();
                submitted = true;
                break;
              } catch (err) {
                // ignore and try next
              }
            }
          }

          if (!submitted) {
            // As a last resort, press Enter in the password field
            await page.press('input[name="password"]', 'Enter');
          }

          // Wait for redirect back to our app (give more time for multi-step login)
          await page.waitForURL(`${baseUrl}/**`, { timeout: 30000 });

          // Handle optional OTP (2FA) step if prompted
          const otpInput = page.locator('input[name="otp"]');
          if (await otpInput.count() > 0) {
            const otp = process.env.GITHUB_TEST_OTP;
            if (otp) {
              console.log('Filling OTP from env...');
              await otpInput.fill(otp);
              // Submit OTP
              await Promise.all([
                page.waitForNavigation({ waitUntil: 'networkidle', timeout: 30000 }),
                page.click('button:has-text("Verify")').catch(() => page.press('input[name="otp"]', 'Enter'))
              ]);
            } else {
              console.log('OTP required but GITHUB_TEST_OTP not set. Pausing for manual entry (120s)...');
              // Allow manual OTP entry
              await page.waitForURL(`${baseUrl}/**`, { timeout: 120000 });
            }
          }
        } else {
          // Manual intervention needed - pause for user to log in
          console.log('⏸️  Pausing for manual GitHub login...');
          console.log('Press Ctrl+C if you want to skip this test.\n');
          
          // Wait for callback (user manually logs in)
          await page.waitForURL(`${baseUrl}/**`, { timeout: 60000 });
        }
      }
      
      console.log('✅ OAuth flow completed');
    } else {
      console.log('✅ Already logged in (cookie found)');
    }
    
    // Should now be on dashboard
    console.log('\n=== STEP 3: Verify we\'re on dashboard ===');
    await page.waitForURL(`${baseUrl}/dashboard`, { timeout: 10000 });
    await expect(page).toHaveURL(`${baseUrl}/dashboard`);
    console.log('✅ On dashboard:', page.url());
    
    // Wait for page to fully load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Find Review card
    console.log('\n=== STEP 4: Find Review card ===');
    const reviewCard = page.locator('a[href="/review"]').first();
    await expect(reviewCard).toBeVisible({ timeout: 5000 });
    
    const cardText = await reviewCard.textContent();
    console.log('Found Review card:', cardText?.substring(0, 50));
    
    // Capture URL before clicking
    const urlBefore = page.url();
    console.log('URL before click:', urlBefore);
    
    // Click Review card
    console.log('\n=== STEP 5: Click Review card ===');
    await reviewCard.click();
    
    // Wait for navigation
    await page.waitForLoadState('networkidle', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    // Check where we ended up
    const urlAfter = page.url();
    console.log('URL after click:', urlAfter);
    
    // Print debug summary
    debugCapture.printSummary();
    
    // Send to logging service (don't fail test if this fails)
    await debugCapture.sendToLoggingService(1);
    
    // Verify we're on Review app (not redirected back to GitHub)
    console.log('\n=== STEP 6: Verify navigation ===');
    
    if (urlAfter.includes('github.com')) {
      console.log('❌ FAILED: Redirected to GitHub OAuth');
      console.log('This means the auth cookie was not sent with the /review request');
      
      // Check cookies
      const finalCookies = await context.cookies();
      console.log('\nCookies present:');
      finalCookies.forEach(c => {
        console.log(`  - ${c.name}: sameSite=${c.sameSite}, path=${c.path}, domain=${c.domain}`);
      });
      
      throw new Error('Navigation to Review failed - redirected to GitHub OAuth (cookie not sent)');
    } else if (urlAfter.includes('/review')) {
      console.log('✅ SUCCESS: On Review app');
      expect(urlAfter).toContain('/review');
    } else {
      console.log('⚠️  Unexpected URL:', urlAfter);
      throw new Error(`Unexpected navigation target: ${urlAfter}`);
    }
    
    // Verify Review page content loaded
    await expect(page.locator('body')).toContainText('Review', { timeout: 5000 });
    console.log('✅ Review page content verified');
    
    console.log('\n=== TEST COMPLETED SUCCESSFULLY ===\n');
  });
});
