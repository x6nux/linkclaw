"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { ContextSearchResult } from "@/lib/types";
import { cn } from "@/lib/utils";

interface Props {
  onClose: () => void;
}

export function ContextSearch({ onClose }: Props) {
  const t = useTranslations("context");
  const [query, setQuery] = useState("");
  const [isSearching, setIsSearching] = useState(false);
  const [results, setResults] = useState<ContextSearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;

    setIsSearching(true);
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/context/search`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ query, directory_ids: [] }),
      });
      if (!res.ok) throw new Error(t("errors.searchFailed"));
      const data = await res.json();
      setResults(data.data || []);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("errors.searchFailed"));
    } finally {
      setIsSearching(false);
      setHasSearched(true);
    }
  }

  function getRelevanceColor(score: number) {
    if (score >= 0.8) return "text-green-400";
    if (score >= 0.5) return "text-yellow-400";
    return "text-zinc-400";
  }

  function getRelevanceLabel(score: number) {
    if (score >= 0.8) return t("relevance.high");
    if (score >= 0.5) return t("relevance.medium");
    return t("relevance.low");
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-3xl mx-4 shadow-2xl max-h-[85vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="border-b border-zinc-800 px-5 pt-5 pb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-zinc-50">{t("title")}</h2>
          <button
            onClick={onClose}
            className="text-zinc-500 hover:text-zinc-300 transition-colors text-xl leading-none"
          >
            &times;
          </button>
        </div>

        {/* Search Form */}
        <form onSubmit={handleSearch} className="p-4 border-b border-zinc-800">
          <div className="flex gap-2">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t("searchPlaceholder")}
              className="flex-1 bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2 text-zinc-50 text-sm focus:outline-none focus:border-blue-500 transition-colors"
            />
            <button
              type="submit"
              disabled={isSearching || !query.trim()}
              className="px-6 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors disabled:opacity-40"
            >
              {isSearching ? t("searching") : t("search")}
            </button>
          </div>
        </form>

        {/* Results */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {!hasSearched && (
            <div className="text-center py-12">
              <p className="text-zinc-500 text-sm">{t("searchHint")}</p>
            </div>
          )}

          {hasSearched && results.length === 0 && !isSearching && (
            <div className="text-center py-12">
              <p className="text-zinc-400">{t("noResults")}</p>
              <p className="text-zinc-600 text-sm mt-2">{t("noResultsHint")}</p>
            </div>
          )}

          {results.map((result, index) => (
            <div
              key={`${result.file_path}-${index}`}
              className="bg-zinc-800/50 border border-zinc-700 rounded-lg p-4 hover:border-zinc-600 transition-colors"
            >
              <div className="flex items-start justify-between mb-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-mono text-blue-400 bg-blue-500/10 px-1.5 py-0.5 rounded">
                      {result.language || "unknown"}
                    </span>
                    <span className={cn("text-xs", getRelevanceColor(result.relevance))}>
                      {getRelevanceLabel(result.relevance)} ({(result.relevance * 100).toFixed(0)}%)
                    </span>
                  </div>
                  <h3 className="text-sm font-medium text-zinc-50 truncate">{result.file_path}</h3>
                </div>
              </div>

              <p className="text-zinc-400 text-sm line-clamp-3 mb-2">{result.summary}</p>

              {result.reason && (
                <p className="text-zinc-500 text-xs italic">
                  <span className="font-medium">{t("matchReason")}：</span>{result.reason}
                </p>
              )}

              {result.line_count && (
                <p className="text-zinc-600 text-xs mt-2">
                  {result.line_count} {t("lines")}
                </p>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
