"use client";

import useSWR, { mutate } from "swr";
import { api } from "@/lib/api";
import type { TaskComment, TaskDependency, TaskDetail } from "@/lib/types";

const getTaskDetailKey = (taskId: string) => `/api/v1/tasks/${taskId}/detail`;

export function useTaskDetail(taskId: string) {
  const { data, error, isLoading, mutate: mutateTask } = useSWR(
    taskId ? getTaskDetailKey(taskId) : null,
    (url) => api.get<TaskDetail>(url)
  );

  return { task: data, isLoading, error, mutate: mutateTask };
}

async function refreshTaskDetail(taskId: string) {
  await mutate(getTaskDetailKey(taskId));
}

export async function addComment(taskId: string, content: string) {
  const res = await api.post<TaskComment>(`/api/v1/tasks/${taskId}/comments`, { content });
  await refreshTaskDetail(taskId);
  return res;
}

export async function deleteComment(taskId: string, commentId: string) {
  const res = await api.delete<{ ok: true }>(`/api/v1/tasks/${taskId}/comments/${commentId}`);
  await refreshTaskDetail(taskId);
  return res;
}

export async function addDependency(taskId: string, dependsOnId: string) {
  const res = await api.post<TaskDependency>(`/api/v1/tasks/${taskId}/dependencies`, {
    depends_on_id: dependsOnId,
  });
  await refreshTaskDetail(taskId);
  return res;
}

export async function removeDependency(taskId: string, depId: string) {
  const res = await api.delete<{ ok: true }>(`/api/v1/tasks/${taskId}/dependencies/${depId}`);
  await refreshTaskDetail(taskId);
  return res;
}

export async function addWatcher(taskId: string) {
  const res = await api.post<{ ok: true }>(`/api/v1/tasks/${taskId}/watchers`, {});
  await refreshTaskDetail(taskId);
  return res;
}

export async function removeWatcher(taskId: string) {
  const res = await api.delete<{ ok: true }>(`/api/v1/tasks/${taskId}/watchers`);
  await refreshTaskDetail(taskId);
  return res;
}

export async function updateTags(taskId: string, tags: string[]) {
  const res = await api.put<{ ok: true }>(`/api/v1/tasks/${taskId}/tags`, { tags });
  await refreshTaskDetail(taskId);
  return res;
}
