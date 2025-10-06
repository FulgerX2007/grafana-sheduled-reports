import { test, expect } from '../fixtures/auth';

/**
 * Smoke tests - Basic functionality checks
 * These tests verify the plugin loads and basic navigation works
 */
test.describe('Smoke Tests', () => {
  test('should load Grafana successfully', async ({ authenticatedPage: page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Verify we're logged in and on Grafana
    const body = await page.locator('body').textContent();
    expect(body).toBeTruthy();
  });

  test('should navigate to app plugin page', async ({ authenticatedPage: page }) => {
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Verify the Schedules header is visible
    await expect(page.locator('h2:has-text("Report Schedules")')).toBeVisible({ timeout: 10000 });
  });

  test('should have settings tab', async ({ authenticatedPage: page }) => {
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Look for Settings tab - using more flexible selector
    const settingsTab = page.locator('button:has-text("Settings"), a:has-text("Settings"), [role="tab"]:has-text("Settings")').first();
    await expect(settingsTab).toBeVisible({ timeout: 10000 });
  });
});
