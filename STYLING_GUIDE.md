# DevSmith Styling Guide

Based on DevSmith Logs design system.

## Color Palette

### Dark Mode (Primary)
```css
--bg-primary: #1a1a2e;
--bg-secondary: #16213e;
--bg-card: rgba(255, 255, 255, 0.05);
--text-primary: #eaeaea;
--text-secondary: #a0a0a0;

--debug-color: #20c997;
--info-color: #0dcaf0;
--warning-color: #ffc107;
--error-color: #dc3545;
--critical-color: #d63384;
```

### Light Mode
```css
--bg-primary: #ffffff;
--bg-secondary: #f8f9fa;
--bg-card: rgba(0, 0, 0, 0.02);
--text-primary: #212529;
--text-secondary: #6c757d;
```

## Component Patterns

### Frosted Card
```css
.frosted-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}
```

### Log Level Badges
```css
.badge-debug { background: var(--debug-color); }
.badge-info { background: var(--info-color); }
.badge-warning { background: var(--warning-color); }
.badge-error { background: var(--error-color); }
.badge-critical { background: var(--critical-color); }
```

## Layout Guidelines

- Max content width: `1200px`
- Card spacing: `1.5rem` padding
- Border radius: `8px` (small), `12px` (medium), `16px` (large)
- Use Bootstrap grid system
- Mobile-first responsive design

## Typography

- Headers: System font stack
- Body: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto
- Monospace (code): 'Courier New', monospace

## Accessibility

- Minimum contrast ratio: 4.5:1
- Focus indicators on interactive elements
- Semantic HTML
- ARIA labels where needed

---
