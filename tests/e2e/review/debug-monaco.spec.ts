import { test, expect } from '../fixtures/auth.fixture';

test('Debug Monaco Editor Selectors', async ({ authenticatedPage: page }) => {
  // Navigate to Review page
  await page.goto('/review', { waitUntil: 'networkidle' });
  
  // Wait for React app to load
  await page.waitForSelector('.analysis-mode-selector', { timeout: 15000 });
  
  // Log all relevant selectors
  console.log('=== Checking Review Page Selectors ===');
  
  // Check for Monaco editor
  const monacoExists = await page.locator('.monaco-editor').count();
  console.log(`Monaco editor (.monaco-editor): ${monacoExists > 0 ? 'FOUND' : 'NOT FOUND'}`);
  
  // Check for code editor container
  const codeEditorExists = await page.locator('.code-editor-container').count();
  console.log(`Code editor container: ${codeEditorExists > 0 ? 'FOUND' : 'NOT FOUND'}`);
  
  // Check for textarea
  const textareaExists = await page.locator('textarea').count();
  console.log(`Textarea count: ${textareaExists}`);
  
  // Get all textarea IDs/classes
  const textareas = await page.locator('textarea').all();
  for (let i = 0; i < textareas.length; i++) {
    const id = await textareas[i].getAttribute('id');
    const className = await textareas[i].getAttribute('class');
    console.log(`  Textarea ${i}: id="${id}" class="${className}"`);
  }
  
  // Check for inputarea (Monaco's hidden textarea)
  const inputareaExists = await page.locator('.inputarea').count();
  console.log(`Monaco inputarea: ${inputareaExists > 0 ? 'FOUND' : 'NOT FOUND'}`);
  
  // Take screenshot for manual inspection
  await page.screenshot({ path: '/tmp/review-page-debug.png', fullPage: true });
  console.log('Screenshot saved to /tmp/review-page-debug.png');
  
  expect(monacoExists).toBeGreaterThan(0);
});
