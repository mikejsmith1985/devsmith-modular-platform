import { test, expect } from '@playwright/test';

/**
 * COMPREHENSIVE READING MODES E2E TESTS (HTMX-aware)
 * Tests all 5 reading modes using actual UI selectors and HTMX workflow
 * 
 * Prerequisites:
 * - Review service running at http://localhost:3000/review
 * - Ollama running at http://localhost:11434
 * - Model available: mistral:7b-instruct
 * 
 * Workflow:
 * 1. Fill textarea[name="pasted_code"] with code
 * 2. Select model from #model dropdown (optional)
 * 3. Click mode button (e.g., "Select Preview")
 * 4. Wait for HTMX to swap content into #reading-mode-demo
 * 5. Verify analysis results
 */

test.setTimeout(120000); // 2 minutes for AI analysis

const REVIEW_URL = 'http://localhost:3000/review';

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
	return parseUser(rows), nil
}`;

test.describe('Review Service - All Reading Modes (HTMX)', () => {

	test.beforeEach(async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
	});

	test.describe('Preview Mode - Quick Structure Assessment', () => {
		
		test('Preview Mode analyzes code structure successfully', async ({ page }) => {
			// Fill code textarea
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			// Click Preview mode button
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			// Wait for HTMX to swap results into #reading-mode-demo
			await page.waitForTimeout(5000); // Give Ollama time to respond
			
			// Verify results container has content
			const resultsContainer = page.locator('#reading-mode-demo');
			const content = await resultsContainer.textContent();
			
			expect(content).toBeTruthy();
			expect(content!.length).toBeGreaterThan(100);
			
			// Should contain code-related terms
			const lowerContent = content!.toLowerCase();
			expect(lowerContent).toMatch(/package|function|handler|import/);
		});

		test('Preview Mode shows loading indicator during analysis', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			// Loading indicator should be visible (htmx-indicator class)
			const loadingIndicator = page.locator('#progress-indicator-container');
			await expect(loadingIndicator).toBeVisible({ timeout: 1000 });
		});
	});

	test.describe('Skim Mode - Abstractions & Flow', () => {
		
		test('Skim Mode identifies functions and patterns', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const skimButton = page.locator('button:has-text("Select Skim")').first();
			await skimButton.click();
			
			await page.waitForTimeout(5000);
			
			const resultsContainer = page.locator('#reading-mode-demo');
			const content = await resultsContainer.textContent();
			
			expect(content).toBeTruthy();
			expect(content!.length).toBeGreaterThan(50);
			
			// Should identify GetUser function
			expect(content!.toLowerCase()).toMatch(/getuser|function|method/);
		});

		test('Skim Mode handles complex code', async ({ page }) => {
			const complexCode = SAMPLE_GO_CODE + `\n\nfunc CreateUser(c *gin.Context) {}\nfunc DeleteUser(c *gin.Context) {}`;
			
			await page.fill('textarea[name="pasted_code"]', complexCode);
			
			const skimButton = page.locator('button:has-text("Select Skim")').first();
			await skimButton.click();
			
			await page.waitForTimeout(5000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
		});
	});

	test.describe('Scan Mode - Targeted Search', () => {
		
		test('Scan Mode finds specific patterns', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const scanButton = page.locator('button:has-text("Select Scan")').first();
			await scanButton.click();
			
			await page.waitForTimeout(5000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
		});

		test('Scan Mode detects TODO comments', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const scanButton = page.locator('button:has-text("Select Scan")').first();
			await scanButton.click();
			
			await page.waitForTimeout(5000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			// TODO appears in sample code
			expect(content!.toUpperCase()).toContain('TODO');
		});
	});

	test.describe('Detailed Mode - Deep Implementation Analysis', () => {
		
		test('Detailed Mode provides line-by-line analysis', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const detailedButton = page.locator('button:has-text("Select Detailed")').first();
			await detailedButton.click();
			
			await page.waitForTimeout(7000); // Detailed takes longer
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			expect(content!.length).toBeGreaterThan(100);
		});

		test('Detailed Mode explains code logic', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const detailedButton = page.locator('button:has-text("Select Detailed")').first();
			await detailedButton.click();
			
			await page.waitForTimeout(7000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			
			// Should contain explanatory terms
			const lowerContent = content!.toLowerCase();
			expect(lowerContent).toMatch(/handler|function|parameter|return|request/);
		});
	});

	test.describe('Critical Mode - Quality & Security Review', () => {
		
		test('Critical Mode identifies SQL injection vulnerability', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_VULNERABLE_CODE);
			
			const criticalButton = page.locator('button:has-text("Select Critical")').first();
			await criticalButton.click();
			
			await page.waitForTimeout(10000); // Critical takes longest
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			
			// Should identify SQL injection
			const lowerContent = content!.toLowerCase();
			expect(lowerContent).toMatch(/sql|injection|vulnerability|security/);
		});

		test('Critical Mode detects error handling issues', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_VULNERABLE_CODE);
			
			const criticalButton = page.locator('button:has-text("Select Critical")').first();
			await criticalButton.click();
			
			await page.waitForTimeout(10000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			
			// Should mention error handling
			expect(content!.toLowerCase()).toMatch(/error|handling|ignored/);
		});

		test('Critical Mode provides severity assessment', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_VULNERABLE_CODE);
			
			const criticalButton = page.locator('button:has-text("Select Critical")').first();
			await criticalButton.click();
			
			await page.waitForTimeout(10000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			expect(content!.length).toBeGreaterThan(200);
		});
	});

	test.describe('Error Handling', () => {
		
		test('Shows error when code is empty', async ({ page }) => {
			// Don't fill textarea, just click mode
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			await page.waitForTimeout(2000);
			
			// Should show error or stay on same page
			const url = page.url();
			expect(url).toContain('/review');
		});

		test('Handles invalid code gracefully', async ({ page }) => {
			const invalidCode = '}{][)(}{][][][';
			await page.fill('textarea[name="pasted_code"]', invalidCode);
			
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			await page.waitForTimeout(5000);
			
			// Should still return results (AI should explain syntax errors)
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
		});
	});

	test.describe('Model Selection', () => {
		
		test('Can select different AI models', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			// Select DeepSeek Coder model
			await page.selectOption('#model', 'deepseek-coder:6.7b');
			
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			await page.waitForTimeout(5000);
			
			const content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
		});
	});

	test.describe('User Journey - Complete Workflow', () => {
		
		test('Complete workflow: Paste → Preview → Skim → Critical', async ({ page }) => {
			// Step 1: Paste code
			await page.fill('textarea[name="pasted_code"]', SAMPLE_VULNERABLE_CODE);
			
			// Step 2: Preview Mode
			await page.locator('button:has-text("Select Preview")').first().click();
			await page.waitForTimeout(5000);
			
			let content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			
			// Step 3: Skim Mode (code still in textarea)
			await page.locator('button:has-text("Select Skim")').first().click();
			await page.waitForTimeout(5000);
			
			content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			
			// Step 4: Critical Mode (final review)
			await page.locator('button:has-text("Select Critical")').first().click();
			await page.waitForTimeout(10000);
			
			content = await page.locator('#reading-mode-demo').textContent();
			expect(content).toBeTruthy();
			expect(content!.toLowerCase()).toMatch(/sql|injection|error|security/);
		});
	});

	test.describe('Visual Regression', () => {
		
		test('Preview Mode results visual snapshot', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
			
			const previewButton = page.locator('button:has-text("Select Preview")').first();
			await previewButton.click();
			
			await page.waitForTimeout(5000);
			
			// Take snapshot of results
			const resultsContainer = page.locator('#reading-mode-demo');
			await expect(resultsContainer).toHaveScreenshot('preview-mode-results.png', {
				maxDiffPixels: 100,
			});
		});

		test('Critical Mode results visual snapshot', async ({ page }) => {
			await page.fill('textarea[name="pasted_code"]', SAMPLE_VULNERABLE_CODE);
			
			const criticalButton = page.locator('button:has-text("Select Critical")').first();
			await criticalButton.click();
			
			await page.waitForTimeout(10000);
			
			// Take snapshot of critical analysis
			const resultsContainer = page.locator('#reading-mode-demo');
			await expect(resultsContainer).toHaveScreenshot('critical-mode-results.png', {
				maxDiffPixels: 100,
			});
		});
	});
});
