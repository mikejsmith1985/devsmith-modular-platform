import { test } from '@playwright/test';
import * as fs from 'fs';

test('Visual landing page CSS inspection', async ({ page }) => {
  const cssRequests: any[] = [];
  
  page.on('response', response => {
    if (response.url().includes('.css')) {
      cssRequests.push({
        url: response.url(),
        status: response.status()
      });
    }
  });
  
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
  
  const bodyBg = await page.evaluate(() => window.getComputedStyle(document.body).backgroundColor);
  const buttonColor = await page.evaluate(() => {
    const btn = document.querySelector('a[href*="auth"]') || document.querySelector('button');
    return btn ? window.getComputedStyle(btn).backgroundColor : 'NOT FOUND';
  });
  
  const stylesheets = await page.evaluate(() => {
    return Array.from(document.styleSheets).map(sheet => ({
      href: sheet.href,
      disabled: sheet.disabled
    }));
  });
  
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  await page.screenshot({ path: `test-results/landing-${timestamp}.png`, fullPage: true });
  
  const report = {
    timestamp: new Date().toISOString(),
    cssRequests,
    computedStyles: { bodyBg, buttonColor },
    stylesheets,
    screenshotPath: `test-results/landing-${timestamp}.png`
  };
  
  fs.writeFileSync('test-results/landing-inspection.json', JSON.stringify(report, null, 2));
  
  console.log('\n=== CSS FILES LOADED ON LANDING PAGE ===');
  cssRequests.forEach(req => console.log(`${req.status}: ${req.url}`));
  
  console.log('\n=== COMPUTED STYLES ===');
  console.log('Body BG:', bodyBg);
  console.log('Button Color:', buttonColor);
  
  console.log('\n=== STYLESHEETS ===');
  stylesheets.forEach(s => console.log(s.href || 'inline'));
});
