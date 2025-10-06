import React, { useState, useEffect } from 'react';
import { AppRootProps } from '@grafana/data';
import { TabsBar, Tab, TabContent } from '@grafana/ui';
import { SchedulesPage } from '../pages/Schedules/SchedulesPage';
import { ScheduleEditPage } from '../pages/Schedules/ScheduleEditPage';
import { RunHistoryPage } from '../pages/RunHistory/RunHistoryPage';
import { SettingsPage } from '../pages/Settings/SettingsPage';
import { DocumentationPage } from '../pages/Documentation/DocumentationPage';

type Page = 'schedules' | 'schedule-new' | 'schedule-edit' | 'run-history' | 'settings' | 'documentation';

export const App: React.FC<AppRootProps> = (props) => {
  const [currentPage, setCurrentPage] = useState<Page>('schedules');
  const [selectedScheduleId, setSelectedScheduleId] = useState<number | null>(null);

  // Check URL path to determine initial page
  useEffect(() => {
    const path = props.path || '';
    if (path.includes('settings')) {
      setCurrentPage('settings');
    } else if (path.includes('documentation')) {
      setCurrentPage('documentation');
    } else {
      setCurrentPage('schedules');
    }
  }, [props.path]);

  const navigate = (page: Page, scheduleId?: number) => {
    setCurrentPage(page);
    if (scheduleId) {
      setSelectedScheduleId(scheduleId);
    }
  };

  const renderPage = () => {
    switch (currentPage) {
      case 'schedules':
        return <SchedulesPage onNavigate={navigate} />;
      case 'schedule-new':
        return <ScheduleEditPage onNavigate={navigate} isNew={true} />;
      case 'schedule-edit':
        return <ScheduleEditPage onNavigate={navigate} isNew={false} scheduleId={selectedScheduleId} />;
      case 'run-history':
        return <RunHistoryPage onNavigate={navigate} scheduleId={selectedScheduleId} />;
      case 'settings':
        return <SettingsPage onNavigate={navigate} />;
      case 'documentation':
        return <DocumentationPage />;
      default:
        return <SchedulesPage onNavigate={navigate} />;
    }
  };

  // Show tabs only on main pages (not on edit/new/history pages)
  const showTabs = currentPage === 'schedules' || currentPage === 'settings' || currentPage === 'documentation';

  return (
    <div>
      {showTabs && (
        <TabsBar>
          {/* @ts-ignore */}
          <Tab
            label="Schedules"
            active={currentPage === 'schedules'}
            onChangeTab={() => navigate('schedules')}
          />
          {/* @ts-ignore */}
          <Tab
            label="Documentation"
            active={currentPage === 'documentation'}
            onChangeTab={() => navigate('documentation')}
          />
          {/* @ts-ignore */}
          <Tab
            label="Settings"
            active={currentPage === 'settings'}
            onChangeTab={() => navigate('settings')}
          />
        </TabsBar>
      )}
      <div style={{ padding: showTabs ? '0' : '16px' }}>
        {renderPage()}
      </div>
    </div>
  );
};
