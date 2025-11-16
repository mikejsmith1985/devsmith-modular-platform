/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./apps/**/*.templ",
    "./apps/**/*.html",
    "./internal/ui/**/*.templ",
  ],
  darkMode: 'class', // Enable class-based dark mode (via 'dark' class on <html>)
  theme: {
    extend: {},
  },
  plugins: [],
}
