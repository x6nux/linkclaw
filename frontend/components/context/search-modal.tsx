"use client";

import { useState, useEffect, useCallback } from "react";
import { FileCode2, Loader2, Search, X } from "lucide-react";
import { api } from "@/lib/api";
import type { SearchResult } from "@/lib/types";

type SearchResponse =
  | SearchResult[]
  | {
      data?: SearchResult[];
      results?: SearchResult[];
    };

function normalizeResults(response: SearchResponse): SearchResult[] {
  if (Array.isArray(response)) return response;
  if (Array.isArray(response.results)) return response.results;
  return response.data ?? [];
}

interface SearchModalProps {
  open: boolean;
  onClose: () => void;
}

export function SearchModal({ open, onClose }: SearchModalProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const [error, setError] = useState("");

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    const nextQuery = query.trim();
    if (!nextQuery) {
      setError("请输入搜索关键词");
      return;
    }

    setIsLoading(true);
    try {
      const response = await api.post<SearchResponse>("/api/v1/indexing/search", {
        query: nextQuery,
      });
      const nextResults = normalizeResults(response)
        .slice()
        .sort((a, b) => b.score - a.score);
      setResults(nextResults);
      setHasSearched(true);
    } catch (err) {
      setResults([]);
      setHasSearched(true);
      setError(err instanceof Error ? err.message : "搜索失败");
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose();
    }
  }, [onClose]);

  useEffect(() => {
    if (!open) return;
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open, handleKeyDown]);

  if (!open) return null;

  const inputClass =
    "w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-3xl mx-4 shadow-2xl max-h-[85vh] flex flex-col">
        {/* Header */}
        <div className="border-b border-zinc-800 px-5 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Search className="w-4 h-4 text-zinc-400" />
            <h2 className="text-sm font-medium text-zinc-200">语义搜索</h2>
          </div>
          <button
            onClick={onClose}
            className="text-zinc-500 hover:text-zinc-300 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Search Form */}
        <form onSubmit={handleSearch} className="p-5 pb-0 space-y-3">
          <div className="flex gap-3">
            <input
              type="text"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setError("");
              }}
              placeholder="例如：处理登录鉴权的逻辑"
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors"
              autoFocus
            />
            <button
              type="submit"
              disabled={isLoading}
              className="px-5 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors flex items-center gap-2 whitespace-nowrap"
            >
              {isLoading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  搜索中...
                </>
              ) : (
                <>
                  <Search className="w-4 h-4" />
                  搜索
                </>
              )}
            </button>
          </div>
          {error && <p className="text-red-400 text-xs">{error}</p>}
        </form>

        {/* Results */}
        <div className="flex-1 overflow-y-auto p-5 space-y-3 min-h-0">
          {hasSearched && !isLoading && (
            <p className="text-xs text-zinc-500">共 {results.length} 条结果</p>
          )}

          {isLoading ? (
            <div className="bg-zinc-950 border border-zinc-800 rounded-md p-6 text-center text-zinc-500 text-sm">
              正在搜索...
            </div>
          ) : hasSearched && results.length === 0 ? (
            <div className="bg-zinc-950 border border-zinc-800 rounded-md p-6 text-center text-zinc-500 text-sm">
              未找到相关代码
            </div>
          ) : (
            results.map((result, index) => (
              <div
                key={`${result.file_path}-${result.chunk_index}-${index}`}
                className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-2"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <div className="flex items-center gap-1.5">
                      <FileCode2 className="w-3.5 h-3.5 text-zinc-500 flex-shrink-0" />
                      <p className="text-sm text-zinc-200 break-all">{result.file_path}</p>
                    </div>
                    <p className="mt-1 text-xs text-zinc-500">
                      行 {result.start_line}-{result.end_line}
                    </p>
                  </div>
                  <span className="text-xs text-zinc-400 font-mono shrink-0">
                    score {result.score.toFixed(3)}
                  </span>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <span className="inline-flex px-2 py-0.5 rounded border text-xs bg-blue-500/10 text-blue-400 border-blue-500/20">
                    {result.language || "unknown"}
                  </span>
                  {result.symbols ? (
                    <span className="text-xs text-zinc-500 break-all">符号：{result.symbols}</span>
                  ) : null}
                </div>

                <pre className="p-2 rounded border border-zinc-800 bg-zinc-900 text-xs text-zinc-300 font-mono whitespace-pre-wrap break-words overflow-x-auto max-h-40 overflow-y-auto">
                  {result.content}
                </pre>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
