# E2E Testing Guide

## Quick Start

```bash
# Build the plugin first
npm run build
go build -o dist/gpx_reporting ./cmd/backend

# Run all E2E tests
npm run test:e2e

# Run with UI (recommended for debugging)
npm run test:e2e:ui

# Run specific test file
npx playwright test e2e/tests/smoke.spec.ts

# Run in headed mode (see the browser)
npm run test:e2e:headed
```

## Test Files

- **smoke.spec.ts** - Basic smoke tests (plugin loads, navigation works)
- **plugin-health.spec.ts** - Plugin installation and configuration
- **schedule-crud.spec.ts** - Schedule creation, editing, deletion
- **report-execution.spec.ts** - Report generation and download
- **settings.spec.ts** - Settings configuration

## Common Issues

### Tests Failing: "Page not found" or "Navigation timeout"

**Cause**: Plugin ID mismatch or plugin not loaded

**Fix**:
1. Ensure `src/plugin.json` has `"id": "sheduled-reports-app"`
2. Rebuild frontend: `npm run build`
3. Rebuild backend: `go build -o dist/gpx_reporting ./cmd/backend`
4. Check Grafana logs: `docker compose -f docker-compose.e2e.yml logs grafana`

### Tests Failing: Element not found

**Cause**: React components not rendered or selectors incorrect

**Fix**:
1. Run test in UI mode: `npm run test:e2e:ui`
2. Inspect the page to see what's actually rendered
3. Update selectors in test files
4. Add more `waitForTimeout` if components are slow to load

### Plugin not appearing in Grafana

**Check**:
```bash
# Verify plugin files exist
ls -la dist/

# Should see:
# - module.js (or module.js.gz)
# - gpx_reporting (backend binary)

# Check Grafana recognizes the plugin
docker compose -f docker-compose.e2e.yml exec grafana ls -la /var/lib/grafana/plugins/sheduled-reports-app/
```

### Backend not starting

**Check**:
```bash
# View backend logs
docker compose -f docker-compose.e2e.yml logs grafana | grep gpx_reporting

# Common issues:
# - Binary not executable: chmod +x dist/gpx_reporting
# - Wrong architecture: rebuild with CGO_ENABLED=1
```

## Debugging Tips

### 1. Visual Debugging

```bash
# Run in headed mode to see the browser
npm run test:e2e:headed

# Run in debug mode with step-through
npm run test:e2e:debug
```

### 2. Screenshots and Videos

Failed tests automatically capture:
- Screenshots: `test-results/[test-name]/test-failed-1.png`
- Videos: `test-results/[test-name]/video.webm`

### 3. Check Grafana Logs

```bash
# Follow logs in real-time
docker compose -f docker-compose.e2e.yml logs -f grafana

# Check for plugin errors
docker compose -f docker-compose.e2e.yml logs grafana | grep -i error
```

### 4. Manual Testing

```bash
# Start the environment manually
docker compose -f docker-compose.e2e.yml up -d

# Access Grafana at http://localhost:3000
# Login: admin / admin
# Navigate to: Apps > Scheduled Reports

# Check MailHog for test emails
# http://localhost:8025
```

## Test Environment

### Services

- **Grafana**: http://localhost:3000 (admin/admin)
- **Image Renderer**: http://localhost:8081
- **MailHog SMTP**: http://localhost:8025

### Clean Start

```bash
# Stop and remove all data
docker compose -f docker-compose.e2e.yml down -v

# Start fresh
docker compose -f docker-compose.e2e.yml up -d

# Wait for health checks
npm run e2e:wait
```

## Writing New Tests

### Template

```typescript
import { test, expect } from '../fixtures/auth';

test.describe('My Feature', () => {
  test.beforeEach(async ({ authenticatedPage: page }) => {
    await page.goto('/a/sheduled-reports-app');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000); // Let React render
  });

  test('should do something', async ({ authenticatedPage: page }) => {
    // Your test code
    const button = page.locator('button:has-text("Click Me")');
    await button.click();

    // Verify result
    await expect(page.locator('text=Success')).toBeVisible();
  });
});
```

### Best Practices

1. **Always wait for network idle**: `await page.waitForLoadState('networkidle');`
2. **Add buffer time for React**: `await page.waitForTimeout(1000);`
3. **Use robust selectors**: `button:has-text("Save")` instead of `.btn-primary`
4. **Check visibility before interaction**: `if (await button.isVisible()) { ... }`
5. **Take screenshots for debugging**: `await page.screenshot({ path: 'debug.png' });`

## CI/CD

Tests run automatically on:
- Push to main/develop
- Pull requests
- Manual workflow dispatch

Matrix testing across Grafana versions:
- 10.0.0
- 11.0.0
- latest

## Troubleshooting Checklist

- [ ] Plugin built: `npm run build && go build -o dist/gpx_reporting ./cmd/backend`
- [ ] Docker services healthy: `docker compose -f docker-compose.e2e.yml ps`
- [ ] Grafana accessible: `curl http://localhost:3000/api/health`
- [ ] Plugin loaded: Check http://localhost:3000/plugins
- [ ] Backend running: Check Grafana logs for plugin startup
- [ ] Correct plugin ID: `sheduled-reports-app` everywhere
