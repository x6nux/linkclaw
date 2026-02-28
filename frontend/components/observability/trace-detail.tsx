"use client";

import { useEffect, useMemo, useState } from "react";
import { Sparkles } from "lucide-react";
import { toast } from "sonner";
import { scoreTrace, useTrace } from "@/hooks/use-observability";
import { formatDate } from "@/lib/utils";
import type { SpanType, TraceRun, TraceSpan, TraceStatus } from "@/lib/types";

interface TraceDetailProps {
  traceId: string | null;
  fallbackTrace?: TraceRun | null;
}

interface SpanRow {
  span: TraceSpan;
  depth: number;
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

const TYPE_LABELS: Record<SpanType, string> = {
  mcp_tool: "MCP 工具",
  llm_call: "LLM 调用",
  workflow_node: "工作流节点",
  kb_retrieval: "知识检索",
  http_call: "HTTP 调用",
  internal: "内部逻辑",
};

function formatDuration(ms: number | null) {
  if (ms === null || ms === undefined) return "—";
  if (ms > 1000) return `${(ms / 1000).toFixed(2)}s`;
  return `${ms}ms`;
}

function formatCost(microdollars: number) {
  return `$${(microdollars / 1_000_000).toFixed(4)}`;
}

function buildSpanRows(spans: TraceSpan[]): SpanRow[] {
  if (!spans.length) return [];

  const idSet = new Set(spans.map((span) => span.id));
  const children = new Map<string, TraceSpan[]>();

  for (const span of spans) {
    const parentId = span.parentSpanId && idSet.has(span.parentSpanId) ? span.parentSpanId : "__root__";
    const group = children.get(parentId) ?? [];
    group.push(span);
    children.set(parentId, group);
  }

  for (const group of children.values()) {
    group.sort((a, b) => new Date(a.startedAt).getTime() - new Date(b.startedAt).getTime());
  }

  const rows: SpanRow[] = [];
  const walk = (parentId: string, depth: number) => {
    const group = children.get(parentId) ?? [];
    for (const span of group) {
      rows.push({ span, depth });
      walk(span.id, depth + 1);
    }
  };

  walk("__root__", 0);
  return rows;
}

export function TraceDetail({ traceId, fallbackTrace }: TraceDetailProps) {
  const { trace, isLoading, error } = useTrace(traceId);
  const [isScoring, setIsScoring] = useState(false);

  const run = trace?.run ?? fallbackTrace ?? null;
  const spanRows = useMemo(() => buildSpanRows(trace?.spans ?? []), [trace?.spans]);

  useEffect(() => {
    if (error) toast.error(error instanceof Error ? error.message : "操作失败");
  }, [error]);

  async function handleScoreTrace() {
    if (!traceId) return;
    setIsScoring(true);
    try {
      await scoreTrace(traceId);
      toast.success("质量评分已生成");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setIsScoring(false);
    }
  }

  if (!traceId) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-zinc-500 text-sm">
        请选择左侧 Trace 查看详情
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-3">
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
      </div>
    );
  }

  if (!run) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-zinc-500 text-sm">
        未找到该 Trace 的详情数据
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div className="p-4 border-b border-zinc-800">
        <div className="flex items-center gap-2">
          <h3 className="text-zinc-50 font-medium">Trace 详情</h3>
          <span className={`inline-flex px-2 py-0.5 rounded border text-xs ${STATUS_STYLES[run.status]}`}>
            {STATUS_LABELS[run.status]}
          </span>
          <button
            onClick={handleScoreTrace}
            disabled={isScoring}
            className="ml-auto inline-flex items-center gap-1 px-2.5 py-1 rounded border border-zinc-700 text-xs text-zinc-300 hover:text-zinc-50 hover:bg-zinc-800 disabled:opacity-50"
          >
            <Sparkles className="w-3 h-3" />
            {isScoring ? "评分中…" : "生成质量评分"}
          </button>
        </div>
        <p className="mt-1 text-xs text-zinc-500 font-mono">{run.id}</p>
      </div>

      <div className="p-4 grid grid-cols-2 lg:grid-cols-4 gap-3 border-b border-zinc-800">
        <div>
          <div className="text-xs text-zinc-500">总成本</div>
          <div className="text-sm text-zinc-200 font-mono">{formatCost(run.totalCostMicrodollars)}</div>
        </div>
        <div>
          <div className="text-xs text-zinc-500">总 Tokens</div>
          <div className="text-sm text-zinc-200 font-mono">
            {run.totalInputTokens}↑ / {run.totalOutputTokens}↓
          </div>
        </div>
        <div>
          <div className="text-xs text-zinc-500">总耗时</div>
          <div className="text-sm text-zinc-200">{formatDuration(run.durationMs)}</div>
        </div>
        <div>
          <div className="text-xs text-zinc-500">开始时间</div>
          <div className="text-sm text-zinc-200">{formatDate(run.startedAt)}</div>
        </div>
      </div>

      <div>
        {spanRows.length === 0 ? (
          <div className="p-6 text-sm text-zinc-500">暂无 Span 数据</div>
        ) : (
          spanRows.map(({ span, depth }) => (
            <div key={span.id} className="border-t border-zinc-800 px-4 py-3">
              <div className="flex items-start gap-3">
                <div className="min-w-0 flex-1" style={{ paddingLeft: `${depth * 16}px` }}>
                  <div className="flex items-center gap-2 min-w-0">
                    <span className="inline-flex px-2 py-0.5 rounded border border-zinc-700 text-xs text-zinc-400">
                      {TYPE_LABELS[span.spanType]}
                    </span>
                    <span className="text-sm text-zinc-200 truncate">{span.name}</span>
                    <span className={`inline-flex px-2 py-0.5 rounded border text-xs ${STATUS_STYLES[span.status]}`}>
                      {STATUS_LABELS[span.status]}
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-zinc-500 font-mono">
                    {span.id.slice(0, 8)}… · {formatDate(span.startedAt)}
                  </div>
                </div>
                <div className="text-right text-xs text-zinc-400 min-w-28">
                  <div>{formatDuration(span.durationMs)}</div>
                  {span.spanType === "llm_call" &&
                    span.costMicrodollars !== null &&
                    span.costMicrodollars !== undefined && (
                      <div className="mt-1 font-mono text-zinc-300">{formatCost(span.costMicrodollars)}</div>
                    )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
