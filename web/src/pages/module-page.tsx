import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { OverallAuditReport } from '../types';
import { ModuleTable } from '../components/module-table';

interface ModulePageProps {
  auditData: OverallAuditReport;
}

export const ModulePage: React.FC<ModulePageProps> = ({ auditData }) => {
  const { moduleName } = useParams<{ moduleName: string }>();
  const report = moduleName ? auditData.module_reports?.[moduleName] : null;

  if (!moduleName || !report) {
    return (
      <div>
        <div className="text-sm text-doc-muted mb-4">
          <Link to="/" className="text-doc-link hover:underline">
            ← Back to Overview
          </Link>
        </div>
        <div className="bg-doc-surface border border-doc-border rounded-lg p-8 text-center text-doc-dim">
          Module <span className="font-mono text-white">node:{moduleName}</span> not found in specification.
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="text-xs font-mono text-doc-muted mb-6 flex items-center gap-2">
        <Link to="/" className="text-doc-link hover:underline">
          Overview
        </Link>
        <span className="text-doc-dim">/</span>
        <span className="text-white font-semibold">node:{moduleName}</span>
      </div>

      <ModuleTable moduleName={moduleName} report={report} />
    </div>
  );
};
