"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { Shell } from "@/components/layout/shell";
import { BudgetPanel } from "@/components/observability/budget-panel";
import { ErrorPolicyPanel } from "@/components/observability/error-policy-panel";
import { QualityScores } from "@/components/observability/quality-scores";
import { TraceDetail } from "@/components/observability/trace-detail";
import { TraceList } from "@/components/observability/trace-list";
import { useTraceOverview } from "@/hooks/use-observability";
import type { Agent, TraceRun } from "@/lib/types";

type TabKey = "overview" | "traces" | "budget" | "error" | "quality";

const TABS: { key: TabKey; label: string }[] = [
  { key: "overview", label: "概览" },
  { key: "traces", label: "Traces" },
  { key: "budget", label: "预算策略" },
  { key: "error", label: "错误策略" },
  { key: "quality", label: "质量评分" },
];

function formatUsd(microdollars: number) {
  return `$${(microdollars / 1_000_000).toFixed(4)}`;
}

export default function ObservabilityPage() {
  const [activeTab, setActiveTab] = useState<TabKey>("overview");
  const [isChairman, setIsChairman] = useState<boolean | null>(null);
  const [selectedTrace, setSelectedTrace] = useState<TraceRun | null>(null);
  const { overview, isLoading: overviewLoading, error: overviewError } = useTraceOverview(isChairman === true);

  useEffect(() => {
    const token = localStorage.getItem("lc_token");
    if (!token) {
      setIsChairman(false);
      return;
    }
    fetch("/api/v1/agents", { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.json())
      .then((d) => {
        const agentId = localStorage.getItem("lc_agent_id");
        const me = (d.data as Agent[])?.find((a) => a.id === agentId);
        setIsChairman(me?.roleType === "chairman");
      })
      .catch(() => setIsChairman(false));
  }, []);

  useEffect(() => {
    if (overviewError) toast.error(overviewError instanceof Error ? overviewError.message : "操作失败");
  }, [overviewError]);

  const overviewStats = useMemo(() => {
    if (!overview) return [];
    return [
      { label: "Trace 总数", value: String(overview.total), hint: "累计调用链路" },
      { label: "成功数", value: String(overview.successCount), hint: "状态为 success" },
      { label: "平均延迟", value: `${overview.avgLatencyMs.toFixed(2)}ms`, hint: "全量均值" },
      { label: "总成本", value: formatUsd(overview.totalCostMicrodollars), hint: "累计消耗" },
    ];
  }, [overview]);

  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">可观测性</h1>
          <p className="text-zinc-400 text-sm mt-1">追踪链路、预算告警、错误策略与会话质量评分</p>
        </div>

        {isChairman === null ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6">
            <div className="animate-pulse h-4 bg-zinc-800 rounded" />
          </div>
        ) : !isChairman ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-sm text-zinc-400">
            仅董事长可访问可观测性面板。
          </div>
        ) : (
          <>
            <div className="inline-flex items-center bg-zinc-900 border border-zinc-800 rounded-lg p-1">
              {TABS.map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key)}
                  className={`px-3 py-1.5 rounded-md text-sm transition-colors ${
                    activeTab === tab.key
                      ? "bg-blue-500/10 text-blue-400"
                      : "text-zinc-400 hover:text-zinc-200"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            {activeTab === "overview" && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {overviewLoading
                  ? Array.from({ length: 4 }).map((_, i) => (
                      <div key={i} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
                      </div>
                    ))
                  : overviewStats.map((item) => (
                      <div key={item.label} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
                        <div className="text-zinc-400 text-sm">{item.label}</div>
                        <div className="text-2xl font-semibold text-zinc-50 mt-1">{item.value}</div>
                        <div className="text-zinc-500 text-xs mt-1">{item.hint}</div>
                      </div>
                    ))}
              </div>
            )}

            {activeTab === "traces" && (
              <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
                <TraceList onSelect={setSelectedTrace} selectedId={selectedTrace?.id} />
                <TraceDetail traceId={selectedTrace?.id ?? null} fallbackTrace={selectedTrace} />
              </div>
            )}

            {activeTab === "budget" && <BudgetPanel />}
            {activeTab === "error" && <ErrorPolicyPanel />}
            {activeTab === "quality" && <QualityScores />}
          </>
        )}
      </div>
    </Shell>
  );
}
