import { defineConfig, devices } from '@playwright/test';
import { BASE_URL, PORT } from './lib/constants';
import './lib/db'; // loads ../.env and ./.env

export default defineConfig({
  testDir: './tests',
  // The suite shares one test database and seeded rows, so tests run one at a time.
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: 0,
  timeout: 30_000,
  expect: { timeout: 10_000 },
  reporter: [['list'], ['html', { open: 'never' }]],
  globalSetup: './global-setup.ts',
  use: {
    baseURL: BASE_URL,
    locale: 'cs-CZ',
    actionTimeout: 10_000,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'setup', testMatch: /.*\.setup\.ts/ },
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } },
      dependencies: ['setup'],
    },
  ],
  webServer: {
    // A built binary (not `go run`) so Playwright stops the actual server process when it is done.
    command: 'go build -o tmp/e2e-server . && ./tmp/e2e-server',
    cwd: '..',
    url: `${BASE_URL}/prihlaseni`,
    reuseExistingServer: false,
    timeout: 180_000,
    stdout: 'ignore',
    stderr: process.env.E2E_SERVER_LOGS === '1' ? 'pipe' : 'ignore',
    // godotenv does not override variables that are already set, so these win over ../.env:
    // emails are swallowed by the mock email service and the IMAP payment checker stays off.
    env: { APP_ENV: 'localhost', APP_PORT: String(PORT), PAYMENTS_CRON: '0' },
  },
});
