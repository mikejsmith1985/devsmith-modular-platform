const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    // Go to dashboard
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    
    // Take screenshot
    await page.screenshot({ path: 'dashboard-actual.png', fullPage: true });
    
    // Get card states
    const cards = await page.$$('.card, [class*="card"]');
    console.log('\n=== DASHBOARD CARD STATES ===');
    
    for (let i = 0; i < cards.length; i++) {
      const title = await cards[i].$eval('[class*="title"], h2, h3', el => el.textContent).catch(() => 'Unknown');
      const badge = await cards[i].$eval('[class*="badge"]', el => el.textContent).catch(() => 'No badge');
      const badgeColor = await cards[i].$eval('[class*="badge"]', el => {
        const classes = el.className;
        if (classes.includes('green')) return 'green';
        if (classes.includes('gray')) return 'gray';
        return 'unknown';
      }).catch(() => 'No badge');
      
      console.log(`Card ${i + 1}: "${title}" - Badge: "${badge}" (${badgeColor})`);
    }
    
    console.log('\n=== SCREENSHOT SAVED: dashboard-actual.png ===\n');
  } catch (error) {
    console.error('Error:', error.message);
  } finally {
    await browser.close();
  }
})();
