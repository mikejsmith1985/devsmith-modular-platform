const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    
    // Get page title
    const title = await page.title();
    console.log('\nPage Title:', title);
    
    // Get body text
    const bodyText = await page.evaluate(() => document.body.innerText);
    console.log('\nBody Text (first 500 chars):');
    console.log(bodyText.substring(0, 500));
    
    // Check for specific elements
    const hasWelcome = await page.$('text=/Welcome/i');
    const hasCodeReview = await page.$('text=/Code Review/i');
    const hasReady = await page.$('text=/Ready/i');
    const hasComingSoon = await page.$('text=/Coming Soon/i');
    
    console.log('\n=== ELEMENT CHECKS ===');
    console.log('Has "Welcome":', !!hasWelcome);
    console.log('Has "Code Review":', !!hasCodeReview);
    console.log('Has "Ready" badge:', !!hasReady);
    console.log('Has "Coming Soon" badge:', !!hasComingSoon);
    
    // Get all visible text containing "Ready" or "Coming Soon"
    const badges = await page.evaluate(() => {
      const allElements = document.querySelectorAll('*');
      const results = [];
      allElements.forEach(el => {
        const text = el.textContent;
        if ((text.includes('Ready') || text.includes('Coming Soon')) && el.children.length === 0) {
          results.push({
            text: text.trim(),
            tag: el.tagName,
            classes: el.className
          });
        }
      });
      return results;
    });
    
    console.log('\n=== ALL BADGE TEXTS ===');
    console.log(JSON.stringify(badges, null, 2));
    
  } catch (error) {
    console.error('Error:', error.message);
  } finally {
    await browser.close();
  }
})();
