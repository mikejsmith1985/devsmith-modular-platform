import { test, expect } from './fixtures/auth.fixture';

test('Navigation button hover states work correctly', async ({ authenticatedPage }) => {
  console.log('=== NAVIGATION BUTTON HOVER VALIDATION ===\n');

  // Test Logs service via authenticated portal dashboard path
  await authenticatedPage.goto('/dashboard/logs');
  await authenticatedPage.waitForLoadState('networkidle');

  // Find the dark mode toggle button
  const darkModeButton = authenticatedPage.locator('button.btn-icon#dark-mode-toggle');
  await expect(darkModeButton).toBeVisible();

  // Get default state (should be transparent)
  const defaultBg = await darkModeButton.evaluate((el) => {
    return window.getComputedStyle(el).backgroundColor;
  });

  console.log('LOGS SERVICE - Dark Mode Toggle:');
  console.log('  Default background:', defaultBg);

  // Hover over button
  await darkModeButton.hover();
  await authenticatedPage.waitForTimeout(300); // Wait for transition

  // Get hover state (should have background)
  const hoverBg = await darkModeButton.evaluate((el) => {
    return window.getComputedStyle(el).backgroundColor;
  });

  console.log('  Hover background:', hoverBg);
  console.log('');

  // Validate
  expect(defaultBg).toBe('rgba(0, 0, 0, 0)'); // Transparent by design
  expect(hoverBg).not.toBe('rgba(0, 0, 0, 0)'); // Should have background on hover
  expect(hoverBg).toMatch(/^rgb/); // Should be an RGB color

  // Test mobile menu button
  const menuButton = authenticatedPage.locator('button.btn-icon[data-app-menu]');
  if (await menuButton.isVisible()) {
    const menuDefaultBg = await menuButton.evaluate((el) => 
      window.getComputedStyle(el).backgroundColor
    );
    
    await menuButton.hover();
    await authenticatedPage.waitForTimeout(300);
    
    const menuHoverBg = await menuButton.evaluate((el) => 
      window.getComputedStyle(el).backgroundColor
    );

    console.log('LOGS SERVICE - Mobile Menu Toggle:');
    console.log('  Default background:', menuDefaultBg);
    console.log('  Hover background:', menuHoverBg);
    console.log('');

    expect(menuDefaultBg).toBe('rgba(0, 0, 0, 0)');
    expect(menuHoverBg).not.toBe('rgba(0, 0, 0, 0)');
  }

  // Test Analytics service
  // Test Analytics service via portal SPA path
  await authenticatedPage.goto('/analytics');
  await authenticatedPage.waitForLoadState('networkidle');

  const analyticsButton = authenticatedPage.locator('button.btn-icon#dark-mode-toggle');
  await expect(analyticsButton).toBeVisible();

  const analyticsDefaultBg = await analyticsButton.evaluate((el) => 
    window.getComputedStyle(el).backgroundColor
  );

  await analyticsButton.hover();
  await authenticatedPage.waitForTimeout(300);

  const analyticsHoverBg = await analyticsButton.evaluate((el) => 
    window.getComputedStyle(el).backgroundColor
  );

  console.log('ANALYTICS SERVICE - Dark Mode Toggle:');
  console.log('  Default background:', analyticsDefaultBg);
  console.log('  Hover background:', analyticsHoverBg);
  console.log('');

  expect(analyticsDefaultBg).toBe('rgba(0, 0, 0, 0)');
  expect(analyticsHoverBg).not.toBe('rgba(0, 0, 0, 0)');

  console.log('✅ ALL HOVER STATES WORKING CORRECTLY');
  console.log('✅ Default: Transparent (intentional design for icon buttons)');
  console.log('✅ Hover: Colored background (user feedback)');
});
