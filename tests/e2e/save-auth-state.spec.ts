import { test } from '@playwright/test';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load env (so BASE_URL is available)
dotenv.config({ path: path.join(__dirname, '../../.env.playwright') });

test('Save authenticated storage state (manual login)', async ({ page, context }) => {
  const baseUrl = process.env.BASE_URL || 'http://localhost:3000';

  console.log('\nOpen headed browser - please complete GitHub login and authorize the app.');
  await page.goto(baseUrl, { waitUntil: 'networkidle' });

  // Click login link (robust selector)
  const loginButton = page.locator('a[href*="/auth/github/login"], a[href="/auth/login"], button:has-text("Login with GitHub")');
  await loginButton.waitFor({ timeout: 60000 });
  await loginButton.click();

  console.log('Browser opened GitHub OAuth flow. Complete login and authorization in the opened window.');

  // Wait for redirect back to our app (user will manually complete GitHub flow)
  await page.waitForURL('http://localhost:3000/**', { timeout: 300000 });

  // Save storage state containing cookies/localStorage for reuse
  const outPath = path.join(__dirname, 'storageState.json');
  await context.storageState({ path: outPath });
  console.log(`✅ Saved storage state to ${outPath}`);
});
