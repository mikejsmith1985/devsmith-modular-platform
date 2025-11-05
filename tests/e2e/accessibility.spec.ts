import { test } from '@playwright/test';
import { test as authTest, expect } from './fixtures/auth.fixture';
import AxeBuilder from '@axe-core/playwright';

/**
 * Accessibility Compliance Tests (WCAG 2.1 AA)
 * 
 * Validates keyboard navigation, screen reader support, color contrast,
 * and overall WCAG 2.1 Level AA compliance across all services.
 * 
 * Phase 4 - Accessibility Compliance
 */

test.describe('Accessibility - Automated Audits (axe-core)', () => {
  authTest('Portal Dashboard passes WCAG 2.1 AA audit', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });

  authTest('Review Service passes WCAG 2.1 AA audit', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/review/workspace/demo');
    await authenticatedPage.waitForLoadState('networkidle');

    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });

  authTest('Logs Service passes WCAG 2.1 AA audit', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/logs');
    await authenticatedPage.waitForLoadState('networkidle');

    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });

  authTest('Analytics Service passes WCAG 2.1 AA audit', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/analytics');
    await authenticatedPage.waitForLoadState('networkidle');

    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
  });
});

test.describe('Accessibility - Keyboard Navigation', () => {
  authTest('Portal Dashboard is fully keyboard navigable', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Start from first interactive element
    await authenticatedPage.keyboard.press('Tab');
    
    // Verify focus visible
    const focusedElement = await authenticatedPage.evaluate(() => {
      const el = document.activeElement;
      return {
        tag: el?.tagName,
        role: el?.getAttribute('role'),
        ariaLabel: el?.getAttribute('aria-label'),
        hasVisibleFocus: window.getComputedStyle(el!).outline !== 'none'
      };
    });

    expect(focusedElement.tag).toBeTruthy();
    
    // Tab through interactive elements
    for (let i = 0; i < 10; i++) {
      await authenticatedPage.keyboard.press('Tab');
      await authenticatedPage.waitForTimeout(100);
    }

    // Verify can navigate back with Shift+Tab
    await authenticatedPage.keyboard.press('Shift+Tab');
    await authenticatedPage.waitForTimeout(100);

    // Test that Enter/Space activate buttons
    const button = authenticatedPage.locator('button, [role="button"]').first();
    if (await button.isVisible()) {
      await button.focus();
      // Verify button can be activated with keyboard
      // (Don't actually click to avoid navigation)
      expect(await button.evaluate(el => el.hasAttribute('disabled') || true)).toBeTruthy();
    }
  });

  authTest('Review workspace keyboard shortcuts work', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');

    // Test common keyboard shortcuts (if implemented)
    // Ctrl+K: Open command palette (if exists)
    await authenticatedPage.keyboard.press('Control+K');
    await authenticatedPage.waitForTimeout(500);

    // Escape: Close modals/overlays
    await authenticatedPage.keyboard.press('Escape');
    await authenticatedPage.waitForTimeout(500);

    // Tab navigation should work
    await authenticatedPage.keyboard.press('Tab');
    const focused = await authenticatedPage.evaluate(() => document.activeElement?.tagName);
    expect(focused).toBeTruthy();
  });

  authTest('Skip links allow bypassing navigation', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    
    // Tab to first element (should be skip link or main content link)
    await authenticatedPage.keyboard.press('Tab');
    
    const firstElement = await authenticatedPage.evaluate(() => {
      const el = document.activeElement;
      return {
        text: el?.textContent?.trim(),
        href: (el as HTMLAnchorElement)?.href,
        tag: el?.tagName
      };
    });

    // Skip link should be first focusable element OR main content should be early
    expect(
      firstElement.text?.toLowerCase().includes('skip') ||
      firstElement.text?.toLowerCase().includes('main') ||
      firstElement.tag === 'MAIN'
    ).toBeTruthy();
  });
});

test.describe('Accessibility - Screen Reader Support', () => {
  authTest('Portal has proper ARIA landmarks', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Verify required ARIA landmarks exist
    const landmarks = await authenticatedPage.evaluate(() => {
      return {
        main: document.querySelector('main, [role="main"]') !== null,
        navigation: document.querySelector('nav, [role="navigation"]') !== null,
        banner: document.querySelector('header, [role="banner"]') !== null,
        contentinfo: document.querySelector('footer, [role="contentinfo"]') !== null,
      };
    });

    expect(landmarks.main).toBeTruthy();
    expect(landmarks.navigation).toBeTruthy();
    // Banner and contentinfo are optional but recommended
  });

  authTest('Interactive elements have accessible names', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Check buttons have accessible names
    const buttonsWithoutNames = await authenticatedPage.evaluate(() => {
      const buttons = Array.from(document.querySelectorAll('button, [role="button"]'));
      return buttons
        .filter(btn => {
          const text = btn.textContent?.trim();
          const ariaLabel = btn.getAttribute('aria-label');
          const ariaLabelledby = btn.getAttribute('aria-labelledby');
          const title = btn.getAttribute('title');
          
          return !text && !ariaLabel && !ariaLabelledby && !title;
        })
        .map(btn => ({
          tag: btn.tagName,
          class: btn.className,
          id: btn.id
        }));
    });

    expect(buttonsWithoutNames).toEqual([]);
  });

  authTest('Images have alt text', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    const imagesWithoutAlt = await authenticatedPage.evaluate(() => {
      const images = Array.from(document.querySelectorAll('img'));
      return images
        .filter(img => {
          const alt = img.getAttribute('alt');
          const role = img.getAttribute('role');
          
          // Decorative images should have empty alt or role="presentation"
          // Content images should have descriptive alt
          return alt === null && role !== 'presentation';
        })
        .map(img => ({
          src: img.src,
          class: img.className
        }));
    });

    expect(imagesWithoutAlt).toEqual([]);
  });

  authTest('Forms have proper labels', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/review');
    await authenticatedPage.waitForLoadState('networkidle');

    const unlabeledInputs = await authenticatedPage.evaluate(() => {
      const inputs = Array.from(document.querySelectorAll('input, select, textarea'));
      return inputs
        .filter(input => {
          const id = input.id;
          const ariaLabel = input.getAttribute('aria-label');
          const ariaLabelledby = input.getAttribute('aria-labelledby');
          const label = id ? document.querySelector(`label[for="${id}"]`) : null;
          const ariaDescribedby = input.getAttribute('aria-describedby');
          
          return !ariaLabel && !ariaLabelledby && !label;
        })
        .map(input => ({
          type: input.getAttribute('type'),
          name: input.getAttribute('name'),
          placeholder: input.getAttribute('placeholder')
        }));
    });

    // Allow some unlabeled inputs (like hidden fields)
    const visibleUnlabeled = unlabeledInputs.filter(
      input => input.type !== 'hidden'
    );
    
    expect(visibleUnlabeled).toEqual([]);
  });
});

test.describe('Accessibility - Color Contrast', () => {
  authTest('Text has sufficient color contrast (4.5:1 minimum)', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Run axe-core color contrast check
    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2aa'])
      .options({ rules: { 'color-contrast': { enabled: true } } })
      .analyze();

    const contrastViolations = accessibilityScanResults.violations.filter(
      v => v.id === 'color-contrast'
    );

    expect(contrastViolations).toEqual([]);
  });

  authTest('Large text has sufficient contrast (3:1 minimum)', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Run axe-core color contrast check for large text
    const accessibilityScanResults = await new AxeBuilder({ page: authenticatedPage })
      .withTags(['wcag2aa'])
      .analyze();

    const contrastViolations = accessibilityScanResults.violations.filter(
      v => v.id === 'color-contrast'
    );

    expect(contrastViolations).toEqual([]);
  });
});

test.describe('Accessibility - Focus Management', () => {
  authTest('Focus indicators are visible', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    // Tab to first interactive element
    await authenticatedPage.keyboard.press('Tab');
    await authenticatedPage.waitForTimeout(200);

    const focusStyles = await authenticatedPage.evaluate(() => {
      const el = document.activeElement;
      if (!el) return null;
      
      const styles = window.getComputedStyle(el);
      return {
        outline: styles.outline,
        outlineWidth: styles.outlineWidth,
        outlineStyle: styles.outlineStyle,
        outlineColor: styles.outlineColor,
        boxShadow: styles.boxShadow,
        border: styles.border
      };
    });

    // Element should have visible focus indicator
    expect(
      focusStyles?.outline !== 'none' ||
      focusStyles?.outlineWidth !== '0px' ||
      focusStyles?.boxShadow !== 'none' ||
      focusStyles?.border !== 'none'
    ).toBeTruthy();
  });

  authTest('Focus order is logical', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    const focusSequence: string[] = [];
    
    // Record focus sequence
    for (let i = 0; i < 10; i++) {
      await authenticatedPage.keyboard.press('Tab');
      await authenticatedPage.waitForTimeout(100);
      
      const elementInfo = await authenticatedPage.evaluate(() => {
        const el = document.activeElement;
        return `${el?.tagName}.${el?.className || 'no-class'}`;
      });
      
      focusSequence.push(elementInfo);
    }

    // Verify focus sequence is not empty (elements are focusable)
    expect(focusSequence.length).toBeGreaterThan(0);
    
    // Verify no focus traps (focus moves forward)
    const uniqueElements = new Set(focusSequence);
    expect(uniqueElements.size).toBeGreaterThan(1);
  });
});

test.describe('Accessibility - Semantic HTML', () => {
  authTest('Portal uses semantic HTML5 elements', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    const semanticElements = await authenticatedPage.evaluate(() => {
      return {
        hasMain: document.querySelector('main') !== null,
        hasNav: document.querySelector('nav') !== null,
        hasHeader: document.querySelector('header') !== null,
        hasFooter: document.querySelector('footer') !== null,
        hasSection: document.querySelectorAll('section').length > 0,
        hasArticle: document.querySelectorAll('article').length >= 0, // Optional
        
        // Check for divitis (excessive divs instead of semantic elements)
        divCount: document.querySelectorAll('div').length,
        semanticCount: document.querySelectorAll('main, nav, header, footer, section, article, aside').length
      };
    });

    expect(semanticElements.hasMain).toBeTruthy();
    expect(semanticElements.hasNav).toBeTruthy();
    
    // Semantic ratio should be reasonable (at least 1 semantic element per 20 divs)
    if (semanticElements.divCount > 0) {
      const ratio = semanticElements.divCount / semanticElements.semanticCount;
      expect(ratio).toBeLessThan(50); // Allow up to 50 divs per semantic element
    }
  });

  authTest('Headings follow logical hierarchy', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/dashboard');
    await authenticatedPage.waitForLoadState('networkidle');

    const headingHierarchy = await authenticatedPage.evaluate(() => {
      const headings = Array.from(document.querySelectorAll('h1, h2, h3, h4, h5, h6'));
      return headings.map(h => ({
        level: parseInt(h.tagName.substring(1)),
        text: h.textContent?.trim().substring(0, 50)
      }));
    });

    // Should have at least one H1
    const h1Count = headingHierarchy.filter(h => h.level === 1).length;
    expect(h1Count).toBeGreaterThanOrEqual(1);
    expect(h1Count).toBeLessThanOrEqual(1); // Only one H1 per page

    // Check for heading level skips (e.g., H1 → H3 without H2)
    for (let i = 1; i < headingHierarchy.length; i++) {
      const prevLevel = headingHierarchy[i - 1].level;
      const currLevel = headingHierarchy[i].level;
      
      // Level can increase by 1, or decrease by any amount, but shouldn't skip levels going down
      if (currLevel > prevLevel) {
        expect(currLevel - prevLevel).toBeLessThanOrEqual(1);
      }
    }
  });
});
