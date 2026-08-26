export interface DocParam {
  name: string;
  type: string;
  desc?: string;
  optional?: boolean;
}

export interface DocReturn {
  type?: string;
  desc?: string;
}

export interface CanonicalAPI {
  module: string;
  class?: string;
  name: string;
  normalized_key: string;
  full_name: string;
  kind: string;
  raw_signature: string;
  desc?: string;
  params?: DocParam[];
  return?: DocReturn;
  stability?: number;
  stability_text?: string;
}

export interface CorpusAPIItem {
  tag: string;
  normalized_key: string;
  module: string;
  file_path: string;
  line_number: number;
  code_snippet?: string;
}

export interface APIAuditResult {
  spec_api: CanonicalAPI;
  status: 'VERIFIED' | 'MISSING';
  corpus_tests?: CorpusAPIItem[];
}

export interface ModuleAuditReport {
  module_name: string;
  total_official_apis: number;
  verified_apis: number;
  missing_apis: number;
  coverage_rate_percent: number;
  results: APIAuditResult[];
  corpus_extra?: CorpusAPIItem[];
}

export interface OverallAuditReport {
  total_modules: number;
  total_official_apis: number;
  total_verified_apis: number;
  total_missing_apis: number;
  overall_coverage_percent: number;
  module_reports: Record<string, ModuleAuditReport>;
}

export interface CategoryStats {
  total: number;
  passed: number;
  failed: number;
}

export interface BenchmarkReport {
  total_cases: number;
  native_passed: number;
  diagnostics_passed: number;
  overall_full_parity: number;
  parity_rate_percent: number;
  execution_time: string;
  runner: string;
  category_stats: Record<string, CategoryStats>;
}
