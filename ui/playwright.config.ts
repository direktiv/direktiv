import * as dotenv from "dotenv";

import { defineConfig, devices } from "@playwright/test";

import { storageState } from "./e2e/setup/auth";

dotenv.config();

const {
  PLAYWRIGHT_CI,
  PLAYWRIGHT_PARALLEL,
  PLAYWRIGHT_UI_BASE_URL,
  PLAYWRIGHT_USE_VITE,
} = process.env;

if (!PLAYWRIGHT_UI_BASE_URL) console.error("PLAYWRIGHT_UI_BASE_URL is not set");

const baseURL = PLAYWRIGHT_UI_BASE_URL
  ? new URL(`${PLAYWRIGHT_UI_BASE_URL}`)
  : undefined;

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: "./e2e",
  /* Run tests sequentially, since parallel runs are curently failing randomly */
  fullyParallel: false,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!PLAYWRIGHT_CI,
  /* Retry on CI only */
  retries: PLAYWRIGHT_CI ? 3 : 0,
  /* Opt out of parallel tests per default. Enable parallel tests by setting
     PLAYWRIGHT_PARALLEL to TRUE. */
  workers: PLAYWRIGHT_PARALLEL === "TRUE" ? undefined : 1,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: "html",
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: baseURL?.toString(),
    storageState,
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    // trace: "on-first-retry",
    /* temporary override: this is expensive, but useful for debugging tests */
    trace: "retain-on-failure",
  },

  timeout: 30000, // defaults to 30000
  expect: { timeout: 10000 },

  /* Configure projects for major browsers */
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },

    /* Test against mobile viewports. */
    // {
    //   name: 'Mobile Chrome',
    //   use: { ...devices['Pixel 5'] },
    // },
    // {
    //   name: 'Mobile Safari',
    //   use: { ...devices['iPhone 12'] },
    // },

    /* Test against branded browsers. */
    // {
    //   name: 'Microsoft Edge',
    //   use: { ...devices['Desktop Edge'], channel: 'msedge' },
    // },
    // {
    //   name: 'Google Chrome',
    //   use: { ..devices['Desktop Chrome'], channel: 'chrome' },
    // },
  ],

  /* Run your local dev server before starting the tests */
  webServer:
    PLAYWRIGHT_USE_VITE === "TRUE"
      ? {
          timeout: 60000,
          command: `pnpm exec vite ${baseURL && `--port ${baseURL.port}`}`,
          url: baseURL?.toString(),
          reuseExistingServer: !PLAYWRIGHT_CI,
        }
      : undefined,
});
