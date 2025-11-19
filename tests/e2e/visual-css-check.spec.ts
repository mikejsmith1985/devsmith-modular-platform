import { test, expect } from '@playwright/test';
import * as fs from 'fs';

test('Visual CSS inspection', async ({ page }) => {
  const cssRequests: any[] = [];
  
  page.on('response', response => {
    if (response.url().includes('.css')) {
      cssRequests.push({
        url: response.url(),
        status: response.status()
      });
    }
  });
  
  await page.goto('/dashboard');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
  
  const headerBg = await page.evaluate(() => {
    const header = document.querySelector('header');
    return header ? window.getComputedStyle(header).backgroundColor : 'NOT FOUND';
  });
  
  const cardBg = await page.evaluate(() => {
    const card = document.querySelector('.card');
    return card ? window.getComputedStyle(card).backgroundColor : 'NOT FOUND';
  });
  
  const bodyBg = await page.evaluate(() => {
    return window.getComputedStyle(document.body).backgroundColor;
  });
  
  const stylesheets = await page.evaluate(() => {
    return Array.from(document.styleSheets).map(sheet => ({
      href: sheet.href,
      disabled: sheet.disabled
    }));
  });
  
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  await page.screenshot({ 
    path: `test-results/visual-css-check-${timestamp}.png`,
    fullPage: true 
  });
  
  const report = {
    timestamp: new Date().toISOString(),
    cssRequests,
    computedStyles: { headerBg, cardBg, bodyBg },
    stylesheets,
    screenshotPath: `test-results/visual-css-check-${timestamp}.png`
  };
  
  fs.writeFileSync('test-results/css-inspection.json', JSON.stringify(report, null, 2));
  
  console.log('\n=== CSS FILES LOADED ===');
  cssRequests.forEach(req => console.log(`${req.status}: ${req.url}`));
  
  console.log('\n=== COMPUTED STYLES ===');
  console.log('Header BG:', headerBg);
  console.log('Card BG:', cardBg);
  console.log('Body BG:', bodyBg);
  
  console.log('\n=== STYLESHEETS ===');
  stylesheets.forEach(s => console.log(s.href));
});
