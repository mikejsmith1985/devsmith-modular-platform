import { test, expect } from '@playwright/test';

/**
 * OAuth Production Flow Test - Replicates ACTUAL user experience
 * 
 * This test does NOT mock anything - it tests the real OAuth flow
 * exactly as a user would experience it in production.
 */

test.describe('OAuth Production Flow - Real User Experience', () => {
  
  test('User clicks login button and initiates OAuth with state stored in Redis', async ({ page }) => {
    // STEP 1: User visits login page
    await page.goto('http://localhost:3000/login');
    await expect(page).toHaveURL('http://localhost:3000/login');
    
    // STEP 2: User clicks "Login with GitHub" button
    const loginButton = page.locator('a.btn-primary:has-text("Login with GitHub")');
    await expect(loginButton).toBeVisible();
    
    // STEP 3: Verify button goes to correct OAuth endpoint (not old /auth/login)
    const href = await loginButton.getAttribute('href');
    expect(href).toBe('/auth/github/login');
    console.log('✅ Login button uses correct endpoint:', href);
    
    // STEP 4: Click button and verify redirect to GitHub
    const [githubPage] = await Promise.all([
      page.waitForEvent('popup'),
      loginButton.click()
    ]);
    
    // STEP 5: Verify we're redirected to GitHub OAuth
    await githubPage.waitForLoadState('domcontentloaded');
    const githubURL = githubPage.url();
    expect(githubURL).toContain('github.com/login/oauth/authorize');
    expect(githubURL).toContain('client_id=');
    expect(githubURL).toContain('state=');
    
    // STEP 6: Extract state parameter from GitHub URL
    const urlObj = new URL(githubURL);
    const state = urlObj.searchParams.get('state');
    expect(state).toBeTruthy();
    expect(state!.length).toBeGreaterThan(20); // Should be 32-byte base64
    console.log('✅ OAuth state parameter:', state);
    
    // STEP 7: Verify state is in Redis (THIS IS THE CRITICAL CHECK)
    const { exec } = require('child_process');
    const redisCheck = await new Promise<string>((resolve) => {
      exec(
        `docker-compose exec -T redis redis-cli GET "oauth_state:${state}"`,
        { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
        (error: any, stdout: string) => {
          resolve(stdout.trim());
        }
      );
    });
    
    expect(redisCheck).toBe('"valid"');
    console.log('✅ State stored in Redis:', redisCheck);
    
    // STEP 8: Simulate OAuth callback (like GitHub would redirect back)
    await page.goto(`http://localhost:3000/auth/github/callback?code=test_fake_code&state=${state}`);
    
    // STEP 9: Should NOT get 401 Unauthorized - should get token exchange error
    const responseText = await page.textContent('body');
    expect(responseText).toContain('OAUTH_TOKEN_EXCHANGE_FAILED');
    expect(responseText).not.toContain('OAUTH_STATE_INVALID');
    expect(responseText).not.toContain('Unauthorized');
    console.log('✅ State validated successfully, reached token exchange');
    
    await githubPage.close();
  });
  
  test('Old /auth/login endpoint should NOT be used (deprecated)', async ({ page }) => {
    // This endpoint exists but doesn't store state - it's the bug!
    const response = await page.goto('http://localhost:3000/auth/login');
    
    // Should redirect to GitHub
    await page.waitForURL(/github\.com/);
    const url = page.url();
    
    // Check if state parameter is present
    const urlObj = new URL(url);
    const state = urlObj.searchParams.get('state');
    
    if (!state) {
      console.log('⚠️  OLD ENDPOINT USED - No state parameter!');
      console.log('⚠️  This is why production login fails!');
      expect(state).toBeTruthy(); // This SHOULD fail if old endpoint used
    } else {
      console.log('✅ State parameter present, but verify it was stored in Redis');
      
      // Check Redis
      const { exec } = require('child_process');
      const redisCheck = await new Promise<string>((resolve) => {
        exec(
          `docker-compose exec -T redis redis-cli GET "oauth_state:${state}"`,
          { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
          (error: any, stdout: string) => {
            resolve(stdout.trim());
          }
        );
      });
      
      expect(redisCheck).toBe('"valid"');
      console.log('✅ State verified in Redis');
    }
  });
  
  test('Multiple login attempts should each get unique states', async ({ page, context }) => {
    const states: string[] = [];
    
    // Attempt 1
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL(/github\.com/);
    const url1 = new URL(page.url());
    const state1 = url1.searchParams.get('state');
    expect(state1).toBeTruthy();
    states.push(state1!);
    
    // Attempt 2
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL(/github\.com/);
    const url2 = new URL(page.url());
    const state2 = url2.searchParams.get('state');
    expect(state2).toBeTruthy();
    states.push(state2!);
    
    // Attempt 3
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL(/github\.com/);
    const url3 = new URL(page.url());
    const state3 = url3.searchParams.get('state');
    expect(state3).toBeTruthy();
    states.push(state3!);
    
    // All states should be unique
    const uniqueStates = new Set(states);
    expect(uniqueStates.size).toBe(3);
    console.log('✅ All states are unique:', states);
    
    // All states should be in Redis
    const { exec } = require('child_process');
    for (const state of states) {
      const redisCheck = await new Promise<string>((resolve) => {
        exec(
          `docker-compose exec -T redis redis-cli GET "oauth_state:${state}"`,
          { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
          (error: any, stdout: string) => {
            resolve(stdout.trim());
          }
        );
      });
      expect(redisCheck).toBe('"valid"');
    }
    console.log('✅ All states verified in Redis');
  });
  
  test('State should expire after 10 minutes', async ({ page }) => {
    // Generate a state
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL(/github\.com/);
    const url = new URL(page.url());
    const state = url.searchParams.get('state');
    expect(state).toBeTruthy();
    
    // Check TTL in Redis
    const { exec } = require('child_process');
    const ttl = await new Promise<number>((resolve) => {
      exec(
        `docker-compose exec -T redis redis-cli TTL "oauth_state:${state}"`,
        { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
        (error: any, stdout: string) => {
          resolve(parseInt(stdout.trim()));
        }
      );
    });
    
    // TTL should be around 600 seconds (10 minutes)
    expect(ttl).toBeGreaterThan(550); // Allow some time for test execution
    expect(ttl).toBeLessThanOrEqual(600);
    console.log('✅ State TTL:', ttl, 'seconds');
  });
  
  test('Used state should be deleted from Redis (single-use)', async ({ page }) => {
    // Generate a state
    await page.goto('http://localhost:3000/auth/github/login');
    await page.waitForURL(/github\.com/);
    const url = new URL(page.url());
    const state = url.searchParams.get('state');
    expect(state).toBeTruthy();
    
    // Verify state exists in Redis
    const { exec } = require('child_process');
    let redisCheck = await new Promise<string>((resolve) => {
      exec(
        `docker-compose exec -T redis redis-cli GET "oauth_state:${state}"`,
        { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
        (error: any, stdout: string) => {
          resolve(stdout.trim());
        }
      );
    });
    expect(redisCheck).toBe('"valid"');
    
    // Use the state in callback
    await page.goto(`http://localhost:3000/auth/github/callback?code=test&state=${state}`);
    
    // Wait for callback to process
    await page.waitForLoadState('load');
    
    // Verify state is now deleted from Redis
    redisCheck = await new Promise<string>((resolve) => {
      exec(
        `docker-compose exec -T redis redis-cli GET "oauth_state:${state}"`,
        { cwd: '/home/mikej/projects/DevSmith-Modular-Platform' },
        (error: any, stdout: string) => {
          resolve(stdout.trim());
        }
      );
    });
    expect(redisCheck).toBe('(nil)');
    console.log('✅ State deleted after use (single-use validated)');
  });
});
