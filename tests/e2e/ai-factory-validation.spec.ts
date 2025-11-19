import { test, expect } from '@playwright/test';

/**
 * AI Factory Connection Test - Automated Validation
 * 
 * FIX VALIDATION FOR: Empty Ollama endpoint should use default http://localhost:11434
 * 
 * WHAT WAS FIXED:
 * - Frontend: Changed || undefined to || "" (always send endpoint field)
 * - Backend: Provide default http://localhost:11434 when endpoint is empty
 * - UI: Show default endpoint in placeholder
 * - Errors: Include troubleshooting suggestions
 * 
 * VALIDATION STRATEGY:
 * Since API requires authentication, we validate by:
 * 1. Checking code changes are deployed (services rebuilt)
 * 2. Verifying error message format has changed
 * 3. Manual testing required for full UI workflow
 */

test.describe('AI Factory - Fix Validation', () => {
  
  test('API accepts empty Ollama endpoint and provides helpful error', async ({ request }) => {
    // Test the actual API endpoint that was failing
    const response = await request.post('http://localhost:3000/api/portal/llm-configs/test', {
      data: {
        provider: 'ollama',
        model: 'qwen2.5-coder:7b',
        api_key: '',
        endpoint: ''  // Empty endpoint - should use default
      },
      headers: {
        'Content-Type': 'application/json'
      },
      failOnStatusCode: false  // Don't fail on 400, we want to inspect the response
    });
    
    const body = await response.json();
    
    // Should NOT return "Ollama endpoint is required" error
    expect(body.message).not.toContain('endpoint is required');
    
    // Should return connection failure (Ollama likely not running)
    // This is EXPECTED and CORRECT behavior
    expect(response.status()).toBe(400);
    expect(body.success).toBe(false);
    
    // Error message should include troubleshooting
    expect(body.details).toContain('Failed to connect to ollama');
    expect(body.details).toContain('Ensure Ollama is running');
    expect(body.details).toContain('http://localhost:11434');
    expect(body.details).toContain('curl');
    
    console.log('✅ API accepts empty endpoint and uses default http://localhost:11434');
    console.log('✅ Error message includes troubleshooting suggestions');
  });
  
  test('API accepts custom Ollama endpoint', async ({ request }) => {
    const response = await request.post('http://localhost:3000/api/portal/llm-configs/test', {
      data: {
        provider: 'ollama',
        model: 'qwen2.5-coder:7b',
        api_key: '',
        endpoint: 'http://custom-ollama:8080'
      },
      headers: {
        'Content-Type': 'application/json'
      },
      failOnStatusCode: false
    });
    
    const body = await response.json();
    
    // Should NOT return "endpoint is required"
    expect(body.message).not.toContain('endpoint is required');
    
    // Should attempt connection to custom endpoint
    if (response.status() === 400) {
      expect(body.details).toContain('custom-ollama:8080');
    }
    
    console.log('✅ API accepts custom endpoint');
  });
  
  test('Fix validation: Empty endpoint no longer causes HTTP 400 "endpoint required"', async ({ request }) => {
    // This is the exact scenario that was failing before the fix
    const response = await request.post('http://localhost:3000/api/portal/llm-configs/test', {
      data: {
        provider: 'ollama',
        model: 'qwen2.5-coder:7b',
        api_key: '',
        endpoint: ''
      },
      headers: {
        'Content-Type': 'application/json'
      },
      failOnStatusCode: false
    });
    
    const body = await response.json();
    
    // CRITICAL: This was the bug - should NOT return this error anymore
    if (body.message === 'Connection failed' && body.details?.includes('Ollama endpoint is required')) {
      throw new Error('BUG NOT FIXED: Still getting "Ollama endpoint is required" error');
    }
    
    // Success: Getting a different error (connection refused) means default endpoint is being used
    expect(body.details).toContain('Failed to connect to ollama');
    expect(body.details).toContain('localhost:11434');
    
    console.log('✅ FIX VERIFIED: No longer returns "endpoint is required" error');
    console.log('✅ Default endpoint http://localhost:11434 is being used');
  });
});
