import { test } from '@playwright/test';
import * as fs from 'fs';

test('Detailed style check for all services', async ({ page }) => {
  const results = {};
  
  const services = [
    { name: 'portal', url: 'http://localhost:3000/' },
    { name: 'review', url: 'http://localhost:3000/review' },
    { name: 'logs', url: 'http://localhost:3000/logs' },
    { name: 'analytics', url: 'http://localhost:3000/analytics' }
  ];
  
  for (const service of services) {
    await page.goto(service.url);
    await page.waitForLoadState('networkidle');
    
    // Capture all computed styles
    const styles = await page.evaluate(() => {
      const elements = {
        buttons: Array.from(document.querySelectorAll('.btn, button, a[class*="btn"]')).map(el => {
          const styles = window.getComputedStyle(el);
          return {
            class: el.className,
            text: el.textContent?.trim().substring(0, 50),
            backgroundColor: styles.backgroundColor,
            color: styles.color,
            padding: styles.padding,
            border: styles.border,
            borderRadius: styles.borderRadius
          };
        }),
        headers: Array.from(document.querySelectorAll('h1, h2, h3')).map(el => {
          const styles = window.getComputedStyle(el);
          return {
            tag: el.tagName,
            text: el.textContent?.trim().substring(0, 50),
            color: styles.color,
            fontSize: styles.fontSize,
            fontWeight: styles.fontWeight
          };
        }),
        cards: Array.from(document.querySelectorAll('.card, [class*="card"]')).map(el => {
          const styles = window.getComputedStyle(el);
          return {
            class: el.className,
            backgroundColor: styles.backgroundColor,
            border: styles.border,
            borderRadius: styles.borderRadius,
            boxShadow: styles.boxShadow
          };
        }),
        body: (() => {
          const styles = window.getComputedStyle(document.body);
          return {
            backgroundColor: styles.backgroundColor,
            color: styles.color,
            fontFamily: styles.fontFamily
          };
        })()
      };
      return elements;
    });
    
    results[service.name] = {
      url: service.url,
      styles: styles
    };
    
    console.log(`\n=== ${service.name.toUpperCase()} SERVICE ===`);
    console.log(`Buttons found: ${styles.buttons.length}`);
    console.log(`Headers found: ${styles.headers.length}`);
    console.log(`Cards found: ${styles.cards.length}`);
    console.log(`Body BG: ${styles.body.backgroundColor}`);
    console.log(`Body Color: ${styles.body.color}`);
  }
  
  fs.writeFileSync('test-results/detailed-style-report.json', JSON.stringify(results, null, 2));
  console.log('\n=== Report saved to test-results/detailed-style-report.json ===');
});
