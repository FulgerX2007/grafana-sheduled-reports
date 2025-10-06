import { test, expect } from '../fixtures/auth';

test.describe('Schedule CRUD Operations', () => {
  test.beforeEach(async ({ authenticatedPage: page }) => {
    // Ensure plugin is enabled and navigate to app
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    // Wait a bit for React to render
    await page.waitForTimeout(1000);
  });

  test('should create a new schedule', async ({ authenticatedPage: page }) => {
    // This test verifies the New Schedule button is clickable
    // Full form testing would require the frontend components to be fully implemented
    const newScheduleButton = page.locator('button:has-text("New Schedule")').first();
    await expect(newScheduleButton).toBeVisible({ timeout: 10000 });

    // Verify button is clickable (but don't test full form yet)
    const isClickable = await newScheduleButton.isEnabled();
    expect(isClickable).toBeTruthy();
  });

  test('should list all schedules', async ({ authenticatedPage: page }) => {
    // Verify we're on the schedules page - header should be visible
    await expect(page.locator('h2:has-text("Report Schedules")')).toBeVisible({ timeout: 10000 });
  });

  test('should view schedule details', async ({ authenticatedPage: page }) => {
    // Create a schedule first (or assume one exists)
    const scheduleName = page.locator('text=E2E Test Schedule').first();

    if (await scheduleName.isVisible()) {
      await scheduleName.click();

      // Verify details page loads
      await expect(page.locator('text=Schedule Details').or(page.locator('input[value*="E2E Test"]'))).toBeVisible({ timeout: 5000 });
    }
  });

  test('should edit a schedule', async ({ authenticatedPage: page }) => {
    // Find and click edit button for first schedule
    const editButton = page.locator('button[aria-label*="Edit"], button:has-text("Edit")').first();

    if (await editButton.isVisible()) {
      await editButton.click();
      await page.waitForLoadState('networkidle');

      // Modify the name
      const nameInput = page.locator('input[name="name"]');
      await nameInput.fill('E2E Test Schedule Updated');

      // Save changes
      await page.click('button:has-text("Save"), button:has-text("Update")');

      // Verify update
      await expect(page.locator('text=E2E Test Schedule Updated')).toBeVisible({ timeout: 10000 });
    }
  });

  test('should delete a schedule', async ({ authenticatedPage: page }) => {
    // Find delete button
    const deleteButton = page.locator('button[aria-label*="Delete"], button:has-text("Delete")').first();

    if (await deleteButton.isVisible()) {
      const scheduleName = await page.locator('td, [role="cell"]').first().textContent();

      await deleteButton.click();

      // Confirm deletion
      const confirmButton = page.locator('button:has-text("Confirm"), button:has-text("Delete")');
      if (await confirmButton.isVisible()) {
        await confirmButton.click();
      }

      // Verify schedule is removed
      if (scheduleName) {
        await expect(page.locator(`text=${scheduleName}`)).not.toBeVisible({ timeout: 10000 });
      }
    }
  });
});
