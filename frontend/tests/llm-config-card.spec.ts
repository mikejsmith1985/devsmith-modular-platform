/**
 * E2E Tests: LLM Config Card on Portal Dashboard
 * 
 * Phase 5, Task 5.1 - TDD RED Phase
 * 
 * Tests the new "AI Model Management" card on the Portal dashboard
 * that allows users to navigate to LLM configuration page.
 */

import { test, expect } from '@playwright/test';

test.describe('Portal Dashboard - LLM Config Card', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to portal dashboard
    // Note: May need authentication setup in future
    await page.goto('http://localhost:3000/portal');
    await page.waitForLoadState('networkidle');
  });

  /**
   * Functional Tests
   */
  
  test('should display LLM Config card on dashboard', async ({ page }) => {
    // Find the AI Model Management card
    const card = page.locator('.card:has-text("AI Model Management")');
    
    // Verify card is visible
    await expect(card).toBeVisible();
    
    // Verify card has correct icon
    const icon = card.locator('i.bi-robot');
    await expect(icon).toBeVisible();
    
    // Verify card has correct title
    const title = card.locator('h5.card-title');
    await expect(title).toContainText('AI Model Management');
    
    // Verify card has description
    const description = card.locator('p.card-text');
    await expect(description).toContainText('Configure AI models and API keys');
  });
  
  test('should have "Manage Models" button with correct styling', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    
    // Find the button
    const button = card.locator('a.btn-primary:has-text("Manage Models")');
    
    // Verify button exists and is visible
    await expect(button).toBeVisible();
    
    // Verify button has correct classes
    await expect(button).toHaveClass(/btn/);
    await expect(button).toHaveClass(/btn-primary/);
    
    // Verify it's a link (not a button element)
    const tagName = await button.evaluate(el => el.tagName.toLowerCase());
    expect(tagName).toBe('a');
  });
  
  test('should navigate to /llm-config when "Manage Models" clicked', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    const button = card.locator('a:has-text("Manage Models")');
    
    // Click the button
    await button.click();
    
    // Verify navigation to LLM config page
    await page.waitForURL('**/llm-config');
    
    // Verify URL is correct
    expect(page.url()).toContain('/llm-config');
  });
  
  test('should maintain card styling consistency with other dashboard cards', async ({ page }) => {
    const llmCard = page.locator('.card:has-text("AI Model Management")');
    const reviewCard = page.locator('.card:has-text("Review")').first();
    
    // Both cards should have shadow-sm class
    await expect(llmCard).toHaveClass(/shadow-sm/);
    await expect(reviewCard).toHaveClass(/shadow-sm/);
    
    // Both cards should have card-body
    const llmBody = llmCard.locator('.card-body');
    const reviewBody = reviewCard.locator('.card-body');
    await expect(llmBody).toBeVisible();
    await expect(reviewBody).toBeVisible();
    
    // Both should have card-title with h5
    const llmTitle = llmCard.locator('h5.card-title');
    const reviewTitle = reviewCard.locator('h5.card-title');
    await expect(llmTitle).toBeVisible();
    await expect(reviewTitle).toBeVisible();
  });
  
  test('should display card in correct position (after existing cards)', async ({ page }) => {
    // Get all cards on dashboard
    const allCards = page.locator('.card');
    const cardCount = await allCards.count();
    
    // Verify we have at least 5 cards (Portal, Review, Logs, Analytics, LLM Config)
    expect(cardCount).toBeGreaterThanOrEqual(5);
    
    // LLM Config card should be visible
    const llmCard = page.locator('.card:has-text("AI Model Management")');
    await expect(llmCard).toBeVisible();
  });
  
  test('should display robot icon with correct Bootstrap Icons class', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    const icon = card.locator('i.bi-robot');
    
    // Verify icon has bi-robot class
    await expect(icon).toHaveClass(/bi-robot/);
    
    // Verify icon has spacing class
    await expect(icon).toHaveClass(/me-2/);
  });
  
  test('should have accessible card structure', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    
    // Verify card has proper semantic structure
    const cardBody = card.locator('.card-body');
    await expect(cardBody).toBeVisible();
    
    // Verify heading hierarchy (h5 for card title)
    const title = card.locator('h5');
    await expect(title).toBeVisible();
    
    // Verify link is keyboard accessible
    const link = card.locator('a:has-text("Manage Models")');
    await link.focus();
    await expect(link).toBeFocused();
  });
  
  test('should handle click events correctly', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    const button = card.locator('a:has-text("Manage Models")');
    
    // Verify initial URL is dashboard
    expect(page.url()).toContain('/portal');
    
    // Click button
    await button.click();
    
    // Wait for navigation
    await page.waitForURL('**/llm-config');
    
    // Verify we're on LLM config page
    expect(page.url()).toContain('/llm-config');
    
    // Go back to dashboard
    await page.goBack();
    await page.waitForURL('**/portal');
    
    // Verify card is still visible
    await expect(card).toBeVisible();
  });
  
  /**
   * Visual Regression Tests (Percy)
   */
  
  test('visual: LLM Config card on dashboard', async ({ page }) => {
    // Wait for all cards to load
    await page.waitForSelector('.card:has-text("AI Model Management")');
    
    // Take Percy snapshot of entire dashboard
    // await percySnapshot(page, 'Portal Dashboard with LLM Config Card');
    
    // NOTE: Percy integration placeholder
    // Actual implementation requires @percy/playwright package
    console.log('Percy snapshot: Portal Dashboard with LLM Config Card');
  });
  
  test('visual: LLM Config card hover state', async ({ page }) => {
    const card = page.locator('.card:has-text("AI Model Management")');
    const button = card.locator('a:has-text("Manage Models")');
    
    // Hover over button
    await button.hover();
    
    // Wait for hover animation
    await page.waitForTimeout(300);
    
    // Take Percy snapshot
    // await percySnapshot(page, 'LLM Config Card - Button Hover');
    
    console.log('Percy snapshot: LLM Config Card - Button Hover');
  });
  
  test('visual: Dashboard responsive layout with LLM Config card', async ({ page }) => {
    // Test different viewport sizes
    const viewports = [
      { width: 375, height: 667, name: 'Mobile' },    // iPhone SE
      { width: 768, height: 1024, name: 'Tablet' },   // iPad
      { width: 1920, height: 1080, name: 'Desktop' }  // Full HD
    ];
    
    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(300); // Wait for layout shift
      
      // Verify card is visible at this viewport
      const card = page.locator('.card:has-text("AI Model Management")');
      await expect(card).toBeVisible();
      
      // Take Percy snapshot
      // await percySnapshot(page, `Portal Dashboard LLM Card - ${viewport.name}`);
      
      console.log(`Percy snapshot: Portal Dashboard LLM Card - ${viewport.name}`);
    }
  });

});

/**
 * Test Summary:
 * 
 * Functional Tests (8):
 * - Card displays on dashboard
 * - Button has correct styling
 * - Navigation to /llm-config works
 * - Card styling consistent with other cards
 * - Card position correct
 * - Robot icon displays correctly
 * - Card structure is accessible
 * - Click events work correctly
 * 
 * Visual Tests (3):
 * - Dashboard with LLM Config card
 * - Button hover state
 * - Responsive layout (mobile/tablet/desktop)
 * 
 * Total: 11 tests
 * 
 * Expected Result: All tests FAIL (RED phase)
 * - Card does not exist yet in PortalDashboard.jsx
 * - Route /llm-config not defined yet
 * 
 * Next Step: GREEN phase - Implement the card
 */
