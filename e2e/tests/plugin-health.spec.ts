import { test, expect } from '../fixtures/auth';

test.describe('Plugin Health Checks', () => {
  test('should load plugin successfully', async ({ authenticatedPage: page }) => {
    // Navigate to plugins page
    await page.goto('/plugins');
    await page.waitForLoadState('networkidle');

    // Search for the plugin
    const searchInput = page.locator('input[placeholder*="Search"]').first();
    await searchInput.fill('Scheduled Reports');
    await page.waitForTimeout(1000);

    // Click on the plugin
    await page.click('text=Scheduled Reports');
    await page.waitForLoadState('networkidle');

    // Verify plugin page loads
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Scheduled Reports' }).first()).toBeVisible({ timeout: 10000 });
  });

  test('should enable plugin', async ({ authenticatedPage: page }) => {
    // Navigate to plugin page
    await page.goto('/plugins/sheduled-reports-app');
    await page.waitForLoadState('networkidle');

    // Check if already enabled
    const enableButton = page.locator('button:has-text("Enable")');
    const configButton = page.locator('button:has-text("Configuration")');
    const isEnabled = await configButton.isVisible().catch(() => false);

    if (!isEnabled && await enableButton.isVisible()) {
      await enableButton.click();
      await page.waitForTimeout(2000);
    }
  });

  test('should navigate to plugin app', async ({ authenticatedPage: page }) => {
    // Navigate to the app
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Verify the Schedules page loads - check for header
    await expect(page.locator('h2:has-text("Report Schedules")')).toBeVisible({ timeout: 10000 });
  });

  test('should display settings page', async ({ authenticatedPage: page }) => {
    // Navigate to settings
    await page.goto('/a/sheduled-reports-app/settings');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Verify settings page loads - look for Settings tab
    await expect(page.locator('[role="tab"]:has-text("Settings")')).toBeVisible({ timeout: 10000 });
  });
});
