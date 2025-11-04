const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  console.log('\n=== TESTING REVIEW AUTHENTICATION ===\n');
  
  try {
    // Step 1: Go to dashboard (should be authenticated already based on screenshot)
    console.log('Step 1: Loading dashboard...');
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
    
    // Check if we're authenticated
    const username = await page.$('text=/mikejsmith1985/i');
    if (username) {
      console.log('✅ User is authenticated (mikejsmith1985 visible)');
    } else {
      console.log('❌ Not authenticated - need to login first');
      await browser.close();
      return;
    }
    
    // Step 2: Get the JWT cookie
    const cookies = await context.cookies();
    const jwtCookie = cookies.find(c => c.name === 'devsmith_token');
    
    if (jwtCookie) {
      console.log('✅ JWT cookie found:', jwtCookie.value.substring(0, 50) + '...');
    } else {
      console.log('❌ No JWT cookie found');
      console.log('All cookies:', cookies.map(c => c.name));
    }
    
    // Step 3: Click "Open Review" button
    console.log('\nStep 2: Clicking "Open Review" button...');
    await page.click('text=/Open Review/i');
    
    // Wait for navigation or response
    await page.waitForTimeout(2000);
    
    // Step 4: Check what page we're on
    const currentURL = page.url();
    console.log('Current URL:', currentURL);
    
    // Step 5: Check page content
    const bodyText = await page.evaluate(() => document.body.innerText);
    
    if (bodyText.includes('Authentication required') || bodyText.includes('Unauthorized')) {
      console.log('❌ AUTHENTICATION FAILED - Review rejected the request');
      console.log('Page content:', bodyText.substring(0, 200));
      
      // Check network logs
      page.on('response', response => {
        console.log('Response:', response.status(), response.url());
      });
      
      // Try direct access to review with cookie
      console.log('\nStep 3: Testing direct access to Review service...');
      const reviewResponse = await page.goto('http://localhost:3000/review', { waitUntil: 'networkidle' });
      console.log('Review response status:', reviewResponse.status());
      
      const reviewBody = await page.evaluate(() => document.body.innerText);
      console.log('Review page content:', reviewBody.substring(0, 200));
      
      // Check request headers
      const headers = await page.evaluate(() => {
        return {
          cookie: document.cookie,
          hasAuth: document.cookie.includes('devsmith_token')
        };
      });
      console.log('Client-side headers:', headers);
      
    } else if (bodyText.includes('Code Review') || bodyText.includes('Reading Modes')) {
      console.log('✅ SUCCESS - Review page loaded');
      console.log('Page title:', await page.title());
    } else {
      console.log('⚠️ UNKNOWN STATE');
      console.log('Page content:', bodyText.substring(0, 500));
    }
    
    // Take screenshot
    await page.screenshot({ path: 'review-auth-test.png', fullPage: true });
    console.log('\n📸 Screenshot saved: review-auth-test.png');
    
  } catch (error) {
    console.error('❌ Error:', error.message);
  } finally {
    await browser.close();
  }
})();
