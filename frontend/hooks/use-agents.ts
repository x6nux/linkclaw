"use client";

import useSWR from "swr";
import { api } from "@/lib/api";
import { Agent, PaginatedResponse } from "@/lib/types";

const fetcher = (url: string) => api.get<PaginatedResponse<Agent>>(url);

export function useAgents() {
  const { data, error, isLoading, mutate } = useSWR("/api/v1/agents", fetcher);
  return {
    agents: data?.data ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    mutate,
  };
}

export async function updateAgent(id: string, body: { name?: string; model?: string; persona?: string }) {
  return api.patch<Agent>(`/api/v1/agents/${id}`, body);
}

export function useAgent(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/agents/${id}` : null,
    (url) => api.get<Agent>(url)
  );
  return { agent: data, isLoading, error, mutate };
}
