import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { RootLayout } from './layouts/root-layout';
import { OverviewPage } from './pages/overview-page';
import { ModulePage } from './pages/module-page';
import { BenchmarkPage } from './pages/benchmark-page';
import { APIsPage } from './pages/apis-page';
import { OverallAuditReport, BenchmarkReport } from './types';
import auditDataJson from './data/audit-report.json';
import benchDataJson from './data/benchmark-report.json';

const auditData = auditDataJson as unknown as OverallAuditReport;
const benchData = benchDataJson as unknown as BenchmarkReport;

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<RootLayout auditData={auditData} />}>
          <Route index element={<OverviewPage auditData={auditData} benchData={benchData} />} />
          <Route path="modules/:moduleName" element={<ModulePage auditData={auditData} />} />
          <Route path="benchmark" element={<BenchmarkPage benchData={benchData} />} />
          <Route path="apis" element={<APIsPage auditData={auditData} />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};
