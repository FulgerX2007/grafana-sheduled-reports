import React, { useState } from 'react';
import { AppPluginMeta, PluginConfigPageProps } from '@grafana/data';
import { Button, Field, FieldSet, Input, Legend, Alert } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';

interface AppConfigProps extends PluginConfigPageProps<AppPluginMeta> {}

export const AppConfig: React.FC<AppConfigProps> = ({ plugin }) => {
  const { enabled, jsonData, secureJsonFields } = plugin.meta;
  const [serviceAccountToken, setServiceAccountToken] = useState('');
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState('');

  const hasToken = secureJsonFields?.serviceAccountToken;

  const handleSaveToken = async () => {
    if (!serviceAccountToken && !hasToken) {
      setErrorMessage('Please enter a service account token');
      setSaveStatus('error');
      return;
    }

    setSaveStatus('saving');
    setErrorMessage('');

    try {
      // Update plugin settings via Grafana API
      await getBackendSrv().post(`/api/plugins/${plugin.meta.id}/settings`, {
        enabled: enabled,
        pinned: true,
        jsonData: jsonData || {},
        secureJsonData: serviceAccountToken ? { serviceAccountToken } : undefined,
      });

      setSaveStatus('success');
      setServiceAccountToken(''); // Clear input after successful save

      // Reset success message after 3 seconds
      setTimeout(() => setSaveStatus('idle'), 3000);
    } catch (error) {
      console.error('Failed to save token:', error);
      setErrorMessage('Failed to save service account token');
      setSaveStatus('error');
    }
  };

  const handleEnablePlugin = async () => {
    try {
      await getBackendSrv().post(`/api/plugins/${plugin.meta.id}/settings`, {
        enabled: true,
        pinned: true,
        jsonData: jsonData || {},
      });
      window.location.reload();
    } catch (error) {
      console.error('Failed to enable plugin:', error);
      setErrorMessage('Failed to enable plugin');
      setSaveStatus('error');
    }
  };

  return (
    <div>
      <FieldSet>
        <Legend>Plugin Configuration</Legend>
        <p>
          This plugin allows you to schedule and email dashboard reports as PDF or HTML files.
        </p>

        <Alert title="Plugin Status" severity={enabled ? 'success' : 'info'}>
          {enabled ? 'Plugin is enabled and ready to use' : 'Enable the plugin to start creating report schedules'}
        </Alert>

        {saveStatus === 'success' && (
          <Alert title="Settings Saved" severity="success">
            Service account token has been saved successfully
          </Alert>
        )}

        {saveStatus === 'error' && errorMessage && (
          <Alert title="Error" severity="error">
            {errorMessage}
          </Alert>
        )}

        <Field
          label="Service Account Token"
          description="Grafana service account token for rendering dashboards. This token is encrypted and stored securely."
        >
          <div style={{ display: 'flex', gap: '8px', alignItems: 'flex-start', flexDirection: 'column' }}>
            <Input
              type="password"
              placeholder={hasToken ? 'Token is configured (enter new token to update)' : 'Enter service account token'}
              value={serviceAccountToken}
              onChange={(e) => setServiceAccountToken(e.currentTarget.value)}
              width={60}
            />
            {hasToken && (
              <div style={{ color: '#52c41a', fontSize: '12px' }}>
                ✓ Service account token is configured
              </div>
            )}
            <div style={{ display: 'flex', gap: '8px' }}>
              {/* @ts-ignore */}
              <Button
                variant="primary"
                onClick={handleSaveToken}
                disabled={saveStatus === 'saving'}
              >
                {saveStatus === 'saving' ? 'Saving...' : 'Save Token'}
              </Button>
              {!enabled && (
                // @ts-ignore
                <Button variant="secondary" onClick={handleEnablePlugin}>
                  Enable Plugin
                </Button>
              )}
            </div>
          </div>
        </Field>

        <Field description="After enabling the plugin, configure SMTP and renderer settings in the Settings page">
          <div />
        </Field>
      </FieldSet>
    </div>
  );
};
