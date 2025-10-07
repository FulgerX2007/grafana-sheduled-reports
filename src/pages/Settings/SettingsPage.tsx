import React, { useState, useEffect } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2, Button, Field, Input, Switch, FieldSet, Form } from '@grafana/ui';
import { Settings, SMTPConfig, RendererConfig, Limits } from '../../types/types';
import { getBackendSrv, getAppEvents } from '@grafana/runtime';
import { AppEvents } from '@grafana/data';

interface SettingsPageProps {
  onNavigate: (page: string) => void;
}

export const SettingsPage: React.FC<SettingsPageProps> = ({ onNavigate }) => {
  const styles = useStyles2(getStyles);
  const [settings, setSettings] = useState<Partial<Settings>>({
    use_grafana_smtp: true,
    renderer_config: {
      url: 'http://renderer:8081/render',
      timeout_ms: 60000,
      delay_ms: 1000,
      viewport_width: 1920,
      viewport_height: 1080,
    },
    limits: {
      max_recipients: 50,
      max_attachment_size_mb: 25,
      max_concurrent_renders: 5,
      retention_days: 30,
    },
  });

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const response = await getBackendSrv().get('/api/plugins/sheduled-reports-app/resources/api/settings');
      if (response) {
        setSettings(response);
      }
    } catch (error) {
      console.error('Failed to load settings:', error);
    }
  };

  const handleSubmit = async () => {
    const appEvents = getAppEvents();
    try {
      await getBackendSrv().post('/api/plugins/sheduled-reports-app/resources/api/settings', settings);
      appEvents.publish({
        type: AppEvents.alertSuccess.name,
        payload: ['Settings saved successfully'],
      });
    } catch (error) {
      console.error('Failed to save settings:', error);
      appEvents.publish({
        type: AppEvents.alertError.name,
        payload: ['Failed to save settings'],
      });
    }
  };

  const updateSMTP = (field: keyof SMTPConfig, value: any) => {
    setSettings({
      ...settings,
      smtp_config: { ...settings.smtp_config!, [field]: value },
    });
  };

  const updateRenderer = (field: keyof RendererConfig, value: any) => {
    setSettings({
      ...settings,
      renderer_config: { ...settings.renderer_config!, [field]: value },
    });
  };

  const updateLimits = (field: keyof Limits, value: any) => {
    setSettings({
      ...settings,
      limits: { ...settings.limits!, [field]: value },
    });
  };

  return (
    <div className={styles.container}>
      <h2>Settings</h2>

      <Form onSubmit={handleSubmit}>
        {() => (
          <>
            <FieldSet label="SMTP Configuration">
              <Field label="Use Grafana SMTP">
                <Switch
                  value={settings.use_grafana_smtp}
                  onChange={(e) => setSettings({ ...settings, use_grafana_smtp: e.currentTarget.checked })}
                />
              </Field>

              {!settings.use_grafana_smtp && (
                <>
                  <Field label="Host">
                    <Input
                      value={settings.smtp_config?.host || ''}
                      onChange={(e) => updateSMTP('host', e.currentTarget.value)}
                      placeholder="smtp.gmail.com"
                    />
                  </Field>
                  <Field label="Port">
                    <Input
                      type="number"
                      value={settings.smtp_config?.port || 587}
                      onChange={(e) => updateSMTP('port', parseInt(e.currentTarget.value))}
                    />
                  </Field>
                  <Field label="Username">
                    <Input
                      value={settings.smtp_config?.username || ''}
                      onChange={(e) => updateSMTP('username', e.currentTarget.value)}
                    />
                  </Field>
                  <Field label="Password">
                    <Input
                      type="password"
                      value={settings.smtp_config?.password || ''}
                      onChange={(e) => updateSMTP('password', e.currentTarget.value)}
                    />
                  </Field>
                  <Field label="From Address">
                    <Input
                      value={settings.smtp_config?.from || ''}
                      onChange={(e) => updateSMTP('from', e.currentTarget.value)}
                      placeholder="noreply@example.com"
                    />
                  </Field>
                  <Field label="Use TLS">
                    <Switch
                      value={settings.smtp_config?.use_tls}
                      onChange={(e) => updateSMTP('use_tls', e.currentTarget.checked)}
                    />
                  </Field>
                </>
              )}
            </FieldSet>

            <FieldSet label="Renderer Configuration">
              <Field label="Grafana URL" description="Base URL of your Grafana instance (leave empty to auto-detect)">
                <Input
                  value={settings.renderer_config?.url || ''}
                  onChange={(e) => updateRenderer('url', e.currentTarget.value)}
                  placeholder="http://localhost:3000"
                />
              </Field>
              <Field label="Timeout (ms)">
                <Input
                  type="number"
                  value={settings.renderer_config?.timeout_ms || 60000}
                  onChange={(e) => updateRenderer('timeout_ms', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Render Delay (ms)">
                <Input
                  type="number"
                  value={settings.renderer_config?.delay_ms || 1000}
                  onChange={(e) => updateRenderer('delay_ms', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Viewport Width">
                <Input
                  type="number"
                  value={settings.renderer_config?.viewport_width || 1920}
                  onChange={(e) => updateRenderer('viewport_width', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Viewport Height">
                <Input
                  type="number"
                  value={settings.renderer_config?.viewport_height || 1080}
                  onChange={(e) => updateRenderer('viewport_height', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Skip TLS Verification" description="Disable TLS certificate verification (use for self-signed certificates)">
                <Switch
                  value={settings.renderer_config?.skip_tls_verify || false}
                  onChange={(e) => updateRenderer('skip_tls_verify', e.currentTarget.checked)}
                />
              </Field>
            </FieldSet>

            <FieldSet label="Limits">
              <Field label="Max Recipients">
                <Input
                  type="number"
                  value={settings.limits?.max_recipients || 50}
                  onChange={(e) => updateLimits('max_recipients', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Max Attachment Size (MB)">
                <Input
                  type="number"
                  value={settings.limits?.max_attachment_size_mb || 25}
                  onChange={(e) => updateLimits('max_attachment_size_mb', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Max Concurrent Renders">
                <Input
                  type="number"
                  value={settings.limits?.max_concurrent_renders || 5}
                  onChange={(e) => updateLimits('max_concurrent_renders', parseInt(e.currentTarget.value))}
                />
              </Field>
              <Field label="Retention Days">
                <Input
                  type="number"
                  value={settings.limits?.retention_days || 30}
                  onChange={(e) => updateLimits('retention_days', parseInt(e.currentTarget.value))}
                />
              </Field>
            </FieldSet>

            {/* @ts-ignore */}
            <Button type="submit" variant="primary">
              Save Settings
            </Button>
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
});
