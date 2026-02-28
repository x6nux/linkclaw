"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useTraces } from "@/hooks/use-observability";
import { formatDate } from "@/lib/utils";
import type { TraceRun, TraceSourceType, TraceStatus } from "@/lib/types";

interface TraceListProps {
  onSelect: (trace: TraceRun) => void;
  selectedId?: string | null;
}

const STATUS_LABELS: Record<TraceStatus, string> = {
  running: "运行中",
  success: "成功",
  error: "错误",
  timeout: "超时",
};

const STATUS_STYLES: Record<TraceStatus, string> = {
  running: "bg-blue-500/10 text-blue-400 border-blue-500/20",
  success: "bg-green-500/10 text-green-400 border-green-500/20",
  error: "bg-red-500/10 text-red-400 border-red-500/20",
  timeout: "bg-amber-500/10 text-amber-400 border-amber-500/20",
};

const SOURCE_LABELS: Record<TraceSourceType, string> = {
  mcp: "MCP",
  http: "HTTP",
  workflow: "Workflow",
  ws: "WebSocket",
};

function formatDuration(ms: number | null) {
  if (ms === null || ms === undefined) return "—";
  if (ms > 1000) return `${(ms / 1000).toFixed(2)}s`;
  return `${ms}ms`;
}

function formatCost(microdollars: number) {
  return `$${(microdollars / 1_000_000).toFixed(4)}`;
}

export function TraceList({ onSelect, selectedId }: TraceListProps) {
  const [status, setStatus] = useState<TraceStatus | "">("");
  const [sourceType, setSourceType] = useState<TraceSourceType | "">("");
  const { traces, total, isLoading, error } = useTraces({
    status: status || undefined,
    sourceType: sourceType || undefined,
    limit: 50,
    offset: 0,
  });

  useEffect(() => {
    if (error) toast.error(error instanceof Error ? error.message : "操作失败");
  }, [error]);

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div className="p-4 border-b border-zinc-800 flex flex-wrap items-center gap-3">
        <select
          value={status}
          onChange={(e) => setStatus((e.target.value as TraceStatus | "") || "")}
          className="bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
        >
          <option value="">全部状态</option>
          <option value="running">运行中</option>
          <option value="success">成功</option>
          <option value="error">错误</option>
          <option value="timeout">超时</option>
        </select>
        <select
          value={sourceType}
          onChange={(e) => setSourceType((e.target.value as TraceSourceType | "") || "")}
          className="bg-zinc-950 border border-zinc-800 rounded px-3 py-1.5 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
        >
          <option value="">全部来源</option>
          <option value="mcp">MCP</option>
          <option value="http">HTTP</option>
          <option value="workflow">Workflow</option>
          <option value="ws">WebSocket</option>
        </select>
        <span className="ml-auto text-xs text-zinc-500">共 {total} 条</span>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left font-medium px-4 py-3">Trace ID</th>
              <th className="text-left font-medium px-4 py-3">状态</th>
              <th className="text-left font-medium px-4 py-3">来源</th>
              <th className="text-left font-medium px-4 py-3">耗时</th>
              <th className="text-left font-medium px-4 py-3">成本</th>
              <th className="text-left font-medium px-4 py-3">开始时间</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 6 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800">
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                </tr>
              ))
            ) : traces.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-zinc-500">
                  暂无 Trace 数据
                </td>
              </tr>
            ) : (
              traces.map((trace) => (
                <tr
                  key={trace.id}
                  onClick={() => onSelect(trace)}
                  className={`border-t border-zinc-800 transition-colors cursor-pointer ${
                    selectedId === trace.id ? "bg-blue-500/5" : "hover:bg-zinc-950/60"
                  }`}
                >
                  <td className="px-4 py-3 text-zinc-300 font-mono">{trace.id.slice(0, 8)}…</td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex px-2 py-0.5 rounded border text-xs ${STATUS_STYLES[trace.status]}`}>
                      {STATUS_LABELS[trace.status]}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-zinc-400">{SOURCE_LABELS[trace.source_type]}</td>
                  <td className="px-4 py-3 text-zinc-400">{formatDuration(trace.duration_ms)}</td>
                  <td className="px-4 py-3 text-zinc-300 font-mono">
                    {formatCost(trace.total_cost_microdollars)}
                  </td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(trace.started_at)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
