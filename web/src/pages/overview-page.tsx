import React from 'react';
import { OverallAuditReport, BenchmarkReport } from '../types';
import { SummaryCards } from '../components/summary-cards';
import { ModuleTable } from '../components/module-table';
import { BenchmarkGrid } from '../components/benchmark-grid';

interface OverviewPageProps {
  auditData: OverallAuditReport;
  benchData: BenchmarkReport;
}

export const OverviewPage: React.FC<OverviewPageProps> = ({ auditData, benchData }) => {
  const moduleReports = auditData?.module_reports || {};
  const entries = Object.entries(moduleReports).sort(([a], [b]) => a.localeCompare(b));

  return (
    <div>
      <header className="pb-5 mb-7 border-b border-doc-border">
        <h1 className="text-3xl font-bold text-white mb-2">
          Node.js API Parity Documentation
        </h1>
        <p className="text-sm text-doc-muted">
          Official Node.js 22 LTS API Specification Audit & Corpus Test Verification Matrix.
        </p>
      </header>

      <SummaryCards auditData={auditData} benchData={benchData} />

      <div className="space-y-4">
        {entries.map(([name, report]) => (
          <ModuleTable key={name} moduleName={name} report={report} />
        ))}
      </div>

      <BenchmarkGrid benchData={benchData} />
    </div>
  );
};
