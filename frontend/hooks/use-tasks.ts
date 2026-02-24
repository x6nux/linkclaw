"use client";

import useSWR from "swr";
import { api } from "@/lib/api";
import { Task, PaginatedResponse } from "@/lib/types";

const fetcher = (url: string) => api.get<PaginatedResponse<Task>>(url);

export function useTasks(params?: { status?: string; assigneeId?: string }) {
  const query = new URLSearchParams();
  if (params?.status) query.set("status", params.status);
  if (params?.assigneeId) query.set("assignee_id", params.assigneeId);
  const qs = query.toString();

  const { data, error, isLoading, mutate } = useSWR(
    `/api/v1/tasks${qs ? `?${qs}` : ""}`,
    fetcher
  );
  return {
    tasks: data?.data ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    mutate,
  };
}

export function useTask(id: string) {
  const { data, error, isLoading, mutate } = useSWR(
    id ? `/api/v1/tasks/${id}` : null,
    (url) => api.get<Task>(url)
  );
  return { task: data, isLoading, error, mutate };
}
