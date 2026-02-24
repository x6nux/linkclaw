"use client";

import useSWR from "swr";
import { api } from "@/lib/api";
import { LLMStatsResponse, ProviderType } from "@/lib/types";
import { ProviderList } from "@/components/llm/provider-list";
import { ProviderStats, DailyUsageChart, RecentLogs } from "@/components/llm/stats-panel";
import { RefreshCw, AlertCircle } from "lucide-react";

export function LLMGatewayContent() {
  const fetcher = (url: string) => api.get<LLMStatsResponse>(url);
  const { data, error, isLoading } = useSWR("/api/v1/llm/stats", fetcher, {
    refreshInterval: 30_000,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-zinc-500">
        <RefreshCw className="w-4 h-4 animate-spin mr-2" />
        加载中...
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2 py-8 text-red-400 text-sm">
        <AlertCircle className="w-4 h-4" />
        加载失败：{error.message}
      </div>
    );
  }

  const models: Record<ProviderType, string[]> = data?.models ?? { anthropic: [], openai: [] };

  return (
    <div className="space-y-8">
      {/* Provider 管理 */}
      <ProviderList models={models} />

      {/* 每日用量折线图 */}
      {data?.daily && data.daily.length > 0 && (
        <section className="p-5 bg-zinc-900 border border-zinc-800 rounded-lg">
          <DailyUsageChart daily={data.daily} />
        </section>
      )}

      {/* Provider 聚合统计 */}
      {data?.providers && data.providers.length > 0 && (
        <section className="p-5 bg-zinc-900 border border-zinc-800 rounded-lg">
          <ProviderStats stats={data.providers} />
        </section>
      )}

      {/* 最近请求日志 */}
      {data?.recent && data.recent.length > 0 && (
        <section className="p-5 bg-zinc-900 border border-zinc-800 rounded-lg">
          <RecentLogs logs={data.recent} />
        </section>
      )}

      {/* 接入说明 */}
      <section className="p-5 bg-zinc-900 border border-zinc-800 rounded-lg">
        <h2 className="text-zinc-50 font-semibold mb-3">Agent 接入方式</h2>
        <div className="space-y-3 text-sm text-zinc-400">
          <p>将 Agent 的 API Base URL 替换为以下地址，即可使用内部网关：</p>
          <div className="space-y-2">
            <div className="bg-zinc-950 rounded p-3 font-mono text-xs">
              <span className="text-zinc-500"># Anthropic (Messages API)</span><br />
              <span className="text-blue-400">BASE_URL=</span>
              <span className="text-zinc-300">http://&lt;server&gt;/llm</span>
              <span className="text-zinc-500">  → POST /llm/v1/messages</span>
            </div>
            <div className="bg-zinc-950 rounded p-3 font-mono text-xs">
              <span className="text-zinc-500"># OpenAI (Chat Completions API)</span><br />
              <span className="text-blue-400">BASE_URL=</span>
              <span className="text-zinc-300">http://&lt;server&gt;/llm</span>
              <span className="text-zinc-500">  → POST /llm/v1/chat/completions</span>
            </div>
          </div>
          <p className="text-zinc-500 text-xs">
            认证：使用 Agent 自身的 Bearer API Key，无需单独配置 LLM API Key。
            网关自动从已配置的 Provider 中按权重选择，失败时自动故障转移。
          </p>
        </div>
      </section>
    </div>
  );
}
