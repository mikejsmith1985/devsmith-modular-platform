/**
 * Comprehensive Review App Testing - All Mode Combinations
 * 
 * Tests ALL combinations of:
 * - User Modes: beginner, novice, intermediate, expert (4)
 * - Output Modes: quick, full (2)
 * - Reading Modes: Preview, Skim, Scan, Detailed, Critical (5)
 * 
 * Total: 40 test combinations (4 * 2 * 5)
 * 
 * This test validates the JSON fix from Phase 27 and ensures
 * Review service returns JSON (not HTML) for all mode combinations.
 */

import { test, expect } from './fixtures/auth.fixture';
import type { Page } from '@playwright/test';

const TEST_CODE = `package main

import (
	"fmt"
	"net/http"
)

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Potential SQL injection vulnerability
	userID := r.URL.Query().Get("id")
	query := "SELECT * FROM users WHERE id = " + userID
	
	// Missing error handling
	db.Query(query)
	
	fmt.Fprintf(w, "User found")
}
`;

const USER_MODES = ['beginner', 'novice', 'intermediate', 'expert'];
const OUTPUT_MODES = ['quick', 'full'];
const READING_MODES = [
  { name: 'Preview', endpoint: '/api/review/modes/preview' },
  { name: 'Skim', endpoint: '/api/review/modes/skim' },
  { name: 'Scan', endpoint: '/api/review/modes/scan' },
  { name: 'Detailed', endpoint: '/api/review/modes/detailed' },
  { name: 'Critical', endpoint: '/api/review/modes/critical' }
];

test.describe('Review App - Comprehensive Mode Testing', () => {
  
  // Test all combinations systematically
  for (const userMode of USER_MODES) {
    for (const outputMode of OUTPUT_MODES) {
      for (const readingMode of READING_MODES) {
        
        test(`${readingMode.name} mode with ${userMode}/${outputMode}`, async ({ authenticatedPage }) => {
          const page = authenticatedPage;
          
          // Navigate to Review app
          await page.goto('/review', { 
            waitUntil: 'networkidle',
            timeout: 60000 
          });
          
          // Wait for React app to load
          await page.waitForSelector('[data-testid="code-input"], textarea', { timeout: 30000 });
          
          // Fill in code
          const codeInput = page.locator('[data-testid="code-input"], textarea').first();
          await codeInput.fill(TEST_CODE);
          
          // Select user mode
          const userModeSelect = page.locator('select[name="user_mode"], #user-mode-select').first();
          await userModeSelect.selectOption(userMode);
          
          // Select output mode
          const outputModeSelect = page.locator('select[name="output_mode"], #output-mode-select').first();
          await outputModeSelect.selectOption(outputMode);
          
          // Select reading mode
          const readingModeSelect = page.locator('select[name="reading_mode"], #reading-mode-select').first();
          await readingModeSelect.selectOption(readingMode.name.toLowerCase());
          
          // Intercept API request
          const responsePromise = page.waitForResponse(
            response => response.url().includes(readingMode.endpoint) && response.status() === 200,
            { timeout: 90000 }
          );
          
          // Click analyze button
          await page.click('button:has-text("Analyze"), [data-testid="analyze-button"]');
          
          // Wait for response
          const response = await responsePromise;
          
          // Verify response is JSON (not HTML)
          const contentType = response.headers()['content-type'];
          expect(contentType).toContain('application/json');
          expect(contentType).not.toContain('text/html');
          
          // Parse response
          const data = await response.json();
          
          // Verify response structure
          expect(data).toBeTruthy();
          expect(data).not.toHaveProperty('error');
          
          // Mode-specific validations
          switch (readingMode.name) {
            case 'Preview':
              expect(data).toHaveProperty('quick_preview');
              expect(data).toHaveProperty('summary');
              break;
            case 'Skim':
              expect(data).toHaveProperty('functions');
              expect(data).toHaveProperty('data_models');
              break;
            case 'Scan':
              expect(data).toHaveProperty('matches');
              break;
            case 'Detailed':
              expect(data).toHaveProperty('line_explanations');
              break;
            case 'Critical':
              expect(data).toHaveProperty('issues');
              break;
          }
          
          // Take screenshot for visual verification
          await page.screenshot({ 
            path: `test-results/review-${readingMode.name.toLowerCase()}-${userMode}-${outputMode}.png`,
            fullPage: true 
          });
          
        });
      }
    }
  }
  
  // Additional error handling tests
  test('Should handle empty code gracefully', async ({ authenticatedPage }) => {
    const page = authenticatedPage;
    
    await page.goto('/review', { waitUntil: 'networkidle' });
    await page.waitForSelector('[data-testid="code-input"], textarea');
    
    // Try to analyze without code
    await page.click('button:has-text("Analyze"), [data-testid="analyze-button"]');
    
    // Should show validation error
    await expect(page.locator('text=Code is required, text=Please enter code')).toBeVisible({ timeout: 5000 });
  });
  
  test('Should handle Ollama unavailable gracefully', async ({ authenticatedPage }) => {
    const page = authenticatedPage;
    
    await page.goto('/review', { waitUntil: 'networkidle' });
    await page.waitForSelector('[data-testid="code-input"], textarea');
    
    // Fill code
    const codeInput = page.locator('[data-testid="code-input"], textarea').first();
    await codeInput.fill(TEST_CODE);
    
    // If Ollama is down, should show friendly error
    const responsePromise = page.waitForResponse(
      response => response.url().includes('/api/review/modes/'),
      { timeout: 90000 }
    );
    
    await page.click('button:has-text("Analyze"), [data-testid="analyze-button"]');
    
    const response = await responsePromise;
    
    if (response.status() !== 200) {
      // Should show error message, not crash
      await expect(page.locator('text=unavailable, text=failed')).toBeVisible({ timeout: 5000 });
    }
  });
  
});
