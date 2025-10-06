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
    await page.waitForTimeout(2000);

    // Take a screenshot for debugging
    await page.screenshot({ path: 'test-results/app-page.png', fullPage: true });

    // Verify page loaded - look for any content
    const pageContent = await page.locator('body').textContent();
    expect(pageContent).toBeTruthy();

    // Check if either the schedules page or error message is shown
    const hasSchedulesHeader = await page.locator('h2:has-text("Report Schedules")').isVisible().catch(() => false);
    const hasNewScheduleButton = await page.locator('button:has-text("New Schedule")').isVisible().catch(() => false);
    const hasSchedulesTab = await page.locator('[role="tab"]:has-text("Schedules")').isVisible().catch(() => false);

    // At least one of these should be visible
    const pageLoaded = hasSchedulesHeader || hasNewScheduleButton || hasSchedulesTab;
    expect(pageLoaded).toBeTruthy();
  });

  test('should have settings tab', async ({ authenticatedPage: page }) => {
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Look for Settings tab
    const settingsTab = page.locator('[role="tab"]:has-text("Settings")');
    const hasSettingsTab = await settingsTab.isVisible().catch(() => false);

    expect(hasSettingsTab).toBeTruthy();
  });
});
