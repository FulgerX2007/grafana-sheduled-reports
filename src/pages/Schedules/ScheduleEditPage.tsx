import React, { useState, useEffect } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2, Button, Field, Input, Select, Switch, TextArea, Form, FieldSet } from '@grafana/ui';
import { ScheduleFormData } from '../../types/types';
import { getBackendSrv } from '@grafana/runtime';
import { DashboardPicker } from '../../components/DashboardPicker';
import { CronEditor } from '../../components/CronEditor';
import { RecipientsEditor } from '../../components/RecipientsEditor';
import { VariablesEditor } from '../../components/VariablesEditor';

interface ScheduleEditPageProps {
  onNavigate: (page: string) => void;
  isNew: boolean;
  scheduleId?: number | null;
}

const intervalOptions = [
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
  { label: 'Custom (Cron)', value: 'cron' },
];

const formatOptions = [
  { label: 'PDF', value: 'pdf' },
  { label: 'HTML', value: 'html' },
];

export const ScheduleEditPage: React.FC<ScheduleEditPageProps> = ({ onNavigate, isNew, scheduleId }) => {
  const styles = useStyles2(getStyles);

  const [formData, setFormData] = useState<ScheduleFormData>({
    name: '',
    dashboard_uid: '',
    range_from: 'now-7d',
    range_to: 'now',
    interval_type: 'daily',
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    format: 'pdf',
    recipients: { to: [] },
    email_subject: 'Grafana Report: {{dashboard.title}}',
    email_body: 'Please find attached the dashboard report for {{timerange}}.',
    enabled: true,
  });

  useEffect(() => {
    if (!isNew && scheduleId) {
      loadSchedule();
    }
  }, [scheduleId]);

  const loadSchedule = async () => {
    try {
      const response = await getBackendSrv().get(`/api/plugins/grafana-app-reporting/resources/api/schedules/${scheduleId}`);
      setFormData(response);
    } catch (error) {
      console.error('Failed to load schedule:', error);
    }
  };

  const handleSubmit = async () => {
    try {
      if (isNew) {
        await getBackendSrv().post('/api/plugins/grafana-app-reporting/resources/api/schedules', formData);
      } else {
        await getBackendSrv().put(`/api/plugins/grafana-app-reporting/resources/api/schedules/${scheduleId}`, formData);
      }
      onNavigate('schedules');
    } catch (error) {
      console.error('Failed to save schedule:', error);
      alert('Failed to save schedule');
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>{isNew ? 'New Schedule' : 'Edit Schedule'}</h2>
      </div>

      <Form onSubmit={handleSubmit}>
        {() => (
          <>
            <FieldSet label="Basic Information">
              <Field label="Name" required>
                <Input
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.currentTarget.value })}
                  placeholder="Weekly Sales Report"
                />
              </Field>

              <Field label="Dashboard" required>
                <DashboardPicker
                  value={formData.dashboard_uid}
                  onChange={(uid) => setFormData({ ...formData, dashboard_uid: uid })}
                />
              </Field>

              <Field label="Format">
                <Select
                  options={formatOptions}
                  value={formData.format}
                  onChange={(v) => setFormData({ ...formData, format: v.value as 'pdf' | 'html' })}
                />
              </Field>

              <Field label="Enabled">
                <Switch
                  value={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.currentTarget.checked })}
                />
              </Field>
            </FieldSet>

            <FieldSet label="Time Range">
              <Field label="From">
                <Input
                  value={formData.range_from}
                  onChange={(e) => setFormData({ ...formData, range_from: e.currentTarget.value })}
                  placeholder="now-7d"
                />
              </Field>

              <Field label="To">
                <Input
                  value={formData.range_to}
                  onChange={(e) => setFormData({ ...formData, range_to: e.currentTarget.value })}
                  placeholder="now"
                />
              </Field>
            </FieldSet>

            <FieldSet label="Schedule">
              <Field label="Interval">
                <Select
                  options={intervalOptions}
                  value={formData.interval_type}
                  onChange={(v) => setFormData({ ...formData, interval_type: v.value as any })}
                />
              </Field>

              {formData.interval_type === 'cron' && (
                <Field label="Cron Expression">
                  <CronEditor
                    value={formData.cron_expr || ''}
                    onChange={(expr) => setFormData({ ...formData, cron_expr: expr })}
                  />
                </Field>
              )}

              <Field label="Timezone">
                <Input
                  value={formData.timezone}
                  onChange={(e) => setFormData({ ...formData, timezone: e.currentTarget.value })}
                />
              </Field>
            </FieldSet>

            <FieldSet label="Dashboard Variables">
              <VariablesEditor
                value={formData.variables || {}}
                onChange={(vars) => setFormData({ ...formData, variables: vars })}
              />
            </FieldSet>

            <FieldSet label="Email">
              <Field label="Recipients" required>
                <RecipientsEditor
                  value={formData.recipients}
                  onChange={(recipients) => setFormData({ ...formData, recipients })}
                />
              </Field>

              <Field label="Subject">
                <Input
                  value={formData.email_subject}
                  onChange={(e) => setFormData({ ...formData, email_subject: e.currentTarget.value })}
                />
              </Field>

              <Field label="Body">
                <TextArea
                  value={formData.email_body}
                  onChange={(e) => setFormData({ ...formData, email_body: e.currentTarget.value })}
                  rows={5}
                />
              </Field>
            </FieldSet>

            <div className={styles.actions}>
              {/* @ts-ignore */}
              <Button type="submit" variant="primary">
                {isNew ? 'Create' : 'Save'}
              </Button>
              {/* @ts-ignore */}
              <Button variant="secondary" onClick={() => onNavigate('schedules')}>
                Cancel
              </Button>
            </div>
          </>
        )}
      </Form>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    padding: ${theme.spacing(2)};
    max-width: 1200px;
  `,
  header: css`
    margin-bottom: ${theme.spacing(3)};
  `,
  actions: css`
    display: flex;
    gap: ${theme.spacing(2)};
    margin-top: ${theme.spacing(3)};
  `,
});
