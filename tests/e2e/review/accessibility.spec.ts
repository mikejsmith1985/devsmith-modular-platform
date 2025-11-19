import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

/**
 * ACCESSIBILITY TESTING WITH AXE-CORE
 * 
 * Tests key flows for WCAG 2.1 Level AA compliance:
 * - Session creation form
 * - Preview Mode results
 * - Detailed Mode results  
 * - Critical Mode results
 * 
 * Fails build on critical violations (serious, critical)
 * Logs moderate/minor violations for tracking
 */

test.setTimeout(120000); // 2 minutes for AI analysis

const REVIEW_URL = '/review';

const SAMPLE_GO_CODE = `package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	id := c.Param("id")
	user := fetchUserFromDB(id)
	c.JSON(http.StatusOK, user)
}`;

test.describe('Accessibility Testing (axe-core)', () => {

	test('Session creation form is accessible', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Run axe accessibility scan on initial page
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Log all violations for tracking
		if (accessibilityScanResults.violations.length > 0) {
			console.log('Accessibility violations found:');
			accessibilityScanResults.violations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description}`);
				console.log(`    Impact: ${violation.impact}`);
				console.log(`    Help: ${violation.help}`);
			});
		}
		
		// Fail test only on serious or critical violations
		const criticalViolations = accessibilityScanResults.violations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalViolations.length, 
			`Found ${criticalViolations.length} critical accessibility violations`
		).toBe(0);
	});

	test('Preview Mode results are accessible', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Fill and analyze code
		await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
		await page.locator('button:has-text("Select Preview")').first().click();
		await page.waitForTimeout(5000); // Wait for AI analysis
		
		// Run axe scan on results page
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Log violations
		if (accessibilityScanResults.violations.length > 0) {
			console.log('Preview Mode accessibility violations:');
			accessibilityScanResults.violations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical violations
		const criticalViolations = accessibilityScanResults.violations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalViolations.length).toBe(0);
	});

	test('Detailed Mode results are accessible', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Fill and analyze code
		await page.fill('textarea[name="pasted_code"]', SAMPLE_GO_CODE);
		await page.locator('button:has-text("Select Detailed")').first().click();
		await page.waitForTimeout(7000); // Detailed takes longer
		
		// Run axe scan
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Log violations
		if (accessibilityScanResults.violations.length > 0) {
			console.log('Detailed Mode accessibility violations:');
			accessibilityScanResults.violations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical violations
		const criticalViolations = accessibilityScanResults.violations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalViolations.length).toBe(0);
	});

	test('Critical Mode results are accessible', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		const vulnerableCode = `package handlers
import "database/sql"

func GetUser(id string) (*User, error) {
	query := "SELECT * FROM users WHERE id = " + id // SQL injection
	rows, _ := db.Query(query) // Error ignored
	return parseUser(rows), nil
}`;
		
		// Fill and analyze code
		await page.fill('textarea[name="pasted_code"]', vulnerableCode);
		await page.locator('button:has-text("Select Critical")').first().click();
		await page.waitForTimeout(10000); // Critical takes longest
		
		// Run axe scan
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Log violations
		if (accessibilityScanResults.violations.length > 0) {
			console.log('Critical Mode accessibility violations:');
			accessibilityScanResults.violations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical violations
		const criticalViolations = accessibilityScanResults.violations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalViolations.length).toBe(0);
	});

	test('Dark mode toggle maintains accessibility', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Toggle dark mode
		const darkModeButton = page.locator('#dark-mode-toggle');
		await darkModeButton.click();
		await page.waitForTimeout(500);
		
		// Run axe scan in dark mode
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Log violations
		if (accessibilityScanResults.violations.length > 0) {
			console.log('Dark mode accessibility violations:');
			accessibilityScanResults.violations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical violations
		const criticalViolations = accessibilityScanResults.violations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalViolations.length).toBe(0);
	});

	test('Navigation links are keyboard accessible', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Check that navigation elements are keyboard accessible
		const accessibilityScanResults = await new AxeBuilder({ page })
			.include('nav')
			.include('a[href]')
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Focus on navigation-specific rules
		const navViolations = accessibilityScanResults.violations.filter(
			v => v.nodes.some(n => n.target.some(t => t.includes('nav') || t.includes('a')))
		);
		
		if (navViolations.length > 0) {
			console.log('Navigation accessibility violations:');
			navViolations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical navigation violations
		const criticalNavViolations = navViolations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalNavViolations.length).toBe(0);
	});

	test('Form inputs have proper labels and ARIA attributes', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Run axe scan focused on forms
		const accessibilityScanResults = await new AxeBuilder({ page })
			.include('form')
			.include('textarea')
			.include('input')
			.include('select')
			.include('button')
			.withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
			.analyze();
		
		// Check form-specific violations
		const formViolations = accessibilityScanResults.violations.filter(
			v => ['label', 'aria', 'input', 'button'].some(keyword => 
				v.id.toLowerCase().includes(keyword)
			)
		);
		
		if (formViolations.length > 0) {
			console.log('Form accessibility violations:');
			formViolations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
			});
		}
		
		// Fail on critical form violations
		const criticalFormViolations = formViolations.filter(
			v => v.impact === 'critical' || v.impact === 'serious'
		);
		
		expect(criticalFormViolations.length).toBe(0);
	});

	test('Color contrast meets WCAG AA standards', async ({ page }) => {
		await page.goto(REVIEW_URL, { waitUntil: 'domcontentloaded' });
		
		// Run axe scan focused on color contrast
		const accessibilityScanResults = await new AxeBuilder({ page })
			.withTags(['wcag2aa'])
			.analyze();
		
		// Filter for color contrast violations
		const contrastViolations = accessibilityScanResults.violations.filter(
			v => v.id.includes('color-contrast')
		);
		
		if (contrastViolations.length > 0) {
			console.log('Color contrast violations:');
			contrastViolations.forEach(violation => {
				console.log(`  - ${violation.id}: ${violation.description} (${violation.impact})`);
				violation.nodes.forEach(node => {
					console.log(`    Element: ${node.html}`);
					console.log(`    Contrast ratio: ${node.any[0]?.data?.contrastRatio || 'N/A'}`);
				});
			});
		}
		
		// Fail on serious contrast violations
		const criticalContrastViolations = contrastViolations.filter(
			v => v.impact === 'serious' || v.impact === 'critical'
		);
		
		expect(criticalContrastViolations.length).toBe(0);
	});
});
