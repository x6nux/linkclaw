"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Components } from "react-markdown";

const components: Components = {
  p: ({ children }) => <p className="my-0.5 leading-relaxed">{children}</p>,
  code: ({ children, className }) => {
    const isBlock = className?.includes("language-");
    if (isBlock) {
      return (
        <code className={`${className} block bg-black/40 rounded px-3 py-2 text-xs overflow-x-auto my-1`}>
          {children}
        </code>
      );
    }
    return (
      <code className="bg-white/10 rounded px-1 py-0.5 text-xs font-mono">{children}</code>
    );
  },
  pre: ({ children }) => <pre className="my-1 overflow-x-auto">{children}</pre>,
  a: ({ href, children }) => (
    <a href={href} target="_blank" rel="noopener noreferrer" className="text-blue-300 underline hover:text-blue-200">
      {children}
    </a>
  ),
  ul: ({ children }) => <ul className="list-disc list-inside my-0.5 space-y-0.5">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal list-inside my-0.5 space-y-0.5">{children}</ol>,
  blockquote: ({ children }) => (
    <blockquote className="border-l-2 border-zinc-500 pl-2 my-1 text-zinc-300 italic">{children}</blockquote>
  ),
  h1: ({ children }) => <p className="font-bold text-base my-0.5">{children}</p>,
  h2: ({ children }) => <p className="font-bold text-sm my-0.5">{children}</p>,
  h3: ({ children }) => <p className="font-semibold text-sm my-0.5">{children}</p>,
  table: ({ children }) => (
    <div className="overflow-x-auto my-1">
      <table className="min-w-full text-xs border-collapse">{children}</table>
    </div>
  ),
  thead: ({ children }) => <thead className="border-b border-zinc-600">{children}</thead>,
  tbody: ({ children }) => <tbody className="divide-y divide-zinc-700/50">{children}</tbody>,
  tr: ({ children }) => <tr>{children}</tr>,
  th: ({ children }) => (
    <th className="px-2 py-1 text-left font-semibold text-zinc-300">{children}</th>
  ),
  td: ({ children }) => (
    <td className="px-2 py-1 text-zinc-400">{children}</td>
  ),
};

export function MarkdownContent({ content }: { content: string }) {
  return (
    <div className="markdown-msg text-sm leading-relaxed [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>{content}</ReactMarkdown>
    </div>
  );
}
