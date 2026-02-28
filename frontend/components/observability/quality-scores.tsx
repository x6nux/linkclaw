"use client";

import { useEffect } from "react";
import { toast } from "sonner";
import { useQualityScores } from "@/hooks/use-observability";
import { formatDate } from "@/lib/utils";

function scoreStyle(score: number | null) {
  if (score === null || score === undefined) return "bg-zinc-500/10 text-zinc-400 border-zinc-500/20";
  if (score >= 0.8) return "bg-green-500/10 text-green-400 border-green-500/20";
  if (score >= 0.6) return "bg-amber-500/10 text-amber-400 border-amber-500/20";
  return "bg-red-500/10 text-red-400 border-red-500/20";
}

export function QualityScores() {
  const { scores, isLoading, error } = useQualityScores({ limit: 50, offset: 0 });

  useEffect(() => {
    if (error) toast.error(error instanceof Error ? error.message : "操作失败");
  }, [error]);

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
      <div className="px-4 py-3 border-b border-zinc-800">
        <h3 className="text-sm font-medium text-zinc-50">会话质量评分</h3>
      </div>
      <div className="divide-y divide-zinc-800">
        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="p-4">
              <div className="animate-pulse h-4 bg-zinc-800 rounded" />
            </div>
          ))
        ) : scores.length === 0 ? (
          <div className="p-8 text-center text-zinc-500 text-sm">暂无质量评分数据</div>
        ) : (
          scores.map((item) => (
            <div key={item.id} className="p-4 hover:bg-zinc-950/50 transition-colors">
              <div className="flex items-center gap-2">
                <span className={`inline-flex px-2 py-0.5 rounded border text-xs font-mono ${scoreStyle(item.overallScore)}`}>
                  {item.overallScore === null || item.overallScore === undefined
                    ? "未评分"
                    : item.overallScore.toFixed(2)}
                </span>
                <span className="text-xs text-zinc-500 font-mono">Trace: {item.traceId.slice(0, 8)}…</span>
                <span className="ml-auto text-xs text-zinc-500">{formatDate(item.createdAt)}</span>
              </div>
              <p className="mt-2 text-sm text-zinc-300 line-clamp-2">{item.feedback ?? "—"}</p>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
