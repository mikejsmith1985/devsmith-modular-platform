const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  // Capture console messages
  const consoleMessages = [];
  page.on('console', msg => {
    consoleMessages.push(`[${msg.type()}] ${msg.text()}`);
  });
  
  // Capture errors
  const errors = [];
  page.on('pageerror', error => {
    errors.push(error.toString());
  });
  
  console.log('Loading http://localhost:3000/...');
  await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
  
  // Wait a bit for React to render
  await page.waitForTimeout(2000);
  
  // Check root div content
  const rootContent = await page.locator('#root').innerHTML();
  
  console.log('\n=== Root div content ===');
  console.log(rootContent.substring(0, 500));
  
  console.log('\n=== Console messages ===');
  consoleMessages.forEach(msg => console.log(msg));
  
  console.log('\n=== Errors ===');
  if (errors.length === 0) {
    console.log('No errors!');
  } else {
    errors.forEach(err => console.log(err));
  }
  
  await browser.close();
})();
