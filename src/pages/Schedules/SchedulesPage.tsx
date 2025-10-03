import React, { useState, useEffect } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2, Button, Icon, LoadingPlaceholder } from '@grafana/ui';
import { Schedule } from '../../types/types';
import { getBackendSrv } from '@grafana/runtime';

interface SchedulesPageProps {
  onNavigate: (page: string, id?: number) => void;
}

export const SchedulesPage: React.FC<SchedulesPageProps> = ({ onNavigate }) => {
  const styles = useStyles2(getStyles);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSchedules();
  }, []);

  const loadSchedules = async () => {
    try {
      const response = await getBackendSrv().get('/api/plugins/grafana-app-reporting/resources/api/schedules');
      console.log('API Response:', response);

      // Handle different response formats
      if (Array.isArray(response)) {
        setSchedules(response);
      } else if (response && response.schedules && Array.isArray(response.schedules)) {
        setSchedules(response.schedules);
      } else if (response && response.schedules === null) {
        // Backend returned null instead of empty array
        setSchedules([]);
      } else {
        console.warn('Unexpected response format:', response);
        setSchedules([]);
      }
    } catch (error) {
      console.error('Failed to load schedules:', error);
      setSchedules([]);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this schedule?')) {
      return;
    }
    try {
      await getBackendSrv().delete(`/api/plugins/grafana-app-reporting/resources/api/schedules/${id}`);
      loadSchedules();
    } catch (error) {
      console.error('Failed to delete schedule:', error);
    }
  };

  const handleRunNow = async (id: number) => {
    try {
      await getBackendSrv().post(`/api/plugins/grafana-app-reporting/resources/api/schedules/${id}/run`);
      alert('Report generation started');
    } catch (error) {
      console.error('Failed to run schedule:', error);
    }
  };

  const handleToggle = async (schedule: Schedule) => {
    try {
      await getBackendSrv().put(`/api/plugins/grafana-app-reporting/resources/api/schedules/${schedule.id}`, {
        ...schedule,
        enabled: !schedule.enabled,
      });
      loadSchedules();
    } catch (error) {
      console.error('Failed to toggle schedule:', error);
    }
  };

  if (loading) {
    return <LoadingPlaceholder text="Loading schedules..." />;
  }

  // Ensure schedules is always an array
  const safeSchedules = Array.isArray(schedules) ? schedules : [];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>Report Schedules</h2>
          {/* @ts-ignore */}
        <Button
          icon="plus"
          onClick={() => onNavigate('schedule-new')}
        >
          New Schedule
        </Button>
      </div>

      {safeSchedules.length === 0 ? (
        <div className={styles.empty}>
          <Icon name="calendar-alt" size="xxxl" />
          <h3>No schedules yet</h3>
          <p>Create your first report schedule to get started</p>
          {/* @ts-ignore */}
          <Button
            icon="plus"
            onClick={() => onNavigate('schedule-new')}
          >
            Create Schedule
          </Button>
        </div>
      ) : (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Name</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Dashboard</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Schedule</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Format</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Status</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Next Run</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {safeSchedules.map((schedule) => (
              <tr key={schedule.id}>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{schedule.name}</td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                  {schedule.dashboard_title || schedule.dashboard_uid}
                </td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                  {schedule.interval_type === 'cron' ? schedule.cron_expr : schedule.interval_type}
                </td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{schedule.format.toUpperCase()}</td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                  <span className={schedule.enabled ? styles.statusEnabled : styles.statusDisabled}>
                    {schedule.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                  {schedule.next_run_at ? new Date(schedule.next_run_at).toLocaleString() : 'N/A'}
                </td>
                <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                  <div className={styles.actions}>
                    {/* @ts-ignore */}
                    <Button
                      size="sm"
                      variant="secondary"
                      icon="play"
                      onClick={() => handleRunNow(schedule.id)}
                      title="Run now"
                    />
                    {/* @ts-ignore */}
                    <Button
                      size="sm"
                      variant="secondary"
                      icon={schedule.enabled ? 'pause' : 'play'}
                      onClick={() => handleToggle(schedule)}
                      title={schedule.enabled ? 'Disable' : 'Enable'}
                    />
                    {/* @ts-ignore */}
                    <Button
                      size="sm"
                      variant="secondary"
                      icon="edit"
                      onClick={() => onNavigate('schedule-edit', schedule.id)}
                      title="Edit"
                    />
                    {/* @ts-ignore */}
                    <Button
                      size="sm"
                      variant="secondary"
                      icon="history"
                      onClick={() => onNavigate('run-history', schedule.id)}
                      title="View history"
                    />
                    {/* @ts-ignore */}
                    <Button
                      size="sm"
                      variant="destructive"
                      icon="trash-alt"
                      onClick={() => handleDelete(schedule.id)}
                      title="Delete"
                    />
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  container: css`
    padding: ${theme.spacing(2)};
  `,
  header: css`
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: ${theme.spacing(3)};
  `,
  actions: css`
    display: flex;
    gap: ${theme.spacing(1)};
  `,
  statusEnabled: css`
    color: ${theme.colors.success.text};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  statusDisabled: css`
    color: ${theme.colors.text.secondary};
  `,
  empty: css`
    text-align: center;
    padding: ${theme.spacing(8)};
    color: ${theme.colors.text.secondary};

    h3 {
      margin: ${theme.spacing(2)} 0;
    }

    p {
      margin-bottom: ${theme.spacing(3)};
    }
  `,
});
