import React from 'react';
import { BenchmarkReport } from '../types';

interface BenchmarkGridProps {
  benchData: BenchmarkReport | null;
}

export const BenchmarkGrid: React.FC<BenchmarkGridProps> = ({ benchData }) => {
  const stats = benchData?.category_stats || {};
  const entries = Object.entries(stats).sort(([a], [b]) => a.localeCompare(b));

  return (
    <section className="mt-10 pt-6 border-t border-doc-border">
      <h2 className="text-lg font-semibold text-white mb-1.5">Corpus Parity Benchmark Suite</h2>
      <p className="text-[13px] text-doc-muted mb-4">
        214 Native execution test cases verified across 17 categories.
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
        {entries.map(([cat, s]) => (
          <div
            key={cat}
            className="bg-doc-surface border border-doc-border rounded-md px-3.5 py-2.5 flex items-center justify-between font-mono text-xs"
          >
            <span className="text-doc-text font-medium">{cat}</span>
            <span className="bg-doc-green-bg text-doc-green font-semibold px-1.5 py-0.5 rounded text-[11px]">
              {s.passed}/{s.total} PASS
            </span>
          </div>
        ))}

        {entries.length === 0 && (
          <div className="text-xs text-doc-dim">Benchmark summary available in CLI.</div>
        )}
      </div>
    </section>
  );
};
