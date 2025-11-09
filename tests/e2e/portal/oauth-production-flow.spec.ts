import { test, expect } from '@playwright/test';

/**
 * OAuth Production Flow Test - Tests the ACTUAL bug that was breaking login
 */

test.describe('OAuth Production Flow', () => {
  
  test('Login button uses correct OAuth endpoint with state management', async ({ page }) => {
    // Visit login page
    await page.goto('http://localhost:3000/login');
    
    // Check the login button href
    const loginButton = page.locator('a:has-text("Login with GitHub")');
    await expect(loginButton).toBeVisible();
    
    const href = await loginButton.getAttribute('href');
    
    // CRITICAL: Must be /auth/github/login (with state), NOT /auth/login (without state)
    expect(href).toBe('/auth/github/login');
    console.log('✅ Login button uses correct endpoint:', href);
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
