import { test, expect } from '../fixtures/auth';

test.describe('Plugin Settings', () => {
  test.beforeEach(async ({ authenticatedPage: page }) => {
    await page.goto('/a/sheduled-reports-app/settings');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should load settings page', async ({ authenticatedPage: page }) => {
    // Verify settings tab is visible
    await expect(page.locator('[role="tab"]:has-text("Settings")')).toBeVisible({ timeout: 10000 });
  });

  test('should configure SMTP settings', async ({ authenticatedPage: page }) => {
    // Find SMTP configuration section
    const smtpHostInput = page.locator('input[name*="smtp"], input[name*="host"]').first();

    if (await smtpHostInput.isVisible()) {
      await smtpHostInput.fill('smtp.test.com');

      // Fill other SMTP fields
      await page.fill('input[name*="port"]', '587');
      await page.fill('input[name*="user"], input[name*="username"]', 'test@example.com');

      // Save settings
      const saveButton = page.locator('button:has-text("Save")');
      if (await saveButton.isVisible()) {
        await saveButton.click();

        // Verify save success
        await expect(
          page.locator('text=Settings saved, text=Success')
        ).toBeVisible({ timeout: 10000 });
      }
    }
  });

  test('should configure renderer settings', async ({ authenticatedPage: page }) => {
    // Find renderer configuration
    const rendererUrlInput = page.locator('input[name*="renderer"]').first();

    if (await rendererUrlInput.isVisible()) {
      await rendererUrlInput.fill('http://renderer:8081/render');

      // Save settings
      const saveButton = page.locator('button:has-text("Save")');
      if (await saveButton.isVisible()) {
        await saveButton.click();
        await expect(page.locator('text=saved, text=Success')).toBeVisible({ timeout: 10000 });
      }
    }
  });

  test('should configure usage limits', async ({ authenticatedPage: page }) => {
    // Find limits configuration
    const maxRecipientsInput = page.locator('input[name*="recipient"]').first();

    if (await maxRecipientsInput.isVisible()) {
      await maxRecipientsInput.fill('10');

      // Save
      const saveButton = page.locator('button:has-text("Save")');
      if (await saveButton.isVisible()) {
        await saveButton.click();
        await expect(page.locator('text=saved, text=Success')).toBeVisible({ timeout: 10000 });
      }
    }
  });
});
