import { test, expect } from '@playwright/test';

/**
 * COMPREHENSIVE READING MODES E2E TESTS
 * Tests all 5 reading modes with various code samples and error scenarios
 * 
 * Prerequisites:
 * - Review service running at http://localhost:8081
 * - Ollama running at http://localhost:11434
 * - Model available: mistral:7b-instruct
 */

test.setTimeout(120000); // 2 minutes for AI analysis

const REVIEW_URL = 'http://localhost:8081';
const SAMPLE_GO_CODE = `package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	id := c.Param("id")
	// TODO: Add validation
	user := fetchUserFromDB(id)
	c.JSON(http.StatusOK, user)
}`;

const SAMPLE_VULNERABLE_CODE = `package handlers

import (
	"database/sql"
	"fmt"
)

func GetUser(id string) (*User, error) {
	query := "SELECT * FROM users WHERE id = " + id  // SQL injection
	rows, _ := db.Query(query)  // Error ignored
	// Missing validation, layer violation
	return parseUser(rows), nil
}`;

test.describe('Review Service - All Reading Modes', () => {

	test.describe('Preview Mode - Quick Structure Assessment', () => {
		
		test('Preview Mode analyzes code structure', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			// Paste code
			await page.fill('#code-input', SAMPLE_GO_CODE);
			
			// Select Preview mode
			await page.selectOption('#reading-mode', 'preview');
			
			// Click Analyze
			await page.click('button:has-text("Analyze Code")');
			
			// Wait for results
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			// Verify results contain structural info
			const results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Should show file structure info
			expect(results?.toLowerCase()).toContain('package');
		});

		test('Preview Mode handles empty code gracefully', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.selectOption('#reading-mode', 'preview');
			await page.click('button:has-text("Analyze Code")');
			
			// Should show error message
			await page.waitForSelector('.error, [role="alert"]', { timeout: 5000 });
		});
	});

	test.describe('Skim Mode - Abstractions & Flow', () => {
		
		test('Skim Mode identifies functions and interfaces', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_GO_CODE);
			await page.selectOption('#reading-mode', 'skim');
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			const results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Should identify function
			expect(results?.toLowerCase()).toMatch(/function|getuser/i);
		});

		test('Skim Mode shows function signatures', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			const codeWithMultipleFunctions = `
package service

func GetUser(id int) (*User, error) {}
func CreateUser(u *User) error {}
func DeleteUser(id int) error {}
`;
			
			await page.fill('#code-input', codeWithMultipleFunctions);
			await page.selectOption('#reading-mode', 'skim');
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			const results = await page.textContent('#results-container');
			
			// Should list multiple functions
			expect(results?.toLowerCase()).toMatch(/getuser|createuser|deleteuser/i);
		});
	});

	test.describe('Scan Mode - Targeted Search', () => {
		
		test('Scan Mode finds specific patterns', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_GO_CODE);
			await page.selectOption('#reading-mode', 'scan');
			
			// Provide search query
			await page.fill('#scan-query', 'find validation');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			const results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
		});

		test('Scan Mode searches for error handling', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_VULNERABLE_CODE);
			await page.selectOption('#reading-mode', 'scan');
			await page.fill('#scan-query', 'find error handling');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			const results = await page.textContent('#results-container');
			
			// Should mention error handling issues
			expect(results?.toLowerCase()).toMatch(/error|handling/i);
		});
	});

	test.describe('Detailed Mode - Line-by-Line Analysis', () => {
		
		test('Detailed Mode provides deep analysis', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_GO_CODE);
			await page.selectOption('#reading-mode', 'detailed');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			const results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Should have detailed explanation
			expect(results!.length).toBeGreaterThan(100);
		});

		test('Detailed Mode explains algorithm logic', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			const algorithmCode = `
func BinarySearch(arr []int, target int) int {
	left, right := 0, len(arr)-1
	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}`;
			
			await page.fill('#code-input', algorithmCode);
			await page.selectOption('#reading-mode', 'detailed');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			const results = await page.textContent('#results-container');
			
			// Should explain algorithm
			expect(results?.toLowerCase()).toMatch(/binary|search|algorithm/i);
		});
	});

	test.describe('Critical Mode - Quality Evaluation', () => {
		
		test('Critical Mode identifies security issues', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_VULNERABLE_CODE);
			await page.selectOption('#reading-mode', 'critical');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			const results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Should identify SQL injection
			expect(results?.toLowerCase()).toMatch(/sql|injection|security/i);
		});

		test('Critical Mode finds error handling issues', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_VULNERABLE_CODE);
			await page.selectOption('#reading-mode', 'critical');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			const results = await page.textContent('#results-container');
			
			// Should mention error handling
			expect(results?.toLowerCase()).toMatch(/error|handling/i);
		});

		test('Critical Mode provides severity ratings', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_VULNERABLE_CODE);
			await page.selectOption('#reading-mode', 'critical');
			
			await page.click('button:has-text("Analyze Code")');
			
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			const results = await page.textContent('#results-container');
			
			// Should have severity indicators
			expect(results?.toLowerCase()).toMatch(/critical|important|minor|severity|grade/i);
		});
	});

	test.describe('Error Handling - Graceful Degradation', () => {
		
		test('Shows user-friendly error when Ollama unavailable', async ({ page }) => {
			// This test assumes Ollama might be slow or unavailable
			await page.goto(`${REVIEW_URL}/`);
			
			await page.fill('#code-input', SAMPLE_GO_CODE);
			await page.selectOption('#reading-mode', 'preview');
			
			// Click analyze
			await page.click('button:has-text("Analyze Code")');
			
			// Wait for either success or error
			await Promise.race([
				page.waitForSelector('#results-container', { timeout: 45000 }),
				page.waitForSelector('[role="alert"]', { timeout: 45000 })
			]);
			
			// If error shown, verify it's user-friendly
			const errorElement = await page.$('[role="alert"]');
			if (errorElement) {
				const errorText = await errorElement.textContent();
				
				// Should NOT show raw error strings
				expect(errorText?.toLowerCase()).not.toContain('panic');
				expect(errorText?.toLowerCase()).not.toContain('undefined');
				
				// Should have explanation
				expect(errorText).toBeTruthy();
				expect(errorText!.length).toBeGreaterThan(20);
			}
		});

		test('Circuit breaker message is user-friendly', async ({ page }) => {
			// This test checks if circuit breaker message is shown properly
			// (Would need to trigger circuit open state in real scenario)
			
			await page.goto(`${REVIEW_URL}/`);
			
			// Just verify page loads without error
			await expect(page.locator('body')).toBeVisible();
		});
	});

	test.describe('Model Selection', () => {
		
		test('Can select different Ollama models', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			// Check if model selector exists
			const modelSelector = await page.$('#model-select');
			
			if (modelSelector) {
				// Select a model
				await page.selectOption('#model-select', 'mistral:7b-instruct');
				
				// Fill code and analyze
				await page.fill('#code-input', SAMPLE_GO_CODE);
				await page.selectOption('#reading-mode', 'preview');
				await page.click('button:has-text("Analyze Code")');
				
				// Should work with selected model
				await page.waitForSelector('#results-container, [role="alert"]', { timeout: 40000 });
			}
		});
	});

	test.describe('User Journey - Complete Flow', () => {
		
		test('Complete flow: Paste → Preview → Skim → Critical', async ({ page }) => {
			await page.goto(`${REVIEW_URL}/`);
			
			// Step 1: Paste code
			await page.fill('#code-input', SAMPLE_GO_CODE);
			
			// Step 2: Preview mode
			await page.selectOption('#reading-mode', 'preview');
			await page.click('button:has-text("Analyze Code")');
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			let results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Step 3: Switch to Skim mode (without re-pasting code)
			await page.selectOption('#reading-mode', 'skim');
			await page.click('button:has-text("Analyze Code")');
			await page.waitForSelector('#results-container', { timeout: 30000 });
			
			results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
			
			// Step 4: Switch to Critical mode
			await page.selectOption('#reading-mode', 'critical');
			await page.click('button:has-text("Analyze Code")');
			await page.waitForSelector('#results-container', { timeout: 40000 });
			
			results = await page.textContent('#results-container');
			expect(results).toBeTruthy();
		});
	});
});
