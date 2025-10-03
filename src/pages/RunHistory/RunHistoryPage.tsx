import React, { useState, useEffect } from 'react';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2, Button, LoadingPlaceholder } from '@grafana/ui';
import { Run } from '../../types/types';
import { getBackendSrv } from '@grafana/runtime';

interface RunHistoryPageProps {
  onNavigate: (page: string) => void;
  scheduleId: number | null | undefined;
}

export const RunHistoryPage: React.FC<RunHistoryPageProps> = ({ onNavigate, scheduleId }) => {
  const styles = useStyles2(getStyles);
  const [runs, setRuns] = useState<Run[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (scheduleId) {
      loadRuns();
    }
  }, [scheduleId]);

  const loadRuns = async () => {
    try {
      const response = await getBackendSrv().get(
        `/api/plugins/grafana-app-reporting/resources/api/schedules/${scheduleId}/runs`
      );
      setRuns(response.runs || []);
    } catch (error) {
      console.error('Failed to load runs:', error);
    } finally {
      setLoading(false);
    }
  };

  const downloadArtifact = async (runId: number) => {
    window.open(`/api/plugins/grafana-app-reporting/resources/api/runs/${runId}/artifact`, '_blank');
  };

  if (loading) {
    return <LoadingPlaceholder text="Loading run history..." />;
  }

  const safeRuns = Array.isArray(runs) ? runs : [];

  return (
    <div className={styles.container}>
      <h2>Run History</h2>
      {safeRuns.length === 0 ? (
        <p>No runs yet</p>
      ) : (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Started</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Status</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Duration</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Pages</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Size</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Error</th>
              <th style={{ textAlign: 'left', padding: '8px', borderBottom: '2px solid #ddd' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {safeRuns.map((run) => {
              const status = run.status;
              const statusClass =
                status === 'completed'
                  ? styles.statusSuccess
                  : status === 'failed'
                  ? styles.statusError
                  : styles.statusPending;

              const duration = run.finished_at
                ? `${((new Date(run.finished_at).getTime() - new Date(run.started_at).getTime()) / 1000).toFixed(1)}s`
                : 'Running...';

              const mb = run.bytes / 1024 / 1024;
              const size = mb > 1 ? `${mb.toFixed(2)} MB` : `${(run.bytes / 1024).toFixed(2)} KB`;

              return (
                <tr key={run.id}>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                    {new Date(run.started_at).toLocaleString()}
                  </td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                    <span className={statusClass}>{status}</span>
                  </td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{duration}</td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{run.rendered_pages}</td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{size}</td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>{run.error_text || '-'}</td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                    {run.status === 'completed' && run.artifact_path ? (
                      // @ts-ignore
                      <Button size="sm" variant="secondary" icon="download-alt" onClick={() => downloadArtifact(run.id)}>
                        Download
                      </Button>
                    ) : null}
                  </td>
                </tr>
              );
            })}
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
  statusSuccess: css`
    color: ${theme.colors.success.text};
  `,
  statusError: css`
    color: ${theme.colors.error.text};
  `,
  statusPending: css`
    color: ${theme.colors.warning.text};
  `,
});
