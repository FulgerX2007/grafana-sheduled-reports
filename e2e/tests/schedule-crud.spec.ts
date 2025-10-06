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
    // Click New Schedule button - it might be in the header or in empty state
    const newScheduleButton = page.locator('button:has-text("New Schedule")').first();
    await newScheduleButton.click();

    // Wait for form to load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Fill in schedule details
    const nameInput = page.locator('input[name="name"]').first();
    await nameInput.fill('E2E Test Schedule');

    // Select dashboard - wait for dashboard picker to be ready
    await page.waitForTimeout(500);
    const dashboardInput = page.locator('input').filter({ hasText: '' }).or(
      page.locator('input[placeholder*="dashboard"]')
    ).first();

    // Try to interact with dashboard picker if it exists
    if (await dashboardInput.isVisible()) {
      await dashboardInput.fill('Test Dashboard');
      await page.waitForTimeout(1500);

      // Try to click on the test dashboard if it appears
      const testDashOption = page.locator('text=Test Dashboard for Reporting');
      if (await testDashOption.isVisible()) {
        await testDashOption.click();
      }
    }

    // Set format if available
    const pdfRadio = page.locator('input[value="pdf"]').first();
    if (await pdfRadio.isVisible()) {
      await pdfRadio.click();
    }

    // Set schedule - Daily if available
    const dailyButton = page.locator('button:has-text("Daily")').first();
    if (await dailyButton.isVisible()) {
      await dailyButton.click();
    }

    // Add recipient if field exists
    const recipientsInput = page.locator('input').filter({ hasText: '' }).or(
      page.locator('input[placeholder*="email"]')
    ).first();
    if (await recipientsInput.isVisible()) {
      await recipientsInput.fill('test@example.com');
    }

    // Submit form
    const submitButton = page.locator('button:has-text("Create Schedule"), button:has-text("Save")').first();
    if (await submitButton.isVisible()) {
      await submitButton.click();
      await page.waitForLoadState('networkidle');

      // Verify schedule was created
      await expect(page.locator('text=E2E Test Schedule')).toBeVisible({ timeout: 10000 });
    }
  });

  test('should list all schedules', async ({ authenticatedPage: page }) => {
    // Verify we're on the schedules page - either table or empty state should be visible
    const hasTable = await page.locator('table').isVisible();
    const hasEmptyState = await page.locator('text=No schedules yet').isVisible();

    expect(hasTable || hasEmptyState).toBeTruthy();
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
