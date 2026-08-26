import React, { useState } from 'react';
import { NavLink, useNavigate, useParams, useLocation } from 'react-router-dom';
import { ModuleAuditReport } from '../types';

interface SidebarProps {
  moduleReports: Record<string, ModuleAuditReport>;
}

export const Sidebar: React.FC<SidebarProps> = ({ moduleReports }) => {
  const [currentFilter, setCurrentFilter] = useState<'all' | 'complete' | 'untested'>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const { moduleName } = useParams<{ moduleName?: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  const entries = Object.entries(moduleReports || {}).sort(([a], [b]) => a.localeCompare(b));

  const totalCount = entries.length;
  const completeCount = entries.filter(
    ([_, m]) => (m.coverage_rate_percent || 0) >= 100 && m.total_official_apis > 0
  ).length;
  const missingCount = totalCount - completeCount;

  const filteredEntries = entries.filter(([name, mod]) => {
    const rate = mod.coverage_rate_percent || 0;
    const isFull = rate >= 100 && mod.total_official_apis > 0;

    if (currentFilter === 'complete' && !isFull) return false;
    if (currentFilter === 'untested' && isFull) return false;

    if (searchTerm) {
      const matchMod = name.toLowerCase().includes(searchTerm.toLowerCase());
      const hasAPI =
        mod.results &&
        mod.results.some((r) =>
          (r.spec_api.full_name || r.spec_api.name || '')
            .toLowerCase()
            .includes(searchTerm.toLowerCase())
        );
      if (!matchMod && !hasAPI) return false;
    }

    return true;
  });

  return (
    <aside className="w-80 min-w-[320px] bg-doc-sidebar border-r border-doc-border flex flex-col h-full select-none">
      {/* Brand Header */}
      <div className="px-5 py-4 border-b border-doc-border">
        <NavLink
          to="/"
          className="flex items-center gap-2.5 text-base font-bold text-doc-text hover:text-white transition-colors"
        >
          <span>ScriptGo Docs</span>
          <span className="text-[11px] px-1.5 py-0.5 bg-doc-border-subtle border border-doc-border rounded text-doc-muted font-mono font-medium">
            v22 LTS
          </span>
        </NavLink>
      </div>

      {/* Main Nav Links */}
      <div className="px-3 py-2 border-b border-doc-border-subtle flex gap-1 text-xs font-semibold">
        <NavLink
          to="/"
          className={({ isActive }) =>
            `flex-1 py-1.5 px-2 text-center rounded transition-colors ${
              isActive && location.pathname === '/'
                ? 'bg-doc-surface text-white border border-doc-border'
                : 'text-doc-muted hover:text-doc-text'
            }`
          }
        >
          Overview
        </NavLink>
        <NavLink
          to="/apis"
          className={({ isActive }) =>
            `flex-1 py-1.5 px-2 text-center rounded transition-colors ${
              isActive
                ? 'bg-doc-surface text-white border border-doc-border'
                : 'text-doc-muted hover:text-doc-text'
            }`
          }
        >
          APIs (1,888)
        </NavLink>
        <NavLink
          to="/benchmark"
          className={({ isActive }) =>
            `flex-1 py-1.5 px-2 text-center rounded transition-colors ${
              isActive
                ? 'bg-doc-surface text-white border border-doc-border'
                : 'text-doc-muted hover:text-doc-text'
            }`
          }
        >
          Benchmark
        </NavLink>
      </div>

      {/* Search Bar */}
      <div className="px-5 py-3 border-b border-doc-border-subtle">
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Search modules & APIs..."
          className="w-full bg-doc-bg border border-doc-border rounded-md px-2.5 py-1.5 text-sm text-doc-text outline-none focus:border-doc-blue placeholder:text-doc-dim"
        />
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-1 px-5 py-2 border-b border-doc-border-subtle bg-white/[0.01]">
        <button
          onClick={() => setCurrentFilter('all')}
          className={`flex-1 py-1 text-[11px] font-semibold rounded border transition-colors ${
            currentFilter === 'all'
              ? 'bg-doc-surface border-doc-border text-doc-text'
              : 'border-transparent text-doc-muted hover:text-doc-text'
          }`}
        >
          All ({totalCount})
        </button>
        <button
          onClick={() => setCurrentFilter('complete')}
          className={`flex-1 py-1 text-[11px] font-semibold rounded border transition-colors ${
            currentFilter === 'complete'
              ? 'bg-doc-surface border-doc-border text-doc-text'
              : 'border-transparent text-doc-muted hover:text-doc-text'
          }`}
        >
          Complete ({completeCount})
        </button>
        <button
          onClick={() => setCurrentFilter('untested')}
          className={`flex-1 py-1 text-[11px] font-semibold rounded border transition-colors ${
            currentFilter === 'untested'
              ? 'bg-doc-surface border-doc-border text-doc-text'
              : 'border-transparent text-doc-muted hover:text-doc-text'
          }`}
        >
          Missing ({missingCount})
        </button>
      </div>

      {/* Module List */}
      <div className="flex-1 overflow-y-auto px-2.5 py-3 space-y-0.5">
        <div className="text-[11px] font-semibold uppercase tracking-wider text-doc-dim px-2.5 py-1.5">
          Node.js Modules
        </div>

        {filteredEntries.map(([name, mod]) => {
          const rate = mod.coverage_rate_percent || 0;
          const isFull = rate >= 100 && mod.total_official_apis > 0;
          const isGuide = mod.total_official_apis === 0;
          const isActive = moduleName === name;

          let badgeText = `${mod.verified_apis}/${mod.total_official_apis}`;
          if (isFull) badgeText = '100%';
          if (isGuide) badgeText = 'Guide';

          return (
            <div
              key={name}
              onClick={() => navigate(`/modules/${name}`)}
              className={`flex items-center justify-between px-2.5 py-1.5 rounded-md font-mono text-[13px] cursor-pointer transition-colors ${
                isActive
                  ? 'bg-doc-blue text-white'
                  : 'text-doc-muted hover:bg-doc-surface hover:text-doc-text'
              }`}
            >
              <span>{isGuide ? name : `node:${name}`}</span>
              <span
                className={`text-[10px] font-semibold px-1.5 py-0.5 rounded ${
                  isActive
                    ? 'bg-white/20 text-white'
                    : isFull
                    ? 'bg-doc-green-bg text-doc-green'
                    : 'bg-doc-border-subtle text-doc-dim'
                }`}
              >
                {badgeText}
              </span>
            </div>
          );
        })}

        {filteredEntries.length === 0 && (
          <div className="text-xs text-doc-dim text-center py-6">No modules match filter</div>
        )}
      </div>
    </aside>
  );
};
