"use client";

import useSWR from "swr";
import { useMemo } from "react";
import { api } from "@/lib/api";
import type { LLMProvider } from "@/lib/types";
import type { AgentImageType } from "@/components/agents/agent-step-type";

interface ProvidersResponse {
  data: LLMProvider[];
  total: number;
}

const fetcher = (url: string) => api.get<ProvidersResponse>(url);

/**
 * 获取可用模型列表，按 agentImage 过滤兼容模型。
 * nanoclaw 仅支持 Anthropic 格式，openclaw 两者皆可。
 */
export function useModels(agentImage?: AgentImageType) {
  const { data, error, isLoading } = useSWR("/api/v1/llm/providers", fetcher);

  const models = useMemo(() => {
    if (!data?.data) return [];
    const seen = new Set<string>();
    const result: string[] = [];
    for (const p of data.data) {
      if (!p.isActive) continue;
      // nanoclaw 只支持 anthropic 格式
      if (agentImage === "nanoclaw" && p.type !== "anthropic") continue;
      for (const m of p.models) {
        if (!seen.has(m)) {
          seen.add(m);
          result.push(m);
        }
      }
    }
    return result.sort();
  }, [data, agentImage]);

  return { models, isLoading, error };
}
