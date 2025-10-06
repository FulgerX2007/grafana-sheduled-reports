import React from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';

export const DocumentationPage: React.FC = () => {
  const styles = useStyles2(getStyles);

  return (
    <div className={styles.container}>
      <h1>Grafana Reporting Documentation</h1>

      <section className={styles.section}>
        <h2>Overview</h2>
        <p>
          The Grafana Reporting plugin allows you to schedule automatic generation and delivery of dashboard
          reports via email. Reports can be generated as PDF or HTML and sent on a daily, weekly, monthly,
          or custom cron schedule.
        </p>
      </section>

      <section className={styles.section}>
        <h2>Getting Started</h2>

        <h3>1. Configure Plugin Settings</h3>
        <p>Before creating schedules, configure the plugin in the Settings page:</p>
        <ul>
          <li><strong>Service Account Token:</strong> Required for authenticating dashboard rendering</li>
          <li><strong>SMTP Settings:</strong> Configure email delivery (or use Grafana's SMTP configuration)</li>
          <li><strong>Renderer Settings:</strong> Configure the Grafana Image Renderer service</li>
        </ul>

        <h3>2. Create a Schedule</h3>
        <ol>
          <li>Click "New Schedule" button</li>
          <li>Fill in the schedule details (see below)</li>
          <li>Click "Create" to save</li>
        </ol>
      </section>

      <section className={styles.section}>
        <h2>Schedule Configuration</h2>

        <h3>Basic Information</h3>
        <ul>
          <li><strong>Name:</strong> A descriptive name for your schedule (e.g., "Daily Sales Report")</li>
          <li><strong>Dashboard:</strong> Select the dashboard to report</li>
          <li><strong>Format:</strong> Choose PDF or HTML output</li>
          <li><strong>Enabled:</strong> Enable or disable the schedule</li>
        </ul>

        <h3>Time Range</h3>
        <ul>
          <li><strong>From:</strong> Start of the time range (e.g., "now-24h", "now-7d", "2024-01-01")</li>
          <li><strong>To:</strong> End of the time range (e.g., "now")</li>
        </ul>

        <h3>Schedule Intervals</h3>
        <ul>
          <li><strong>Daily:</strong> Runs once per day at the time the schedule was created</li>
          <li><strong>Weekly:</strong> Runs once per week</li>
          <li><strong>Monthly:</strong> Runs once per month</li>
          <li><strong>Custom (Cron):</strong> Use cron expressions for precise scheduling</li>
        </ul>

        <h3>Cron Expression Format</h3>
        <p>Cron expressions use 5 fields: <code>minute hour day-of-month month day-of-week</code></p>

        <div className={styles.codeBlock}>
          <table>
            <thead>
              <tr>
                <th>Expression</th>
                <th>Description</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><code>0 8 * * *</code></td>
                <td>Every day at 8:00 AM</td>
              </tr>
              <tr>
                <td><code>0 9 * * 1</code></td>
                <td>Every Monday at 9:00 AM</td>
              </tr>
              <tr>
                <td><code>0 18 * * 1-5</code></td>
                <td>Every weekday at 6:00 PM</td>
              </tr>
              <tr>
                <td><code>0 0 1 * *</code></td>
                <td>First day of every month at midnight</td>
              </tr>
              <tr>
                <td><code>*/15 * * * *</code></td>
                <td>Every 15 minutes</td>
              </tr>
              <tr>
                <td><code>0 8 * * 0</code></td>
                <td>Every Sunday at 8:00 AM</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3>Dashboard Variables</h3>
        <p>
          You can set dashboard variables that will be applied when rendering the report.
          For example, if your dashboard has a "datacenter" variable, you can set it to a specific value.
        </p>

        <h3>Email Configuration</h3>
        <ul>
          <li><strong>Recipients:</strong> Comma-separated email addresses (To, CC, BCC)</li>
          <li><strong>Subject:</strong> Email subject line (supports template variables)</li>
          <li><strong>Body:</strong> Email body text (supports template variables)</li>
        </ul>

        <h4>Template Variables</h4>
        <p>You can use the following variables in email subject and body:</p>
        <ul>
          <li><code>{'{{schedule.name}}'}</code> - Name of the schedule</li>
          <li><code>{'{{dashboard.title}}'}</code> - Dashboard title</li>
          <li><code>{'{{timerange}}'}</code> - Time range used for the report</li>
          <li><code>{'{{run.started_at}}'}</code> - When the report generation started</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Managing Schedules</h2>

        <h3>Schedule Actions</h3>
        <ul>
          <li><strong>▶ Run Now:</strong> Execute the schedule immediately</li>
          <li><strong>⏸ Pause/Resume:</strong> Disable or enable the schedule</li>
          <li><strong>✏️ Edit:</strong> Modify schedule configuration</li>
          <li><strong>🕐 History:</strong> View past report runs and results</li>
          <li><strong>🗑️ Delete:</strong> Remove the schedule permanently</li>
        </ul>

        <h3>Run History</h3>
        <p>
          The Run History page shows all past executions of a schedule, including:
        </p>
        <ul>
          <li>Execution time and duration</li>
          <li>Status (completed, failed, running)</li>
          <li>Number of pages rendered</li>
          <li>File size</li>
          <li>Error messages (if failed)</li>
          <li>Download button for successful reports</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Reporting Info Panel</h2>
        <p>
          You can add a "Reporting Info" panel to any dashboard to display reporting schedules
          for that dashboard. This provides visibility into which reports are configured.
        </p>

        <h3>Adding the Panel</h3>
        <ol>
          <li>Edit your dashboard</li>
          <li>Add a new panel</li>
          <li>Select "Reporting Info" from the visualization list</li>
          <li>Resize and position the panel as needed</li>
          <li>Save the dashboard</li>
        </ol>

        <p>
          The panel will automatically show:
        </p>
        <ul>
          <li>Number of active scheduled reports</li>
          <li>List of all schedules for the dashboard</li>
          <li>Schedule intervals and next run times</li>
          <li>Link to manage reports</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Settings</h2>

        <h3>SMTP Configuration</h3>
        <p>
          You can either use Grafana's SMTP settings or configure custom SMTP settings for the plugin:
        </p>
        <ul>
          <li><strong>Use Grafana SMTP:</strong> Uses environment variables from Grafana configuration</li>
          <li><strong>Custom SMTP:</strong> Configure separate SMTP settings for reporting</li>
        </ul>

        <h3>Renderer Configuration</h3>
        <ul>
          <li><strong>URL:</strong> Grafana Image Renderer service URL (default: http://renderer:8081/render)</li>
          <li><strong>Timeout:</strong> Maximum time to wait for rendering (milliseconds)</li>
          <li><strong>Delay:</strong> Wait time before capturing to allow queries to complete</li>
          <li><strong>Viewport:</strong> Browser viewport dimensions for rendering</li>
        </ul>

        <h3>Limits</h3>
        <ul>
          <li><strong>Max Recipients:</strong> Maximum number of email recipients per schedule</li>
          <li><strong>Max Attachment Size:</strong> Maximum report file size in MB</li>
          <li><strong>Max Concurrent Renders:</strong> Number of reports that can render simultaneously</li>
          <li><strong>Retention Days:</strong> How long to keep report artifacts</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Troubleshooting</h2>

        <h3>Reports Not Being Generated</h3>
        <ul>
          <li>Check that the schedule is enabled</li>
          <li>Verify the "Next Run" time is in the future</li>
          <li>Check the Run History for error messages</li>
          <li>Verify the service account token is configured</li>
        </ul>

        <h3>Rendering Errors</h3>
        <ul>
          <li>Ensure Grafana Image Renderer service is running</li>
          <li>Check renderer URL in settings</li>
          <li>Increase timeout if dashboards have slow queries</li>
          <li>Verify service account has access to the dashboard</li>
        </ul>

        <h3>Email Delivery Issues</h3>
        <ul>
          <li>Verify SMTP configuration is correct</li>
          <li>Check email addresses are valid</li>
          <li>Look for error messages in Run History</li>
          <li>Test SMTP settings with "Test Email" button (if available)</li>
        </ul>

        <h3>Dashboard Variables Not Working</h3>
        <ul>
          <li>Ensure variable names match exactly (case-sensitive)</li>
          <li>Check that variables are defined in the dashboard</li>
          <li>Use the format <code>var-variableName</code> in the configuration</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Best Practices</h2>

        <ul>
          <li><strong>Use descriptive names:</strong> Make it easy to identify schedules at a glance</li>
          <li><strong>Set appropriate time ranges:</strong> Match the schedule interval (e.g., "now-24h" for daily)</li>
          <li><strong>Test before scheduling:</strong> Use "Run Now" to verify reports look correct</li>
          <li><strong>Monitor Run History:</strong> Check for failed runs regularly</li>
          <li><strong>Use template variables:</strong> Make email content dynamic and informative</li>
          <li><strong>Set retention policies:</strong> Clean up old reports to save storage</li>
          <li><strong>Limit recipients:</strong> Keep recipient lists focused to reduce email volume</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Technical Details</h2>

        <h3>Architecture</h3>
        <p>The reporting plugin consists of:</p>
        <ul>
          <li><strong>Frontend:</strong> React-based UI for managing schedules</li>
          <li><strong>Backend:</strong> Go service for scheduling and report generation</li>
          <li><strong>Database:</strong> SQLite for storing schedules and run history</li>
          <li><strong>Renderer:</strong> Grafana Image Renderer for capturing dashboards</li>
        </ul>

        <h3>Data Storage</h3>
        <ul>
          <li><strong>Database:</strong> /var/lib/grafana/plugin-data/reporting.db</li>
          <li><strong>Artifacts:</strong> /var/lib/grafana/plugin-data/artifacts/org_[id]/</li>
        </ul>

        <h3>Security</h3>
        <ul>
          <li>Service account tokens are stored securely in Grafana's encrypted settings</li>
          <li>All schedules are scoped by organization ID</li>
          <li>Users must have Editor role to create schedules</li>
          <li>Admin role required to modify plugin settings</li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2>Support</h2>
        <p>
          For additional help or to report issues, please contact your Grafana administrator
          or visit the plugin documentation repository.
        </p>
      </section>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    padding: ${theme.spacing(3)};
    max-width: 1200px;
    margin: 0 auto;

    h1 {
      margin-bottom: ${theme.spacing(3)};
      border-bottom: 2px solid ${theme.colors.border.medium};
      padding-bottom: ${theme.spacing(2)};
    }

    h2 {
      margin-top: ${theme.spacing(4)};
      margin-bottom: ${theme.spacing(2)};
      color: ${theme.colors.primary.text};
    }

    h3 {
      margin-top: ${theme.spacing(3)};
      margin-bottom: ${theme.spacing(1.5)};
      color: ${theme.colors.text.primary};
    }

    h4 {
      margin-top: ${theme.spacing(2)};
      margin-bottom: ${theme.spacing(1)};
    }

    p {
      margin-bottom: ${theme.spacing(2)};
      line-height: 1.6;
    }

    ul, ol {
      margin-bottom: ${theme.spacing(2)};
      padding-left: ${theme.spacing(3)};
      line-height: 1.8;
    }

    li {
      margin-bottom: ${theme.spacing(0.5)};
    }

    code {
      background: ${theme.colors.background.secondary};
      padding: ${theme.spacing(0.25)} ${theme.spacing(0.75)};
      border-radius: ${theme.shape.radius.default};
      font-family: ${theme.typography.fontFamilyMonospace};
      font-size: 0.9em;
      color: ${theme.colors.primary.text};
    }

    strong {
      font-weight: ${theme.typography.fontWeightMedium};
      color: ${theme.colors.text.primary};
    }

    table {
      width: 100%;
      border-collapse: collapse;
      margin: ${theme.spacing(2)} 0;
    }

    th {
      background: ${theme.colors.background.secondary};
      padding: ${theme.spacing(1)} ${theme.spacing(2)};
      text-align: left;
      border-bottom: 2px solid ${theme.colors.border.medium};
      font-weight: ${theme.typography.fontWeightMedium};
    }

    td {
      padding: ${theme.spacing(1)} ${theme.spacing(2)};
      border-bottom: 1px solid ${theme.colors.border.weak};
    }
  `,
  section: css`
    margin-bottom: ${theme.spacing(4)};
  `,
  codeBlock: css`
    background: ${theme.colors.background.secondary};
    border: 1px solid ${theme.colors.border.weak};
    border-radius: ${theme.shape.radius.default};
    padding: ${theme.spacing(2)};
    margin: ${theme.spacing(2)} 0;
    overflow-x: auto;
  `,
});
