import React from 'react';
import { Input, Button } from '@grafana/ui';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';

interface VariablesEditorProps {
  value: Record<string, string>;
  onChange: (value: Record<string, string>) => void;
}

export const VariablesEditor: React.FC<VariablesEditorProps> = ({ value, onChange }) => {
  const styles = useStyles2(getStyles);
  const entries = Object.entries(value);

  const addVariable = () => {
    onChange({ ...value, '': '' });
  };

  const updateKey = (oldKey: string, newKey: string) => {
    const updated = { ...value };
    const val = updated[oldKey];
    delete updated[oldKey];
    updated[newKey] = val;
    onChange(updated);
  };

  const updateValue = (key: string, val: string) => {
    onChange({ ...value, [key]: val });
  };

  const removeVariable = (key: string) => {
    const updated = { ...value };
    delete updated[key];
    onChange(updated);
  };

  return (
    <div>
      <div className={styles.header}>
        <span>Variable Name</span>
        <span>Value</span>
        <span></span>
      </div>
      {entries.map(([key, val], index) => (
        <div key={index} className={styles.row}>
          <Input
            value={key}
            onChange={(e) => updateKey(key, e.currentTarget.value)}
            placeholder="variable_name"
          />
          <Input
            value={val}
            onChange={(e) => updateValue(key, e.currentTarget.value)}
            placeholder="value"
          />
              {/* @ts-ignore */}
          <Button size="sm" variant="secondary" icon="trash-alt" onClick={() => removeVariable(key)} />
        </div>
      ))}
              {/* @ts-ignore */}
      <Button size="sm" variant="secondary" icon="plus" onClick={addVariable}>
        Add Variable
      </Button>
    </div>
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  header: css`
    display: grid;
    grid-template-columns: 1fr 1fr 40px;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(1)};
    font-weight: ${theme.typography.fontWeightMedium};
  `,
  row: css`
    display: grid;
    grid-template-columns: 1fr 1fr 40px;
    gap: ${theme.spacing(1)};
    margin-bottom: ${theme.spacing(1)};
  `,
});
