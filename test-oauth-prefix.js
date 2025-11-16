// Quick test to verify no double /api/api/ prefix
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  // Track all requests
  const requests = [];
  page.on('request', req => {
    if (req.url().includes('/api/')) {
      requests.push(req.url());
      console.log('API Request:', req.url());
    }
  });
  
  console.log('Loading Portal...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  
  console.log('\nAll API requests made:');
  requests.forEach(url => console.log('  -', url));
  
  // Check for double prefix
  const hasDoublePrefix = requests.some(url => url.includes('/api/api/'));
  
  if (hasDoublePrefix) {
    console.log('\n❌ FAIL: Found double /api/api/ prefix');
    process.exit(1);
  } else {
    console.log('\n✅ PASS: No double prefix found');
  }
  
  await browser.close();
})();
