import React from 'react';
import { ModuleAuditReport, CanonicalAPI } from '../types';
import { CodeBlock } from './code-block';

interface ModuleTableProps {
  moduleName: string;
  report: ModuleAuditReport;
  searchTerm?: string;
}

export const ModuleTable: React.FC<ModuleTableProps> = ({ moduleName, report, searchTerm }) => {
  const isGuide = report.total_official_apis === 0;
  const rate = report.coverage_rate_percent || 0;
  const isFull = rate >= 100 && report.total_official_apis > 0;

  const filteredResults = (report.results || []).filter((res) => {
    if (!searchTerm) return true;
    const term = searchTerm.toLowerCase();
    return (
      moduleName.toLowerCase().includes(term) ||
      (res.spec_api.full_name || res.spec_api.name || '').toLowerCase().includes(term)
    );
  });

  // Extract all verified items with their docs & corpus code examples
  const verifiedDocItems: {
    specApi: CanonicalAPI;
    filePath: string;
    lineNumber: number;
    snippet: string;
    anchorId: string;
  }[] = [];

  filteredResults.forEach((res, idx) => {
    if (res.status === 'VERIFIED' && res.corpus_tests && res.corpus_tests.length > 0) {
      const t = res.corpus_tests[0];
      if (t.code_snippet) {
        verifiedDocItems.push({
          specApi: res.spec_api,
          filePath: t.file_path,
          lineNumber: t.line_number,
          snippet: t.code_snippet,
          anchorId: `doc-${moduleName}-${idx}`,
        });
      }
    }
  });

  return (
    <section className="mb-16" id={`mod-${moduleName}`}>
      {/* Module Title Header */}
      <div className="flex items-center justify-between pb-3 mb-6 border-b border-doc-border-subtle">
        <h2 className="font-mono text-2xl font-bold text-white">
          {isGuide ? moduleName : `node:${moduleName}`}
        </h2>
        <span
          className={`text-xs font-mono font-semibold px-2.5 py-1 rounded border ${
            isGuide
              ? 'bg-doc-border-subtle text-doc-muted border-doc-border'
              : isFull
              ? 'bg-doc-green-bg text-doc-green border-emerald-500/30'
              : rate > 0
              ? 'bg-doc-border-subtle text-doc-muted border-doc-border'
              : 'bg-doc-red-bg text-doc-red border-red-500/30'
          }`}
        >
          {isGuide
            ? 'Documentation Guide'
            : `${report.verified_apis} / ${report.total_official_apis} Verified (${rate.toFixed(0)}%)`}
        </span>
      </div>

      {/* Content */}
      {isGuide ? (
        <div className="bg-doc-surface border border-doc-border rounded-md p-4 text-doc-muted text-[13px] leading-relaxed">
          <span className="text-doc-link font-semibold mr-1.5">
            ℹ️ Documentation Guide / C++ Specification:
          </span>
          This section in the official Node.js 22 LTS documentation is an architectural guide,
          command-line manual, or C/C++ native addon binding reference rather than a JavaScript
          module export.
        </div>
      ) : filteredResults.length > 0 ? (
        <>
          {/* Main API Members Table (Kept Clean) */}
          <div className="border border-doc-border rounded-lg overflow-hidden bg-doc-surface mb-12 shadow-sm">
            <table className="w-full text-left text-[13px] border-collapse">
              <thead>
                <tr className="bg-white/[0.02] border-b border-doc-border text-xs text-doc-muted font-semibold">
                  <th className="py-3 px-4 w-[34%]">Member</th>
                  <th className="py-3 px-4 w-[14%]">Kind</th>
                  <th className="py-3 px-4 w-[16%]">Status</th>
                  <th className="py-3 px-4">Corpus Verification</th>
                </tr>
              </thead>
              <tbody>
                {filteredResults.map((res, idx) => {
                  const api = res.spec_api;
                  const isVerified = res.status === 'VERIFIED';
                  const anchorId = `doc-${moduleName}-${idx}`;
                  const displayName = api.full_name || api.name;

                  let testFile = <span className="text-doc-dim">—</span>;
                  if (res.corpus_tests && res.corpus_tests.length > 0) {
                    const t = res.corpus_tests[0];
                    let shortPath = (t.file_path || '')
                      .replace(/.*testdata\/corpus\//, '')
                      .replace(/.*corpus\//, '');
                    if (!shortPath && t.tag) shortPath = t.tag;
                    const lineSuffix = t.line_number ? `:${t.line_number}` : '';
                    const extraCount =
                      res.corpus_tests.length > 1 ? ` (+${res.corpus_tests.length - 1})` : '';

                    testFile = (
                      <a
                        href={`#${anchorId}`}
                        className="font-mono text-xs text-doc-green hover:underline cursor-pointer"
                        title="Click to view API details & corpus example below"
                      >
                        {shortPath + lineSuffix + extraCount}
                      </a>
                    );
                  }

                  return (
                    <tr
                      key={idx}
                      className="border-b border-doc-border-subtle last:border-b-0 hover:bg-white/[0.015] transition-colors"
                    >
                      <td className="py-2.5 px-4">
                        <span className="font-mono font-medium text-[#79c0ff]">{displayName}</span>
                      </td>
                      <td className="py-2.5 px-4">
                        <span className="font-mono text-[11px] text-doc-dim">
                          {api.kind || 'method'}
                        </span>
                      </td>
                      <td className="py-2.5 px-4">
                        <span
                          className={`inline-block font-mono text-[11px] font-semibold px-2 py-0.5 rounded ${
                            isVerified
                              ? 'bg-doc-green-bg text-doc-green'
                              : 'bg-doc-red-bg text-doc-red'
                          }`}
                        >
                          {isVerified ? 'VERIFIED' : 'MISSING'}
                        </span>
                      </td>
                      <td className="py-2.5 px-4">{testFile}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Clean API Details & Corpus Examples */}
          {verifiedDocItems.length > 0 && (
            <div className="pt-4">
              <div className="pb-3 mb-8 border-b border-doc-border flex items-center justify-between">
                <h3 className="text-lg font-bold text-white font-mono">
                  API Details & Corpus Examples
                </h3>
                <span className="text-xs text-doc-dim font-mono">
                  {verifiedDocItems.length} verified APIs
                </span>
              </div>

              <div className="space-y-12">
                {verifiedDocItems.map((item, idx) => {
                  const api = item.specApi;
                  const signature = api.raw_signature || api.full_name || api.name;

                  return (
                    <article
                      key={item.anchorId}
                      id={item.anchorId}
                      className="scroll-mt-8 space-y-3.5"
                    >
                      {/* API Heading */}
                      <div>
                        <div className="flex flex-wrap items-center gap-2.5 mb-2">
                          <h4 className="text-lg font-bold font-mono text-[#79c0ff] tracking-tight">
                            {signature}
                          </h4>
                          <span className="text-[11px] font-mono text-doc-dim px-2 py-0.5 rounded bg-doc-surface border border-doc-border">
                            {api.kind || 'method'}
                          </span>
                          {api.stability_text && (
                            <span className="text-[11px] font-mono text-doc-muted px-2 py-0.5 rounded bg-doc-surface border border-doc-border">
                              {api.stability_text}
                            </span>
                          )}
                          <span className="text-[10px] font-mono font-semibold px-2 py-0.5 rounded bg-doc-green-bg text-doc-green border border-emerald-500/30">
                            VERIFIED
                          </span>
                        </div>

                        {/* Import & Usage Signature */}
                        {api.kind === 'class' ? (
                          <div className="text-xs font-mono py-1">
                            <span className="text-[#ff7b72]">import</span>{' '}
                            <span className="text-doc-text">&#123;</span>{' '}
                            <span className="text-[#d2a8ff] font-medium">{api.name}</span>{' '}
                            <span className="text-doc-text">&#125;</span>{' '}
                            <span className="text-[#ff7b72]">from</span>{' '}
                            <span className="text-[#7ee787]">"node:{moduleName}"</span>;
                            <span className="text-doc-dim ml-3">// const instance = new {api.name}(...);</span>
                          </div>
                        ) : (
                          <div className="text-xs font-mono py-1">
                            <span className="text-[#ff7b72]">import</span>{' '}
                            <span className="text-doc-text">&#123;</span>{' '}
                            <span className="text-[#79c0ff] font-medium">{api.name}</span>{' '}
                            <span className="text-doc-text">&#125;</span>{' '}
                            <span className="text-[#ff7b72]">from</span>{' '}
                            <span className="text-[#7ee787]">"node:{moduleName}"</span>;
                          </div>
                        )}
                      </div>

                      {/* Parameters List */}
                      {api.params && api.params.length > 0 && (
                        <div className="space-y-1 pt-1">
                          <div className="text-xs font-mono font-semibold text-doc-muted uppercase tracking-wider">
                            Parameters
                          </div>
                          <ul className="list-disc list-inside space-y-1 text-xs font-mono text-doc-text">
                            {api.params.map((p, pIdx) => (
                              <li key={pIdx}>
                                <span className="font-semibold text-white">{p.name}</span>
                                {p.type && (
                                  <span className="text-[#79c0ff] ml-1.5 font-normal">
                                    &lt;{p.type}&gt;
                                  </span>
                                )}
                                {p.optional && (
                                  <span className="text-doc-muted text-[11px] ml-1.5 font-normal">
                                    (optional)
                                  </span>
                                )}
                              </li>
                            ))}
                          </ul>
                        </div>
                      )}

                      {/* Return Info */}
                      {api.return && api.return.type && (
                        <div className="text-xs font-mono text-doc-muted">
                          <span className="font-semibold text-doc-muted uppercase tracking-wider text-[11px] mr-2">
                            Returns:
                          </span>
                          <span className="text-doc-green font-semibold">&lt;{api.return.type}&gt;</span>
                        </div>
                      )}

                      {/* Corpus Test Example (Highlighted TypeScript CodeBlock) */}
                      <div className="pt-2">
                        <div className="mb-1 text-xs font-mono font-semibold text-doc-muted uppercase tracking-wider text-[11px]">
                          Example (Corpus Test Case)
                        </div>

                        <CodeBlock
                          code={item.snippet}
                          filePath={item.filePath}
                          lineNumber={item.lineNumber}
                          language="typescript"
                        />
                      </div>

                      {/* Section Separator */}
                      {idx < verifiedDocItems.length - 1 && (
                        <hr className="border-doc-border-subtle pt-4" />
                      )}
                    </article>
                  );
                })}
              </div>
            </div>
          )}
        </>
      ) : (
        <div className="text-doc-dim text-[13px] py-3">No members match search query.</div>
      )}
    </section>
  );
};
