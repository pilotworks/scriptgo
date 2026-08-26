import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { OverallAuditReport } from '../types';

interface APIsPageProps {
  auditData: OverallAuditReport;
}

export const APIsPage: React.FC<APIsPageProps> = ({ auditData }) => {
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'verified' | 'missing'>('all');
  const [selectedModule, setSelectedModule] = useState<string>('all');

  const moduleReports = auditData?.module_reports || {};
  const modulesList = Object.keys(moduleReports).sort();

  // Collect all APIs
  const allAPIs: {
    module: string;
    fullName: string;
    name: string;
    kind: string;
    isVerified: boolean;
    corpusTest?: string;
  }[] = [];

  for (const [modName, mod] of Object.entries(moduleReports)) {
    if (!mod.results) continue;
    for (const res of mod.results) {
      const api = res.spec_api;
      const isVerified = res.status === 'VERIFIED';
      let corpusTest = '';
      if (res.corpus_tests && res.corpus_tests.length > 0) {
        const t = res.corpus_tests[0];
        corpusTest = (t.file_path || '').replace(/.*testdata\/corpus\//, '').replace(/.*corpus\//, '');
        if (t.line_number) corpusTest += `:${t.line_number}`;
      }

      allAPIs.push({
        module: modName,
        fullName: api.full_name || api.name,
        name: api.name,
        kind: api.kind || 'method',
        isVerified,
        corpusTest,
      });
    }
  }

  const filteredAPIs = allAPIs.filter((api) => {
    if (selectedModule !== 'all' && api.module !== selectedModule) return false;
    if (statusFilter === 'verified' && !api.isVerified) return false;
    if (statusFilter === 'missing' && api.isVerified) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        api.fullName.toLowerCase().includes(q) ||
        api.module.toLowerCase().includes(q) ||
        api.kind.toLowerCase().includes(q)
      );
    }
    return true;
  });

  return (
    <div>
      <div className="text-xs font-mono text-doc-muted mb-6 flex items-center gap-2">
        <Link to="/" className="text-doc-link hover:underline">
          Overview
        </Link>
        <span className="text-doc-dim">/</span>
        <span className="text-white font-semibold">Canonical API Catalog</span>
      </div>

      <header className="pb-5 mb-7 border-b border-doc-border">
        <h1 className="text-3xl font-bold text-white mb-2">Canonical API Catalog</h1>
        <p className="text-sm text-doc-muted">
          Browse and search all {allAPIs.length.toLocaleString()} official Node.js 22 LTS API members.
        </p>
      </header>

      {/* Filter Controls */}
      <div className="flex flex-wrap gap-3 mb-6">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search API member (e.g. readFileSync, setTimeout, createHash)..."
          className="flex-1 min-w-[280px] bg-doc-surface border border-doc-border rounded-md px-3 py-2 text-sm text-white outline-none focus:border-doc-blue"
        />

        <select
          value={selectedModule}
          onChange={(e) => setSelectedModule(e.target.value)}
          className="bg-doc-surface border border-doc-border rounded-md px-3 py-2 text-sm text-doc-text outline-none font-mono"
        >
          <option value="all">All Modules ({modulesList.length})</option>
          {modulesList.map((m) => (
            <option key={m} value={m}>
              node:{m}
            </option>
          ))}
        </select>

        <div className="flex bg-doc-surface border border-doc-border rounded-md p-0.5 text-xs font-semibold">
          <button
            onClick={() => setStatusFilter('all')}
            className={`px-3 py-1.5 rounded ${
              statusFilter === 'all' ? 'bg-doc-blue text-white' : 'text-doc-muted hover:text-white'
            }`}
          >
            All ({allAPIs.length})
          </button>
          <button
            onClick={() => setStatusFilter('verified')}
            className={`px-3 py-1.5 rounded ${
              statusFilter === 'verified'
                ? 'bg-doc-green-bg text-doc-green'
                : 'text-doc-muted hover:text-white'
            }`}
          >
            Verified (153)
          </button>
          <button
            onClick={() => setStatusFilter('missing')}
            className={`px-3 py-1.5 rounded ${
              statusFilter === 'missing'
                ? 'bg-doc-red-bg text-doc-red'
                : 'text-doc-muted hover:text-white'
            }`}
          >
            Missing ({allAPIs.length - 153})
          </button>
        </div>
      </div>

      {/* APIs Table */}
      <div className="border border-doc-border rounded-lg overflow-hidden bg-doc-surface">
        <table className="w-full text-left text-[13px] border-collapse">
          <thead>
            <tr className="bg-white/[0.02] border-b border-doc-border text-xs text-doc-muted font-semibold">
              <th className="py-2.5 px-3.5 w-[20%]">Module</th>
              <th className="py-2.5 px-3.5 w-[34%]">Member</th>
              <th className="py-2.5 px-3.5 w-[14%]">Kind</th>
              <th className="py-2.5 px-3.5 w-[16%]">Status</th>
              <th className="py-2.5 px-3.5">Corpus Verification</th>
            </tr>
          </thead>
          <tbody>
            {filteredAPIs.slice(0, 500).map((api, idx) => (
              <tr
                key={idx}
                className="border-b border-doc-border-subtle last:border-b-0 hover:bg-white/[0.015] transition-colors"
              >
                <td className="py-2 px-3.5">
                  <Link
                    to={`/modules/${api.module}`}
                    className="font-mono text-xs text-doc-muted hover:text-doc-link hover:underline"
                  >
                    node:{api.module}
                  </Link>
                </td>
                <td className="py-2 px-3.5">
                  <span className="font-mono font-medium text-[#79c0ff]">{api.fullName}</span>
                </td>
                <td className="py-2 px-3.5">
                  <span className="font-mono text-[11px] text-doc-dim">{api.kind}</span>
                </td>
                <td className="py-2 px-3.5">
                  <span
                    className={`inline-block font-mono text-[11px] font-semibold px-1.5 py-0.5 rounded ${
                      api.isVerified
                        ? 'bg-doc-green-bg text-doc-green'
                        : 'bg-doc-red-bg text-doc-red'
                    }`}
                  >
                    {api.isVerified ? 'VERIFIED' : 'MISSING'}
                  </span>
                </td>
                <td className="py-2 px-3.5">
                  {api.corpusTest ? (
                    <Link
                      to={`/modules/${api.module}`}
                      className="font-mono text-xs text-doc-green hover:underline"
                    >
                      {api.corpusTest}
                    </Link>
                  ) : (
                    <span className="text-doc-dim">—</span>
                  )}
                </td>
              </tr>
            ))}

            {filteredAPIs.length === 0 && (
              <tr>
                <td colSpan={5} className="py-8 text-center text-doc-dim">
                  No matching APIs found for query.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {filteredAPIs.length > 500 && (
        <div className="text-xs text-doc-dim text-center py-4">
          Showing first 500 of {filteredAPIs.length.toLocaleString()} matching APIs.
        </div>
      )}
    </div>
  );
};
