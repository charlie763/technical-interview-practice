import { defineConfig } from '@playwright/test'

export default defineConfig({
  globalSetup: './global-setup.js',
  testDir: './tests',
  timeout: 30_000,
  use: {
    baseURL: 'http://localhost:5173',
    headless: true,
  },
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI && !process.env.PRACTICE_ANSWER,
    timeout: 30_000,
  },
})
