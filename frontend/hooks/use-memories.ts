import useSWR, { mutate } from "swr";
import { api } from "@/lib/api";
import { Memory } from "@/lib/types";

interface MemoryListParams {
  agentId?: string;
  category?: string;
  importance?: number;
  limit?: number;
  offset?: number;
  orderBy?: string;
}

const buildUrl = (params: MemoryListParams) => {
  const q = new URLSearchParams();
  if (params.agentId) q.set("agent_id", params.agentId);
  if (params.category) q.set("category", params.category);
  if (params.importance !== undefined) q.set("importance", String(params.importance));
  if (params.limit) q.set("limit", String(params.limit));
  if (params.offset) q.set("offset", String(params.offset));
  if (params.orderBy) q.set("order_by", params.orderBy);
  const qs = q.toString();
  return `/api/v1/memories${qs ? `?${qs}` : ""}`;
};

const fetcher = (url: string) =>
  api.get<{ data: Memory[]; total: number }>(url).then((r) => r);

export function useMemories(params: MemoryListParams = {}) {
  const url = buildUrl(params);
  const { data, error, isLoading } = useSWR(url, fetcher);
  return {
    memories: data?.data ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    refresh: () => mutate(url),
  };
}

export async function createMemory(body: {
  agent_id: string;
  content: string;
  category?: string;
  tags?: string[];
  importance?: number;
  source?: string;
}) {
  return api.post<Memory>("/api/v1/memories", body);
}

export async function updateMemory(
  id: string,
  body: { content: string; category?: string; tags?: string[]; importance?: number }
) {
  return api.put<Memory>(`/api/v1/memories/${id}`, body);
}

export async function deleteMemory(id: string) {
  return api.delete(`/api/v1/memories/${id}`);
}

export async function batchDeleteMemories(ids: string[]) {
  return api.post<{ ok: boolean; deleted: number }>("/api/v1/memories/batch-delete", { ids });
}

export async function searchMemories(query: string, agentId?: string, limit?: number) {
  const q = agentId ? `?agent_id=${agentId}` : "";
  return api.post<{ data: Memory[]; total: number }>(
    `/api/v1/memories/search${q}`,
    { query, limit: limit ?? 10 }
  );
}
