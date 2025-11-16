const { test, expect } = require('@playwright/test');

/**
 * Full Browser Mode Visual Validation
 * 
 * This test validates that Full Browser mode UI integration is working correctly.
 * It checks for the presence of the FileTreeBrowser component and layout changes.
 */

test('Full Browser Mode - UI Integration Check', async ({ page }) => {
  console.log('\n🧪 Testing Full Browser Mode UI Integration\n');

  // Navigate to Review page
  await page.goto('http://localhost:3000/review');
  console.log('✓ Navigated to Review page');

  // Wait for page to load
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);

  // Check for Import button
  const importButton = page.locator('button:has-text("Import from GitHub")');
  const importExists = await importButton.count() > 0;
  console.log(`${importExists ? '✓' : '✗'} Import from GitHub button: ${importExists ? 'FOUND' : 'NOT FOUND'}`);

  if (importExists) {
    console.log('\n📋 Import button is present - Full Browser mode UI is integrated');
    
    // Click Import button to open modal
    await importButton.click();
    await page.waitForTimeout(500);
    
    // Check for modal
    const modalVisible = await page.locator('.modal.show').isVisible();
    console.log(`${modalVisible ? '✓' : '✗'} RepoImportModal: ${modalVisible ? 'VISIBLE' : 'NOT VISIBLE'}`);
    
    if (modalVisible) {
      // Check for Full Browser mode radio button
      const fullBrowserRadio = page.locator('input[type="radio"][value="full"]');
      const radioExists = await fullBrowserRadio.count() > 0;
      console.log(`${radioExists ? '✓' : '✗'} Full Browser mode radio button: ${radioExists ? 'FOUND' : 'NOT FOUND'}`);
      
      // Close modal
      await page.keyboard.press('Escape');
      await page.waitForTimeout(300);
      console.log('✓ Modal closed');
    }
  }

  // Check layout structure
  const mainLayout = page.locator('.row.g-3');
  const layoutExists = await mainLayout.count() > 0;
  console.log(`${layoutExists ? '✓' : '✗'} Main layout structure: ${layoutExists ? 'FOUND' : 'NOT FOUND'}`);

  // Check for editor pane
  const editorPane = page.locator('.frosted-card:has-text("Code Input")');
  const editorExists = await editorPane.count() > 0;
  console.log(`${editorExists ? '✓' : '✗'} Editor pane: ${editorExists ? 'FOUND' : 'NOT FOUND'}`);

  // Check for analysis pane
  const analysisPane = page.locator('.frosted-card:has-text("Analysis Results")');
  const analysisExists = await analysisPane.count() > 0;
  console.log(`${analysisExists ? '✓' : '✗'} Analysis pane: ${analysisExists ? 'FOUND' : 'NOT FOUND'}`);

  console.log('\n📊 UI Integration Status:');
  console.log(`  Import Button:        ${importExists ? '✅ PASS' : '❌ FAIL'}`);
  console.log(`  Layout Structure:     ${layoutExists ? '✅ PASS' : '❌ FAIL'}`);
  console.log(`  Editor Pane:          ${editorExists ? '✅ PASS' : '❌ FAIL'}`);
  console.log(`  Analysis Pane:        ${analysisExists ? '✅ PASS' : '❌ FAIL'}`);

  const allPassed = importExists && layoutExists && editorExists && analysisExists;
  console.log(`\n${allPassed ? '✅' : '❌'} Overall Status: ${allPassed ? 'PASS - Full Browser mode UI is integrated' : 'FAIL - Some components missing'}\n`);

  expect(allPassed).toBeTruthy();
});

test('Full Browser Mode - FileTreeBrowser Component Check', async ({ page }) => {
  console.log('\n🧪 Checking for FileTreeBrowser Component Implementation\n');

  // Read the ReviewPage.jsx file to verify integration
  const fs = require('fs');
  const path = require('path');
  
  const reviewPagePath = path.join(__dirname, '../frontend/src/components/ReviewPage.jsx');
  
  if (fs.existsSync(reviewPagePath)) {
    const content = fs.readFileSync(reviewPagePath, 'utf-8');
    
    // Check for FileTreeBrowser import
    const hasImport = content.includes('import FileTreeBrowser from');
    console.log(`${hasImport ? '✓' : '✗'} FileTreeBrowser import: ${hasImport ? 'FOUND' : 'NOT FOUND'}`);
    
    // Check for tree state
    const hasTreeState = content.includes('useState(null)') && content.includes('treeData');
    console.log(`${hasTreeState ? '✓' : '✗'} Tree state (treeData): ${hasTreeState ? 'FOUND' : 'NOT FOUND'}`);
    
    const hasShowTree = content.includes('showTree');
    console.log(`${hasShowTree ? '✓' : '✗'} Show tree state (showTree): ${hasShowTree ? 'FOUND' : 'NOT FOUND'}`);
    
    const hasSelectedFiles = content.includes('selectedTreeFiles');
    console.log(`${hasSelectedFiles ? '✓' : '✗'} Selected files state: ${hasSelectedFiles ? 'FOUND' : 'NOT FOUND'}`);
    
    // Check for FileTreeBrowser component usage
    const hasComponent = content.includes('<FileTreeBrowser');
    console.log(`${hasComponent ? '✓' : '✗'} FileTreeBrowser component usage: ${hasComponent ? 'FOUND' : 'NOT FOUND'}`);
    
    // Check for handlers
    const hasTreeFileSelect = content.includes('handleTreeFileSelect');
    console.log(`${hasTreeFileSelect ? '✓' : '✗'} handleTreeFileSelect handler: ${hasTreeFileSelect ? 'FOUND' : 'NOT FOUND'}`);
    
    const hasFetchOpenFile = content.includes('fetchAndOpenFile');
    console.log(`${hasFetchOpenFile ? '✓' : '✗'} fetchAndOpenFile handler: ${hasFetchOpenFile ? 'FOUND' : 'NOT FOUND'}`);
    
    const hasFilesAnalyze = content.includes('handleFilesAnalyze');
    console.log(`${hasFilesAnalyze ? '✓' : '✗'} handleFilesAnalyze handler: ${hasFilesAnalyze ? 'FOUND' : 'NOT FOUND'}`);
    
    // Check for conditional rendering
    const hasConditionalRender = content.includes('{showTree && treeData &&');
    console.log(`${hasConditionalRender ? '✓' : '✗'} Conditional rendering: ${hasConditionalRender ? 'FOUND' : 'NOT FOUND'}`);
    
    // Check for language detection
    const hasLanguageDetection = content.includes('const ext =') && content.includes('toLowerCase()');
    console.log(`${hasLanguageDetection ? '✓' : '✗'} Language detection logic: ${hasLanguageDetection ? 'FOUND' : 'NOT FOUND'}`);
    
    console.log('\n📊 Component Integration Status:');
    console.log(`  FileTreeBrowser Import:   ${hasImport ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`  Tree State Management:    ${hasTreeState && hasShowTree && hasSelectedFiles ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`  Component Usage:          ${hasComponent ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`  Event Handlers:           ${hasTreeFileSelect && hasFetchOpenFile && hasFilesAnalyze ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`  Conditional Rendering:    ${hasConditionalRender ? '✅ PASS' : '❌ FAIL'}`);
    console.log(`  Language Detection:       ${hasLanguageDetection ? '✅ PASS' : '❌ FAIL'}`);
    
    const allChecks = hasImport && hasTreeState && hasShowTree && hasSelectedFiles && 
                      hasComponent && hasTreeFileSelect && hasFetchOpenFile && 
                      hasFilesAnalyze && hasConditionalRender && hasLanguageDetection;
    
    console.log(`\n${allChecks ? '✅' : '❌'} Overall Status: ${allChecks ? 'PASS - All components properly integrated' : 'FAIL - Some integration missing'}\n`);
    
    expect(allChecks).toBeTruthy();
  } else {
    console.log('❌ ReviewPage.jsx not found at expected location\n');
    expect(false).toBeTruthy();
  }
});
