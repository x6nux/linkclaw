import useSWR from "swr";
import { api } from "@/lib/api";
import type { PromptListResponse } from "@/lib/types";

const fetcher = (url: string) => api.get<PromptListResponse>(url);

export function usePrompts() {
  const { data, error, isLoading, mutate } = useSWR("/api/v1/prompts", fetcher);
  return { data: data ?? null, isLoading, error, mutate };
}

export async function upsertPrompt(type: string, key: string, content: string) {
  return api.put<{ ok: boolean }>(`/api/v1/prompts/${type}/${encodeURIComponent(key)}`, { content });
}

export async function deletePrompt(type: string, key: string) {
  return api.delete<{ ok: boolean }>(`/api/v1/prompts/${type}/${encodeURIComponent(key)}`);
}

export async function previewPrompt(agentId: string) {
  return api.get<{ prompt: string }>(`/api/v1/prompts/preview/${agentId}`);
}
