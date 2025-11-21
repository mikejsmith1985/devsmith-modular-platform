import { test, expect } from '@playwright/test';

test.describe('OAuth Flow', () => {
  test('should generate state, store in Redis, and validate callback', async ({ page, request }) => {
    // Step 1: Initiate OAuth login and capture redirect URL
    const loginResponse = await request.get('/auth/github/login', {
      maxRedirects: 0,
    });
    
    expect(loginResponse.status()).toBe(302);
    
    const locationHeader = loginResponse.headers()['location'];
    expect(locationHeader).toContain('github.com/login/oauth/authorize');
    expect(locationHeader).toContain('client_id=');
    expect(locationHeader).toContain('state=');
    expect(locationHeader).toContain('prompt=consent');
    
    // Extract state parameter from GitHub redirect URL
    const stateMatch = locationHeader.match(/state=([^&]+)/);
    expect(stateMatch).not.toBeNull();
    const stateEncoded = stateMatch![1];
    
    console.log(`State in GitHub redirect URL (URL-encoded): ${stateEncoded}`);
    
    // Decode the state to get what's stored in Redis
    const stateDecoded = decodeURIComponent(stateEncoded);
    console.log(`State decoded (stored in Redis): ${stateDecoded}`);
    
    // Verify state has = padding (base64)
    expect(stateDecoded).toMatch(/=$/);
    
    // Step 2: Simulate GitHub callback with the encoded state
    // GitHub will send the state URL-encoded, Gin will decode it
    const callbackResponse = await request.get(
      `/auth/github/callback?code=fake_code_for_testing&state=${stateEncoded}`
    );
    
    const callbackBody = await callbackResponse.json();
    console.log('Callback response:', JSON.stringify(callbackBody, null, 2));
    
    // We expect "Failed to exchange code for token" because we're using a fake code
    // But state validation should PASS (not "Invalid OAuth state parameter")
    expect(callbackBody.error).toBe('Failed to exchange code for token');
    expect(callbackBody.error_code).toBe('OAUTH_TOKEN_EXCHANGE_FAILED');
    
    // Should NOT be state validation error
    expect(callbackBody.error).not.toContain('state');
    expect(callbackBody.error_code).not.toBe('OAUTH_STATE_INVALID');
  });
  
  test('should reject callback with invalid state', async ({ request }) => {
    // Test with completely fake state that doesn't exist in Redis
    const callbackResponse = await request.get(
      '/auth/github/callback?code=fake_code&state=invalid_state_12345'
    );
    
    const callbackBody = await callbackResponse.json();
    
    // Should fail with state validation error
    expect(callbackResponse.status()).toBe(401);
    expect(callbackBody.error).toContain('OAuth state');
    expect(callbackBody.action).toContain('OAUTH_STATE_INVALID');
  });
  
  test('should reject callback with missing state', async ({ request }) => {
    // Test with no state parameter
    const callbackResponse = await request.get(
      '/auth/github/callback?code=fake_code'
    );
    
    const callbackBody = await callbackResponse.json();
    
    // Should fail with missing state error
    expect(callbackResponse.status()).toBe(400);
    expect(callbackBody.error).toContain('state');
    expect(callbackBody.action).toContain('OAUTH_STATE_MISSING');
  });
  
  test('should reject reused state (one-time use CSRF protection)', async ({ request }) => {
    // Step 1: Generate fresh state
    const loginResponse = await request.get('/auth/github/login', {
      maxRedirects: 0,
    });
    
    const locationHeader = loginResponse.headers()['location'];
    const stateMatch = locationHeader.match(/state=([^&]+)/);
    const stateEncoded = stateMatch![1];
    
    // Step 2: Use state once (should succeed with token exchange error)
    const firstCallback = await request.get(
      `/auth/github/callback?code=fake_code&state=${stateEncoded}`
    );
    
    const firstBody = await firstCallback.json();
    expect(firstBody.error).toBe('Failed to exchange code for token');
    
    // Step 3: Try to reuse same state (should fail - state deleted after first use)
    const secondCallback = await request.get(
      `/auth/github/callback?code=fake_code&state=${stateEncoded}`
    );
    
    const secondBody = await secondCallback.json();
    expect(secondCallback.status()).toBe(401);
    expect(secondBody.error).toContain('OAuth state');
    expect(secondBody.action).toContain('OAUTH_STATE_INVALID');
  });
});
