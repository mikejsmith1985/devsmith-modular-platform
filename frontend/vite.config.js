import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  publicDir: 'public',  // Explicit public directory for static assets
  base: '/',            // Explicit base path
  server: {
    host: '0.0.0.0',
    port: 5173
  },
  build: {
    // CRITICAL FIX: Disable all caching mechanisms to prevent stale code
    minify: false,        // Expose raw variable names (temporary for debugging)
    cssCodeSplit: false,  // Bundle all CSS into one file
    rollupOptions: {
      cache: false,       // Disable Rollup's persistent module graph cache
      output: {
        // Remove manualChunks to prevent cascading hash issues
      }
    }
  }
})
