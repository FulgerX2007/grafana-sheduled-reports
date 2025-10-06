import { test as base } from '@playwright/test';

type AuthFixtures = {
  authenticatedPage: any;
};

export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // Login to Grafana
    await page.goto('/login');
    await page.fill('input[name="user"]', process.env.GF_ADMIN_USER || 'admin');
    await page.fill('input[name="password"]', process.env.GF_ADMIN_PASSWORD || 'admin');
    await page.click('button[type="submit"]');

    // Wait for navigation to complete
    await page.waitForURL('**/');

    // Skip welcome modal if present
    const skipButton = page.locator('button:has-text("Skip")');
    if (await skipButton.isVisible()) {
      await skipButton.click();
    }

    await use(page);
  },
});

export { expect } from '@playwright/test';
