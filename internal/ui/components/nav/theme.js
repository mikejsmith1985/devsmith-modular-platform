// theme.js - Dark mode theme management

(function() {
  const THEME_KEY = 'devsmith-theme';
  const LIGHT_THEME = 'light';
  const DARK_THEME = 'dark';

  // Initialize theme on page load
  function initTheme() {
    const html = document.documentElement;
    const themeToggle = document.getElementById('theme-toggle');
    
    // Get stored preference or use system preference
    let theme = localStorage.getItem(THEME_KEY);
    if (!theme) {
      theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? DARK_THEME : LIGHT_THEME;
    }
    
    applyTheme(theme);
    updateThemeIcon(theme);
    
    // Setup toggle button
    if (themeToggle) {
      themeToggle.addEventListener('click', toggleTheme);
    }
  }

  function applyTheme(theme) {
    const html = document.documentElement;
    if (theme === DARK_THEME) {
      html.setAttribute('data-theme', 'dark');
      html.style.colorScheme = 'dark';
    } else {
      html.removeAttribute('data-theme');
      html.style.colorScheme = 'light';
    }
    localStorage.setItem(THEME_KEY, theme);
  }

  function toggleTheme() {
    const html = document.documentElement;
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === DARK_THEME ? LIGHT_THEME : DARK_THEME;
    applyTheme(newTheme);
    updateThemeIcon(newTheme);
  }

  function updateThemeIcon(theme) {
    const sunIcon = document.querySelector('.sun-icon');
    const moonIcon = document.querySelector('.moon-icon');
    
    if (theme === DARK_THEME) {
      sunIcon.style.display = 'none';
      moonIcon.style.display = 'block';
    } else {
      sunIcon.style.display = 'block';
      moonIcon.style.display = 'none';
    }
  }

  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTheme);
  } else {
    initTheme();
  }
})();
