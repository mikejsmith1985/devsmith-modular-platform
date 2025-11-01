/**
 * Dark Mode Toggle - Vanilla JavaScript implementation
 * Handles dark mode switching with localStorage persistence
 * No dependencies required (no Alpine.js)
 */

function initDarkModeToggle() {
  const html = document.documentElement;
  const toggleBtn = document.getElementById('dark-mode-toggle');
  
  if (!toggleBtn) return;
  
  // Load saved preference or detect system preference
  const savedMode = localStorage.getItem('darkMode');
  const isDark = savedMode !== null 
    ? savedMode === 'true' 
    : window.matchMedia('(prefers-color-scheme: dark)').matches;
  
  // Apply saved/detected preference
  if (isDark) {
    html.classList.add('dark');
  } else {
    html.classList.remove('dark');
  }
  
  // Toggle handler
  toggleBtn.addEventListener('click', () => {
    const isCurrentlyDark = html.classList.toggle('dark');
    localStorage.setItem('darkMode', isCurrentlyDark);
  });
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initDarkModeToggle);
} else {
  initDarkModeToggle();
}
