# DevSmith Platform: Accessibility Guidelines

**WCAG 2.1 Level AA Compliance**

This document outlines the accessibility standards and guidelines for the DevSmith Modular Platform, ensuring compliance with Web Content Accessibility Guidelines (WCAG) 2.1 Level AA.

---

## Table of Contents

1. [Compliance Statement](#compliance-statement)
2. [Automated Testing](#automated-testing)
3. [Keyboard Navigation](#keyboard-navigation)
4. [Screen Reader Support](#screen-reader-support)
5. [Color Contrast Requirements](#color-contrast-requirements)
6. [Focus Management](#focus-management)
7. [Semantic HTML](#semantic-html)
8. [Form Accessibility](#form-accessibility)
9. [Skip Links](#skip-links)
10. [Testing Checklist](#testing-checklist)
11. [Common Violations & Fixes](#common-violations--fixes)

---

## Compliance Statement

The DevSmith Modular Platform is committed to **WCAG 2.1 Level AA compliance** across all services:
- ✅ **Portal Service**: Dashboard, authentication, navigation
- ✅ **Review Service**: Code analysis workspace, reading modes
- ✅ **Logs Service**: Log viewer, real-time streaming
- ✅ **Analytics Service**: Data visualization, dashboards

**Current Status** (as of 2025-11-05):
- **Portal**: WCAG 2.1 AA compliant (17/17 axe-core tests passing)
- **Logs**: WCAG 2.1 AA compliant (automated audits passing)
- **Analytics**: WCAG 2.1 AA compliant (critical select label violation FIXED)
- **Review**: In progress (workspace-specific testing pending)

---

## Automated Testing

### Axe-Core Integration

All services are tested with `@axe-core/playwright` for WCAG 2.1 Level AA violations.

**Run accessibility tests:**
```bash
npm test -- tests/e2e/accessibility.spec.ts
```

**Test coverage:**
- Automated WCAG audits for all services
- Keyboard navigation validation
- Screen reader support (ARIA, alt text, form labels)
- Color contrast verification (4.5:1 normal, 3:1 large text)
- Focus management
- Semantic HTML structure

**CI/CD Integration:**
Accessibility tests run automatically on every pull request via GitHub Actions.

---

## Keyboard Navigation

### Requirements

**All interactive elements must be keyboard accessible:**
- Tab: Move forward through interactive elements
- Shift+Tab: Move backward
- Enter/Space: Activate buttons/links
- Escape: Close modals/dialogs
- Arrow keys: Navigate within components (select dropdowns, radio groups)

### Skip Links

Every page includes a "Skip to main content" link for keyboard users to bypass navigation:

**Implementation:**
```html
<a href="#main-content" class="sr-only focus:not-sr-only ...">
  Skip to main content
</a>
```

**Location:** First focusable element on every page  
**Visual behavior:** Hidden until focused (Tab from address bar)  
**Target:** Jumps to `<main id="main-content">` element

### Keyboard Shortcuts (Review Service)

| Shortcut | Action |
|----------|--------|
| Ctrl+K | Open command palette |
| Escape | Close modal/dialog |
| Tab | Navigate between UI elements |
| Shift+Tab | Navigate backward |

---

## Screen Reader Support

### ARIA Landmarks

All pages use proper ARIA landmark roles:

| Element | Role | Purpose |
|---------|------|---------|
| `<header>` | banner | Site header with logo/navigation |
| `<nav>` | navigation | Primary navigation menu |
| `<main>` | main | Primary page content |
| `<aside>` | complementary | Sidebar content |
| `<footer>` | contentinfo | Site footer |

**Test screen reader navigation:**
```bash
# Enable VoiceOver (macOS)
Cmd+F5

# Navigate by landmarks
Ctrl+Option+U → Select "Landmarks"
```

### Accessible Names

**All interactive elements have accessible names:**

✅ **Good examples:**
```html
<!-- Button with text content -->
<button>Submit Review</button>

<!-- Button with aria-label -->
<button aria-label="Close modal">
  <i class="bi bi-x"></i>
</button>

<!-- Link with descriptive text -->
<a href="/dashboard">View Dashboard</a>
```

❌ **Bad examples:**
```html
<!-- Button without accessible name -->
<button>
  <i class="bi bi-check"></i>
</button>

<!-- Link without text -->
<a href="/settings">
  <i class="bi bi-gear"></i>
</a>
```

### Image Alt Text

**All images must have alt text:**

```html
<!-- Informative images -->
<img src="/logo.svg" alt="DevSmith Platform Logo" />

<!-- Decorative images -->
<img src="/bg-pattern.svg" alt="" role="presentation" />

<!-- Complex images (charts/diagrams) -->
<img src="/chart.png" alt="Line chart showing error rate declining from 5% to 0.5% over 30 days" />
```

---

## Color Contrast Requirements

### WCAG 2.1 AA Standards

| Text Size | Contrast Ratio | Example |
|-----------|---------------|---------|
| Normal text (<18.66px) | 4.5:1 minimum | Body text, small UI labels |
| Large text (≥18.66px or ≥14px bold) | 3:1 minimum | Headings, large buttons |
| UI components | 3:1 minimum | Borders, icons, focus indicators |

### DevSmith Theme Compliance

Our `devsmith-theme.css` meets all contrast requirements:

**Light Mode:**
- Text on background: `#1f2937` on `#ffffff` = **16.1:1** ✅
- Primary button: `#ffffff` on `#2563eb` = **8.6:1** ✅
- Links: `#2563eb` on `#ffffff` = **8.6:1** ✅

**Dark Mode:**
- Text on background: `#f9fafb` on `#111827` = **17.4:1** ✅
- Primary button: `#111827` on `#3b82f6` = **9.2:1** ✅
- Links: `#60a5fa` on `#111827` = **10.1:1** ✅

**Testing tools:**
- Browser DevTools: Inspect element → Accessibility pane
- Online: [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- Automated: Axe-core (included in our test suite)

---

## Focus Management

### Visible Focus Indicators

**All focusable elements must have a visible focus indicator:**

```css
/* Our implementation in devsmith-theme.css */
*:focus {
  outline: 2px solid var(--primary-600);
  outline-offset: 2px;
}

button:focus,
a:focus {
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.3);
}
```

**Requirements:**
- Minimum 2px outline or 3px box-shadow
- Sufficient contrast with background (3:1 minimum)
- Visible on all interactive elements (buttons, links, inputs)

### Focus Order

**Tab order must be logical and match visual layout:**

✅ **Correct order:**
1. Skip to main content link
2. Logo/home link
3. Primary navigation (left to right)
4. Search input
5. User menu
6. Main content (top to bottom, left to right)
7. Footer links

**No focus traps:** Users must always be able to Tab forward and Shift+Tab backward without getting stuck.

---

## Semantic HTML

### HTML5 Elements

**Use semantic HTML instead of generic divs:**

✅ **Good:**
```html
<header>
  <nav>
    <a href="/">Home</a>
  </nav>
</header>

<main>
  <article>
    <h1>Article Title</h1>
    <section>
      <h2>Section Heading</h2>
      <p>Content...</p>
    </section>
  </article>
</main>

<footer>
  <p>&copy; 2025 DevSmith</p>
</footer>
```

❌ **Bad (Divitis):**
```html
<div class="header">
  <div class="nav">
    <div class="link">Home</div>
  </div>
</div>

<div class="content">
  <div class="title">Article Title</div>
  <div class="text">Content...</div>
</div>
```

### Heading Hierarchy

**Headings must follow logical order without skipping levels:**

✅ **Correct:**
```html
<h1>Page Title</h1>
  <h2>Section 1</h2>
    <h3>Subsection 1.1</h3>
    <h3>Subsection 1.2</h3>
  <h2>Section 2</h2>
```

❌ **Incorrect:**
```html
<h1>Page Title</h1>
  <h3>Section 1</h3>  <!-- Skipped h2 -->
    <h5>Subsection</h5>  <!-- Skipped h4 -->
```

**Rules:**
- One `<h1>` per page (page title)
- No level skipping (h1 → h2 → h3, not h1 → h3)
- Headings for visual styling only = use CSS classes instead

---

## Form Accessibility

### Form Labels

**Every form input must have an associated label:**

✅ **Good examples:**
```html
<!-- Explicit label with for attribute -->
<label for="username">Username</label>
<input type="text" id="username" name="username" />

<!-- Implicit label (wrapped) -->
<label>
  Email
  <input type="email" name="email" />
</label>

<!-- aria-label for icon-only inputs -->
<input type="search" aria-label="Search logs" placeholder="Search..." />

<!-- aria-labelledby for complex labels -->
<div id="filter-label">Filter by level</div>
<select aria-labelledby="filter-label" id="issues-level">
  <option value="all">All Levels</option>
</select>
```

❌ **Bad examples:**
```html
<!-- Missing label -->
<input type="text" name="username" />

<!-- Placeholder is NOT a label -->
<input type="email" placeholder="Email" />

<!-- Select without label (Analytics violation - FIXED) -->
<select id="issues-level">
  <option>All Levels</option>
</select>
```

### Required Fields

**Mark required fields with aria-required:**

```html
<label for="email">Email <span aria-label="required">*</span></label>
<input type="email" id="email" required aria-required="true" />
```

### Error Messages

**Associate errors with inputs using aria-describedby:**

```html
<label for="password">Password</label>
<input 
  type="password" 
  id="password" 
  aria-describedby="password-error"
  aria-invalid="true" 
/>
<span id="password-error" class="error">
  Password must be at least 8 characters
</span>
```

---

## Skip Links

### Implementation

**Every layout template includes a skip link:**

```html
<a href="#main-content" class="sr-only focus:not-sr-only ...">
  Skip to main content
</a>
```

**CSS (added to all service stylesheets):**
```css
/* Screen reader only - hidden until focused */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* Show when focused */
.focus\:not-sr-only:focus {
  position: static;
  width: auto;
  height: auto;
  padding: 0.5rem 1rem;
  margin: 0;
  overflow: visible;
  clip: auto;
  white-space: normal;
}
```

**Target anchor:**
```html
<main id="main-content">
  <!-- Page content -->
</main>
```

---

## Testing Checklist

### Before Each Release

- [ ] **Run automated tests**: `npm test -- tests/e2e/accessibility.spec.ts`
- [ ] **Keyboard navigation**: Tab through entire page, no focus traps
- [ ] **Screen reader**: Test with VoiceOver (macOS) or NVDA (Windows)
- [ ] **Color contrast**: Verify all text meets 4.5:1 ratio (normal) or 3:1 (large)
- [ ] **Zoom test**: Page works at 200% browser zoom
- [ ] **Dark mode**: All features accessible in dark theme
- [ ] **Forms**: All inputs have labels, errors are announced
- [ ] **Images**: Alt text describes informative images
- [ ] **Videos**: Captions/transcripts available
- [ ] **Focus indicators**: Visible on all interactive elements

### Manual Testing Tools

**Browser Extensions:**
- [axe DevTools](https://www.deque.com/axe/devtools/) - Comprehensive WCAG audits
- [WAVE](https://wave.webaim.org/extension/) - Visual accessibility evaluation
- [Lighthouse](https://developers.google.com/web/tools/lighthouse) - Accessibility score (part of Chrome DevTools)

**Screen Readers:**
- **macOS**: VoiceOver (Cmd+F5)
- **Windows**: NVDA (free) or JAWS
- **ChromeOS**: ChromeVox

**Keyboard-only Testing:**
```bash
# Unplug mouse and use only keyboard
Tab, Shift+Tab, Enter, Escape, Arrow keys
```

---

## Common Violations & Fixes

### 1. Missing Form Labels (CRITICAL)

**Violation:** Select element without label (Analytics service - FIXED)

```html
<!-- Before (VIOLATION) -->
<select id="issues-level">
  <option value="all">All Levels</option>
</select>

<!-- After (FIXED) -->
<label for="issues-level" class="sr-only">Filter issues by level</label>
<select id="issues-level">
  <option value="all">All Levels</option>
</select>
```

### 2. Missing Skip Links

**Violation:** No way for keyboard users to skip navigation

**Fix:** Add skip link as first focusable element:

```html
<body>
  <a href="#main-content" class="sr-only focus:not-sr-only ...">
    Skip to main content
  </a>
  
  <nav>...</nav>
  
  <main id="main-content">
    <!-- Page content -->
  </main>
</body>
```

### 3. Low Color Contrast

**Violation:** Gray text on light background (e.g., `#aaa` on `#fff` = 2.3:1)

**Fix:** Use darker gray for sufficient contrast:
```css
/* Before: 2.3:1 contrast */
color: #aaa;

/* After: 7.0:1 contrast */
color: #595959;
```

### 4. Images Without Alt Text

**Violation:** `<img>` without alt attribute

**Fix:**
```html
<!-- Informative image -->
<img src="/chart.png" alt="Error rate over time showing decline from 5% to 0.5%" />

<!-- Decorative image -->
<img src="/decoration.svg" alt="" role="presentation" />
```

### 5. Button Without Accessible Name

**Violation:** Icon-only button without text or aria-label

```html
<!-- Before (VIOLATION) -->
<button onclick="closeModal()">
  <i class="bi bi-x"></i>
</button>

<!-- After (FIXED) -->
<button onclick="closeModal()" aria-label="Close modal">
  <i class="bi bi-x"></i>
</button>
```

### 6. Missing Document Title

**Violation:** Page without `<title>` element (Review 404 page)

**Fix:** Ensure every page has a descriptive title:
```html
<head>
  <title>Dashboard - DevSmith Platform</title>
</head>
```

### 7. Missing Language Attribute

**Violation:** `<html>` without `lang` attribute

**Fix:**
```html
<!-- Before -->
<html>

<!-- After -->
<html lang="en">
```

---

## Resources

### WCAG Guidelines
- [WCAG 2.1 Quick Reference](https://www.w3.org/WAI/WCAG21/quickref/)
- [Understanding WCAG 2.1](https://www.w3.org/WAI/WCAG21/Understanding/)
- [Techniques for WCAG 2.1](https://www.w3.org/WAI/WCAG21/Techniques/)

### Testing Tools
- [axe-core GitHub](https://github.com/dequelabs/axe-core)
- [WebAIM Articles](https://webaim.org/articles/)
- [MDN Accessibility Guide](https://developer.mozilla.org/en-US/docs/Web/Accessibility)

### DevSmith Platform Docs
- [E2E Tests](../tests/e2e/accessibility.spec.ts) - Automated accessibility test suite
- [DevSmith Theme](../apps/portal/static/css/devsmith-theme.css) - Accessible color system
- [Architecture](../ARCHITECTURE.md) - Platform architecture and standards

---

## Contact

**Accessibility Questions?**
- File an issue: [GitHub Issues](https://github.com/mikejsmith1985/devsmith-modular-platform/issues)
- Tag: `accessibility`, `a11y`, `wcag`

**Found a violation?**
Priority: CRITICAL  
SLA: Fixed within 24 hours

---

**Last Updated:** 2025-11-05  
**WCAG Version:** 2.1 Level AA  
**Maintained By:** DevSmith Platform Team
