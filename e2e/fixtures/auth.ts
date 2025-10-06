import { test as base, request } from '@playwright/test';

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
    if (await skipButton.isVisible().catch(() => false)) {
      await skipButton.click();
    }

    // Enable the plugin if not already enabled
    try {
      const response = await page.request.post('/api/plugins/sheduled-reports-app/settings', {
        data: {
          enabled: true,
          pinned: true,
          jsonData: {}
        }
      });

      if (response.ok()) {
        console.log('Plugin enabled successfully');
        // Wait a bit for plugin to initialize
        await page.waitForTimeout(2000);
      }
    } catch (error) {
      console.log('Plugin already enabled or error enabling:', error);
    }

    await use(page);
  },
});

export { expect } from '@playwright/test';
