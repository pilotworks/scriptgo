import React from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from '../components/sidebar';
import { OverallAuditReport } from '../types';

interface RootLayoutProps {
  auditData: OverallAuditReport;
}

export const RootLayout: React.FC<RootLayoutProps> = ({ auditData }) => {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-doc-bg text-doc-text font-sans antialiased">
      <Sidebar moduleReports={auditData.module_reports || {}} />
      <main className="flex-1 h-full overflow-y-auto px-6 py-6 sm:px-10 sm:py-8 lg:px-12 lg:py-10">
        <div className="w-full">
          <Outlet />
        </div>
      </main>
    </div>
  );
};
