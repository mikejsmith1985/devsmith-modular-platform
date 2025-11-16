const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  console.log('Loading http://localhost:3000/login...');
  await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);
  
  // Take screenshot
  await page.screenshot({ path: '/tmp/portal-login-page.png', fullPage: true });
  console.log('Screenshot saved to /tmp/portal-login-page.png');
  
  // Check for login buttons
  const githubButton = await page.locator('button:has-text("GitHub"), a:has-text("GitHub")').count();
  const loginButtons = await page.locator('button:has-text("Login")').count();
  
  console.log(`\nGitHub buttons found: ${githubButton}`);
  console.log(`Login buttons found: ${loginButtons}`);
  
  // Get all button text
  const allButtons = await page.locator('button, a.btn').allTextContents();
  console.log('\nAll buttons on page:');
  allButtons.forEach(text => console.log(`  - "${text}"`));
  
  // Try to find and click GitHub login
  const githubLoginButton = page.locator('button:has-text("Login with GitHub"), a:has-text("Login with GitHub")').first();
  const isVisible = await githubLoginButton.isVisible().catch(() => false);
  console.log(`\n"Login with GitHub" button visible: ${isVisible}`);
  
  if (isVisible) {
    console.log('Clicking GitHub login button...');
    await githubLoginButton.click();
    await page.waitForTimeout(1000);
    console.log(`Current URL: ${page.url()}`);
  }
  
  await browser.close();
})();
