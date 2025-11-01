import { defineConfig, devices } from '@playwright/test';
import * as os from 'os';

/**
 * Playwright Configuration for DevSmith E2E Tests
 * 
 * This configuration provides:
 * - Proper test directory scoping (./tests/e2e only)
 * - Multiple test projects (quick, full)
 * - Limited workers to prevent resource contention
 * - Optimized timeouts and retry behavior
 * - Local and CI/CD support
 * - Basic Auth support for nginx-protected endpoints
 */

// Determine base URL based on platform (Linux vs Docker Desktop)
const isDockerDesktop = ['darwin', 'win32'].includes(os.platform());
const host = isDockerDesktop ? 'host.docker.internal' : 'localhost';
const baseURL = `http://${host}:3000`;

// Build Basic Auth header if credentials provided
const getAuthHeader = () => {
  const basicAuth = process.env.PLAYWRIGHT_BASIC_AUTH;
  if (!basicAuth) return {};
  
  const encoded = Buffer.from(basicAuth).toString('base64');
  return {
    Authorization: `Basic ${encoded}`,
  };
};

export default defineConfig({
  // Only run E2E tests (./tests/e2e)
  testDir: './tests/e2e',
  
  // Glob patterns to ignore
  testIgnore: '**/node_modules/**',

  // Output directory for test results
  outputDir: 'test-results',

  // Test configuration
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  
  // Limit workers to prevent resource contention
  // Local: 2 workers, CI: 1 worker per job
  workers: process.env.CI ? 1 : 2,

  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list'],
  ],

  // Shared settings for all projects
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000,
    extraHTTPHeaders: getAuthHeader(),
  },

  // Global timeout settings
  timeout: 30000,
  expect: {
    timeout: 5000,
  },

  // WebSocket timeout
  webServer: undefined, // Services run in Docker, not spawned by Playwright

  // Test projects: quick for fast feedback, full for comprehensive coverage
  projects: [
    {
      name: 'smoke',
      testMatch: '**/smoke/**/*.spec.ts',
      use: { ...devices['Desktop Chrome'] },
      timeout: 15000,
    },
    {
      name: 'quick',
      testMatch: '**/authentication.spec.ts',
      use: { ...devices['Desktop Chrome'] },
      timeout: 15000,
    },
    {
      name: 'full',
      testMatch: '**/*.spec.ts',
      use: { ...devices['Desktop Chrome'] },
      timeout: 30000,
    },
  ],

  // Global setup/teardown (if needed)
  // globalSetup: require.resolve('./tests/e2e/global-setup.ts'),
  // globalTeardown: require.resolve('./tests/e2e/global-teardown.ts'),
});
