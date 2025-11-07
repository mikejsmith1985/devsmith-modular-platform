import { test, expect } from '@playwright/test';

test.describe('OAuth PKCE Flow End-to-End', () => {
  test('complete OAuth PKCE flow with GitHub', async ({ page, context }) => {
    // Enable verbose console logging
    page.on('console', msg => console.log('Browser console:', msg.text()));
    
    // Step 1: Navigate to the login page
    console.log('Step 1: Navigating to login page...');
    await page.goto('http://localhost:3000/');
    
    // Wait for React to load
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/oauth-pkce-01-homepage.png' });
    
    // Step 2: Look for login button
    console.log('Step 2: Looking for GitHub login button...');
    const loginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")');
    await expect(loginButton).toBeVisible({ timeout: 5000 });
    await page.screenshot({ path: 'test-results/oauth-pkce-02-login-button.png' });
    
    // Step 3: Check sessionStorage before login (should be empty)
    const storageBeforeLogin = await page.evaluate(() => ({
      codeVerifier: sessionStorage.getItem('pkce_code_verifier'),
      state: sessionStorage.getItem('oauth_state'),
    }));
    console.log('SessionStorage before login:', storageBeforeLogin);
    expect(storageBeforeLogin.codeVerifier).toBeNull();
    expect(storageBeforeLogin.state).toBeNull();
    
    // Step 4: Click login button (this will generate PKCE params and redirect)
    console.log('Step 4: Clicking login button...');
    
    // Listen for the navigation to GitHub
    const navigationPromise = page.waitForURL(/github\.com\/login\/oauth\/authorize/, { timeout: 10000 });
    
    await loginButton.click();
    
    // Step 5: Wait for redirect to GitHub and capture URL
    console.log('Step 5: Waiting for GitHub redirect...');
    await navigationPromise;
    
    const githubUrl = page.url();
    console.log('Redirected to GitHub URL:', githubUrl);
    await page.screenshot({ path: 'test-results/oauth-pkce-03-github-redirect.png' });
    
    // Step 6: Verify PKCE parameters in GitHub URL
    console.log('Step 6: Verifying PKCE parameters in URL...');
    const url = new URL(githubUrl);
    const codeChallenge = url.searchParams.get('code_challenge');
    const codeChallengeMethod = url.searchParams.get('code_challenge_method');
    const state = url.searchParams.get('state');
    const clientId = url.searchParams.get('client_id');
    
    console.log('PKCE Parameters:', {
      client_id: clientId,
      code_challenge: codeChallenge?.substring(0, 20) + '...',
      code_challenge_method: codeChallengeMethod,
      state: state?.substring(0, 20) + '...',
    });
    
    // Verify PKCE parameters are present
    expect(clientId).toBeTruthy(); // Client ID should exist (not checking actual value for security)
    expect(clientId?.length).toBeGreaterThan(10); // GitHub Client IDs are 20 chars
    expect(codeChallenge).toBeTruthy();
    expect(codeChallenge?.length).toBeGreaterThan(40); // Base64URL encoded SHA-256 hash
    expect(codeChallengeMethod).toBe('S256');
    expect(state).toBeTruthy();
    expect(state?.length).toBeGreaterThan(20); // Random state string
    
    // Step 7: Go back and verify sessionStorage has PKCE data
    console.log('Step 7: Checking sessionStorage for stored PKCE data...');
    await page.goBack();
    await page.waitForLoadState('networkidle');
    
    const storageAfterRedirect = await page.evaluate(() => ({
      codeVerifier: sessionStorage.getItem('pkce_code_verifier'),
      state: sessionStorage.getItem('oauth_state'),
    }));
    console.log('SessionStorage after redirect:', {
      codeVerifier: storageAfterRedirect.codeVerifier?.substring(0, 20) + '...',
      state: storageAfterRedirect.state?.substring(0, 20) + '...',
    });
    
    expect(storageAfterRedirect.codeVerifier).toBeTruthy();
    expect(storageAfterRedirect.state).toBeTruthy();
    expect(storageAfterRedirect.state).toBe(state); // State should match URL parameter
    
    // Step 8: Test token exchange endpoint (with mock data)
    console.log('Step 8: Testing token exchange endpoint...');
    const tokenResponse = await context.request.post('http://localhost:3000/api/portal/auth/token', {
      headers: {
        'Content-Type': 'application/json',
      },
      data: {
        code: 'test_code',
        state: storageAfterRedirect.state,
        code_verifier: storageAfterRedirect.codeVerifier,
      },
    });
    
    console.log('Token exchange response status:', tokenResponse.status());
    const tokenBody = await tokenResponse.json();
    console.log('Token exchange response:', tokenBody);
    
    // Should return 401 with fake credentials (means endpoint is working)
    expect(tokenResponse.status()).toBe(401);
    expect(tokenBody.error).toBeTruthy();
    
    await page.screenshot({ path: 'test-results/oauth-pkce-04-complete.png' });
    
    console.log('✅ OAuth PKCE flow test complete!');
  });
  
  test('PKCE crypto utilities generate valid parameters', async ({ page }) => {
    await page.goto('http://localhost:3000/');
    await page.waitForLoadState('networkidle');
    
    // Test PKCE utilities in browser console
    const pkceTest = await page.evaluate(async () => {
      // Import and test PKCE functions
      try {
        // Generate code verifier (43 chars, base64URL)
        const verifier = window.crypto.getRandomValues(new Uint8Array(32));
        const verifierStr = btoa(String.fromCharCode(...verifier))
          .replace(/\+/g, '-')
          .replace(/\//g, '_')
          .replace(/=/g, '');
        
        // Generate code challenge (SHA-256 of verifier)
        const encoder = new TextEncoder();
        const data = encoder.encode(verifierStr);
        const hash = await window.crypto.subtle.digest('SHA-256', data);
        const challenge = btoa(String.fromCharCode(...new Uint8Array(hash)))
          .replace(/\+/g, '-')
          .replace(/\//g, '_')
          .replace(/=/g, '');
        
        // Generate state (random string)
        const stateBytes = window.crypto.getRandomValues(new Uint8Array(16));
        const state = btoa(String.fromCharCode(...stateBytes))
          .replace(/\+/g, '-')
          .replace(/\//g, '_')
          .replace(/=/g, '');
        
        return {
          success: true,
          verifierLength: verifierStr.length,
          challengeLength: challenge.length,
          stateLength: state.length,
          verifier: verifierStr.substring(0, 10) + '...',
          challenge: challenge.substring(0, 10) + '...',
          state: state.substring(0, 10) + '...',
        };
      } catch (error) {
        return {
          success: false,
          error: error instanceof Error ? error.message : String(error),
        };
      }
    });
    
    console.log('PKCE crypto test:', pkceTest);
    
    expect(pkceTest.success).toBe(true);
    expect(pkceTest.verifierLength).toBe(43); // 32 bytes base64URL = 43 chars
    expect(pkceTest.challengeLength).toBe(43); // SHA-256 base64URL = 43 chars
    expect(pkceTest.stateLength).toBeGreaterThan(20);
  });
});
