const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  console.log('Loading http://localhost:3000/...');
  await page.goto('http://localhost:3000/', { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);
  
  // Take screenshot
  await page.screenshot({ path: '/tmp/portal-homepage.png', fullPage: true });
  console.log('Screenshot saved to /tmp/portal-homepage.png');
  
  // Check for login buttons
  const githubButton = await page.locator('button:has-text("GitHub"), a:has-text("GitHub")').count();
  const loginButtons = await page.locator('button:has-text("Login"), button:has-text("Sign in")').count();
  
  console.log(`\nGitHub buttons found: ${githubButton}`);
  console.log(`Login buttons found: ${loginButtons}`);
  
  // Get all button text
  const allButtons = await page.locator('button, a.btn').allTextContents();
  console.log('\nAll buttons on page:');
  allButtons.forEach(text => console.log(`  - "${text}"`));
  
  await browser.close();
})();
