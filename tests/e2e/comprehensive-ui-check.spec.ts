import { test, expect } from '@playwright/test';
import * as fs from 'fs';

test('Comprehensive UI validation across all services', async ({ page }) => {
  const results = {
    timestamp: new Date().toISOString(),
    services: {}
  };

  // Test Portal Landing Page
  await page.goto('http://localhost:3000/');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'test-results/portal-landing.png', fullPage: true });
  
  const portalButton = await page.evaluate(() => {
    const btn = document.querySelector('.btn');
    if (!btn) return { found: false };
    const styles = window.getComputedStyle(btn);
    return {
      found: true,
      backgroundColor: styles.backgroundColor,
      color: styles.color,
      padding: styles.padding,
      borderRadius: styles.borderRadius
    };
  });
  
  results.services['portal-landing'] = {
    url: 'http://localhost:3000/',
    screenshot: 'portal-landing.png',
    button: portalButton
  };

  // Test Review Service (unauthenticated)
  await page.goto('http://localhost:3000/review');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'test-results/review-landing.png', fullPage: true });
  
  const reviewElements = await page.evaluate(() => {
    const btn = document.querySelector('.btn');
    const header = document.querySelector('h1, h2, .header');
    return {
      button: btn ? {
        backgroundColor: window.getComputedStyle(btn).backgroundColor,
        exists: true
      } : { exists: false },
      header: header ? {
        color: window.getComputedStyle(header).color,
        exists: true
      } : { exists: false }
    };
  });
  
  results.services['review'] = {
    url: 'http://localhost:3000/review',
    screenshot: 'review-landing.png',
    elements: reviewElements
  };

  // Test Logs Service
  await page.goto('http://localhost:3000/logs');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'test-results/logs-landing.png', fullPage: true });
  
  const logsElements = await page.evaluate(() => {
    const btn = document.querySelector('.btn');
    const card = document.querySelector('.card');
    return {
      button: btn ? { exists: true, bg: window.getComputedStyle(btn).backgroundColor } : { exists: false },
      card: card ? { exists: true, bg: window.getComputedStyle(card).backgroundColor } : { exists: false }
    };
  });
  
  results.services['logs'] = {
    url: 'http://localhost:3000/logs',
    screenshot: 'logs-landing.png',
    elements: logsElements
  };

  // Test Analytics Service
  await page.goto('http://localhost:3000/analytics');
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: 'test-results/analytics-landing.png', fullPage: true });
  
  const analyticsElements = await page.evaluate(() => {
    const btn = document.querySelector('.btn');
    const container = document.querySelector('.container');
    return {
      button: btn ? { exists: true, bg: window.getComputedStyle(btn).backgroundColor } : { exists: false },
      container: container ? { exists: true } : { exists: false }
    };
  });
  
  results.services['analytics'] = {
    url: 'http://localhost:3000/analytics',
    screenshot: 'analytics-landing.png',
    elements: analyticsElements
  };

  // Write comprehensive report
  fs.writeFileSync('test-results/comprehensive-ui-report.json', JSON.stringify(results, null, 2));
  
  console.log('\n=== COMPREHENSIVE UI VALIDATION ===');
  console.log(JSON.stringify(results, null, 2));
  console.log('\n=== SCREENSHOTS SAVED ===');
  console.log('- test-results/portal-landing.png');
  console.log('- test-results/review-landing.png');
  console.log('- test-results/logs-landing.png');
  console.log('- test-results/analytics-landing.png');
});
