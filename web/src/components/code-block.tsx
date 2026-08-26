import React, { useState } from 'react';
import Prism from 'prismjs';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-typescript';

interface CodeBlockProps {
  code: string;
  language?: string;
  filePath?: string;
  lineNumber?: number;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({
  code,
  language = 'typescript',
  filePath,
  lineNumber,
}) => {
  const [copied, setCopied] = useState(false);

  const highlightedHtml = Prism.highlight(
    code,
    Prism.languages[language] || Prism.languages.typescript || Prism.languages.javascript,
    language
  );

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  let shortPath = (filePath || '').replace(/.*testdata\/corpus\//, '').replace(/.*corpus\//, '');
  const lineSuffix = lineNumber ? `:${lineNumber}` : '';

  const cleanPath = (filePath || '').replace(/^\/?/, '');
  const githubUrl = cleanPath
    ? `https://github.com/pilotworks/scriptgo/blob/main/${cleanPath}${lineNumber ? `#L${lineNumber}` : ''}`
    : null;

  return (
    <div className="rounded-lg border border-doc-border overflow-hidden bg-[#03070d] shadow-sm my-2">
      {/* Code Header Bar */}
      <div className="flex items-center justify-between px-3.5 py-1.5 bg-[#090d14] border-b border-doc-border text-xs font-mono select-none">
        <div className="flex items-center gap-2.5">
          {githubUrl ? (
            <a
              href={githubUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 text-doc-green hover:underline font-medium hover:text-[#7ee787] transition-colors"
              title="Open source file on GitHub"
            >
              <span>📄 {shortPath}{lineSuffix}</span>
              <svg
                className="w-3 h-3 text-doc-dim"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                />
              </svg>
            </a>
          ) : shortPath ? (
            <span className="text-doc-green font-medium">
              📄 {shortPath}
              {lineSuffix}
            </span>
          ) : null}

          <span className="text-[11px] text-doc-dim uppercase font-semibold">{language}</span>
        </div>

        <div className="flex items-center gap-3">
          {githubUrl && (
            <a
              href={githubUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[11px] text-doc-muted hover:text-[#79c0ff] flex items-center gap-1 transition-colors"
            >
              <span>GitHub</span>
              <svg
                className="w-3 h-3"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                />
              </svg>
            </a>
          )}

          <button
            onClick={handleCopy}
            className="text-[11px] text-doc-muted hover:text-white px-2 py-0.5 rounded hover:bg-white/[0.06] transition-colors"
          >
            {copied ? '✓ Copied' : 'Copy'}
          </button>
        </div>
      </div>

      {/* Code Content */}
      <pre className="p-4 font-mono text-[13px] text-doc-text overflow-x-auto leading-relaxed whitespace-pre selection:bg-doc-blue selection:text-white">
        <code dangerouslySetInnerHTML={{ __html: highlightedHtml }} />
      </pre>
    </div>
  );
};
