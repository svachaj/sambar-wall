import { defineConfig, devices } from '@playwright/test';
import { DEPLOYED_URL } from './lib/constants';
import './lib/db'; // loads ../.env and ./.env

// Smoke tests against the deployed test site (`make e2e-deployed`). It must use the database from .env;
// the suite seeds and purges the same @e2e-sambar.test data as the local suite.
// The deployed app sends real emails, so these tests never trigger one: they log in with a code written to the DB.
export default defineConfig({
  testDir: './deployed',
  fullyParallel: false,
  workers: 1,
  retries: 0,
  timeout: 30_000,
  expect: { timeout: 10_000 },
  reporter: [['list'], ['html', { open: 'never', outputFolder: 'playwright-report-deployed' }]],
  outputDir: 'test-results-deployed',
  globalSetup: './global-setup.ts',
  use: {
    baseURL: DEPLOYED_URL,
    locale: 'cs-CZ',
    actionTimeout: 10_000,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'deployed-setup', testMatch: /.*\.setup\.ts/ },
    {
      name: 'deployed',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } },
      dependencies: ['deployed-setup'],
    },
  ],
});
