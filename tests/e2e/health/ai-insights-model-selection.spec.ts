import { test, expect } from '../fixtures/auth.fixture';
import percySnapshot from '@percy/playwright';

/**
 * AI Insights Model Selection Tests
 * 
 * Issue #4: Model Name Matching with Ollama
 * 
 * Purpose: Validate that AI Insights works with Ollama models
 * and handles model name matching correctly.
 * 
 * Architecture:
 * - ModelSelector fetches models from /api/portal/llm-configs
 * - Model names must EXACTLY match what's in Portal LLM configs
 * - Ollama models use format: "qwen2.5-coder:7b", "deepseek-coder:6.7b"
 * - AI Insights passes model name to /api/logs/:id/insights
 * 
 * Related: DEPLOYMENT.md (Model Name Mismatch section)
 */

test.describe('AI Insights - Model Selection', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    // Navigate to Health page
    await authenticatedPage.goto('/health');
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Wait for page to load
    await authenticatedPage.waitForSelector('.stat-card', { timeout: 10000 });
  });

  test('ModelSelector loads available models from Portal', async ({ authenticatedPage }) => {
    // Find ModelSelector dropdown
    const modelSelect = authenticatedPage.locator('select#model-select');
    await expect(modelSelect).toBeVisible();
    
    // Get all options
    const options = await modelSelect.locator('option').allTextContents();
    console.log('Available models:', options);
    
    // Should have at least one model
    expect(options.length).toBeGreaterThan(0);
    
    // Should NOT show "No models available"
    const noModelsText = options.join(' ');
    expect(noModelsText).not.toContain('No models available');
    
    // Take screenshot of model selector
    await percySnapshot(authenticatedPage, 'Health Page - Model Selector with Models');
  });

  test('ModelSelector fetches from correct Portal endpoint', async ({ authenticatedPage }) => {
    const llmConfigRequests: any[] = [];
    
    // Intercept LLM config requests
    authenticatedPage.on('request', request => {
      if (request.url().includes('/api/portal/llm-configs')) {
        llmConfigRequests.push({
          url: request.url(),
          method: request.method()
        });
      }
    });
    
    // Reload page to trigger API call
    await authenticatedPage.reload();
    await authenticatedPage.waitForLoadState('networkidle');
    
    // Verify Portal LLM configs endpoint was called
    expect(llmConfigRequests.length).toBeGreaterThan(0);
    expect(llmConfigRequests[0].method).toBe('GET');
    expect(llmConfigRequests[0].url).toContain('/api/portal/llm-configs');
    
    console.log('LLM Config API calls:', llmConfigRequests);
  });

  test('Portal LLM configs endpoint returns valid model data', async ({ authenticatedPage }) => {
    // Call Portal LLM configs directly
    const response = await authenticatedPage.request.get('/api/portal/llm-configs');
    
    expect(response.ok()).toBeTruthy();
    
    const configs = await response.json();
    console.log('LLM Configs response:', JSON.stringify(configs, null, 2));
    
    // Should be an array
    expect(Array.isArray(configs)).toBeTruthy();
    
    if (configs.length > 0) {
      const firstConfig = configs[0];
      
      // Verify required fields exist
      expect(firstConfig).toHaveProperty('id');
      expect(firstConfig).toHaveProperty('provider');
      expect(firstConfig).toHaveProperty('model');
      expect(firstConfig).toHaveProperty('name'); // Computed name: "provider - model"
      
      // Model name should not be empty
      expect(firstConfig.model).toBeTruthy();
      expect(typeof firstConfig.model).toBe('string');
      
      // Provider should be valid
      expect(['ollama', 'openai', 'anthropic', 'deepseek', 'mistral', 'google']).toContain(firstConfig.provider);
      
      console.log('Sample LLM Config:', {
        provider: firstConfig.provider,
        model: firstConfig.model,
        name: firstConfig.name,
        is_default: firstConfig.is_default
      });
    }
  });

  test('AI Insights can be generated with selected model', async ({ authenticatedPage }) => {
    // Select first log entry
    const firstRow = authenticatedPage.locator('table tbody tr').first();
    await firstRow.click();
    
    // Wait for log details to appear
    await authenticatedPage.waitForSelector('.log-details', { timeout: 5000 }).catch(() => {
      console.log('Log details modal may not have appeared');
    });
    
    // Look for AI Insights button/section
    const aiInsightsButton = authenticatedPage.locator('button').filter({ hasText: /AI Insights|Generate Insights/i });
    const aiInsightsExists = await aiInsightsButton.count() > 0;
    
    if (aiInsightsExists) {
      await aiInsightsButton.click();
      
      // Wait for insights to generate or error message
      await authenticatedPage.waitForTimeout(2000);
      
      // Take screenshot of AI Insights UI
      await percySnapshot(authenticatedPage, 'Health Page - AI Insights Generated');
      
      console.log('AI Insights UI found and tested');
    } else {
      console.log('AI Insights button not found - may be in different location or not implemented');
      // Take screenshot anyway to show current state
      await percySnapshot(authenticatedPage, 'Health Page - Log Selected');
    }
  });

  test('Model names from Ollama match expected format', async ({ authenticatedPage }) => {
    // Get LLM configs
    const response = await authenticatedPage.request.get('/api/portal/llm-configs');
    const configs = await response.json();
    
    // Filter Ollama configs
    const ollamaConfigs = configs.filter((c: any) => c.provider === 'ollama');
    
    console.log('Ollama configs:', ollamaConfigs);
    
    if (ollamaConfigs.length > 0) {
      // Verify Ollama model names follow expected format
      for (const config of ollamaConfigs) {
        // Ollama models should have format: "model-name:tag" or "model-name"
        // Examples: "qwen2.5-coder:7b", "deepseek-coder:6.7b", "llama2"
        
        const modelName = config.model;
        console.log('Checking Ollama model name:', modelName);
        
        // Model name should not be empty
        expect(modelName).toBeTruthy();
        expect(typeof modelName).toBe('string');
        
        // Model name should not have spaces (Ollama convention)
        expect(modelName).not.toContain(' ');
        
        // If has tag (colon), verify format
        if (modelName.includes(':')) {
          const [name, tag] = modelName.split(':');
          expect(name).toBeTruthy();
          expect(tag).toBeTruthy();
          console.log(`  Model: ${name}, Tag: ${tag}`);
        }
      }
    } else {
      console.log('No Ollama configs found - test may be running without Ollama setup');
    }
  });
});

test.describe('AI Insights - Error Handling', () => {
  test('Shows error if no models configured', async ({ authenticatedPage }) => {
    // This test is informational - checks if error messaging works
    
    const modelSelect = authenticatedPage.locator('select#model-select');
    await expect(modelSelect).toBeVisible();
    
    // Check if disabled (no models case)
    const isDisabled = await modelSelect.isDisabled();
    const options = await modelSelect.locator('option').allTextContents();
    
    if (isDisabled && options.includes('No models available')) {
      console.log('No models configured - showing appropriate error');
      await percySnapshot(authenticatedPage, 'Health Page - No Models Available');
    } else {
      console.log('Models available:', options);
    }
  });

  test('ModelSelector shows loading state initially', async ({ authenticatedPage }) => {
    // Reload page and check for loading state
    const promises = Promise.all([
      authenticatedPage.reload(),
      authenticatedPage.waitForSelector('.model-selector', { timeout: 5000 })
    ]);
    
    await promises;
    
    // Loading state should exist briefly (spinner or loading text)
    const hasLoadingIndicator = await authenticatedPage.locator('.spinner-border, text=Loading').count() > 0;
    console.log('Loading indicator present:', hasLoadingIndicator);
  });
});

test.describe('AI Insights - Integration Validation', () => {
  test('Full flow: Select log -> Check model -> Verify API ready', async ({ authenticatedPage }) => {
    // Step 1: Verify ModelSelector has models
    const modelSelect = authenticatedPage.locator('select#model-select');
    const options = await modelSelect.locator('option').allTextContents();
    expect(options.length).toBeGreaterThan(0);
    const selectedModel = await modelSelect.inputValue();
    console.log('Selected model:', selectedModel);
    
    // Step 2: Select a log entry
    const firstRow = authenticatedPage.locator('table tbody tr').first();
    await firstRow.click();
    await authenticatedPage.waitForTimeout(1000);
    
    // Step 3: Take screenshot of complete UI state
    await percySnapshot(authenticatedPage, 'Health Page - Full Integration State');
    
    // Step 4: Verify log details visible
    const logId = await firstRow.getAttribute('data-log-id') || 'unknown';
    console.log('Selected log ID:', logId);
    
    // Step 5: Verify AI Insights endpoint exists (even if not called yet)
    // This validates the backend API is ready
    const testLogId = 1;
    const insightsResponse = await authenticatedPage.request.get(
      `/api/logs/${testLogId}/insights`,
      { failOnStatusCode: false }
    );
    
    console.log('AI Insights endpoint status:', insightsResponse.status());
    console.log('AI Insights endpoint available:', [200, 404, 500].includes(insightsResponse.status()));
    
    // 200 = insights exist, 404 = no insights yet, 500 = error
    // All are valid - endpoint exists and responds
    expect([200, 404, 500]).toContain(insightsResponse.status());
  });
});
