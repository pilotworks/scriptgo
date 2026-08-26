import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface MarkdownViewProps {
  content?: string;
  className?: string;
}

export const MarkdownView: React.FC<MarkdownViewProps> = ({ content, className = '' }) => {
  if (!content) return null;

  return (
    <div
      className={`text-[13px] text-doc-text leading-relaxed font-sans opacity-90 space-y-2 [&_p]:mb-2 [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:bg-[#161b22] [&_code]:border [&_code]:border-[#30363d] [&_code]:text-[#79c0ff] [&_code]:font-mono [&_code]:text-xs [&_a]:text-doc-link [&_a]:underline [&_ul]:list-disc [&_ul]:list-inside [&_ul]:space-y-1 [&_strong]:text-white [&_strong]:font-semibold ${className}`}
    >
      <ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeRaw]}>
        {content}
      </ReactMarkdown>
    </div>
  );
};
