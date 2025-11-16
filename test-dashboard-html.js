const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    
    // Get the full HTML of cards section
    const cardsHTML = await page.evaluate(() => {
      const cardsContainer = document.querySelector('[class*="grid"]');
      return cardsContainer ? cardsContainer.innerHTML : 'No grid found';
    });
    
    console.log('\n=== CARDS HTML ===');
    console.log(cardsHTML);
    
    // Get badge elements specifically
    const badges = await page.$$eval('[class*="badge"]', elements => 
      elements.map(el => ({
        text: el.textContent.trim(),
        classes: el.className
      }))
    );
    
    console.log('\n=== BADGE STATES ===');
    console.log(JSON.stringify(badges, null, 2));
    
  } catch (error) {
    console.error('Error:', error.message);
  } finally {
    await browser.close();
  }
})();
