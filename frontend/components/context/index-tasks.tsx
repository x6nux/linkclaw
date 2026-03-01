"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Database, GitBranch, Loader2, Plus, RefreshCw, Search, RotateCcw } from "lucide-react";
import { api } from "@/lib/api";
import type { IndexStatus, IndexTask } from "@/lib/types";

type TaskListResponse =
  | IndexTask[]
  | {
      data?: IndexTask[];
      tasks?: IndexTask[];
    };

const STATUS_META: Record<IndexStatus, { label: string; className: string }> = {
  pending: {
    label: "待处理",
    className: "bg-zinc-500/10 text-zinc-400 border-zinc-500/20",
  },
  running: {
    label: "进行中",
    className: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  },
  completed: {
    label: "已完成",
    className: "bg-green-500/10 text-green-400 border-green-500/20",
  },
  failed: {
    label: "失败",
    className: "bg-red-500/10 text-red-400 border-red-500/20",
  },
};

function normalizeTasks(response: TaskListResponse): IndexTask[] {
  if (Array.isArray(response)) return response;
  if (Array.isArray(response.tasks)) return response.tasks;
  return response.data ?? [];
}

function getProgress(task: IndexTask): number {
  if (task.total_files <= 0) return 0;
  const value = Math.round((task.indexed_files / task.total_files) * 100);
  return Math.min(100, Math.max(0, value));
}

function formatDate(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleString("zh-CN", {
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return dateStr;
  }
}

interface IndexTasksProps {
  onOpenSearch?: () => void;
}

export function IndexTasks({ onOpenSearch }: IndexTasksProps) {
  const [tasks, setTasks] = useState<IndexTask[]>([]);
  const [repositoryUrl, setRepositoryUrl] = useState("");
  const [branch, setBranch] = useState("main");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [retryingId, setRetryingId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [createError, setCreateError] = useState("");
  const [success, setSuccess] = useState("");

  const hasRunningTask = useMemo(
    () => tasks.some((task) => task.status === "running"),
    [tasks]
  );

  const fetchTasks = useCallback(async (silent = false) => {
    if (silent) {
      setIsRefreshing(true);
    } else {
      setIsLoading(true);
    }
    setError("");

    try {
      const response = await api.get<TaskListResponse>("/api/v1/indexing/tasks");
      const nextTasks = normalizeTasks(response)
        .slice()
        .sort(
          (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
      setTasks(nextTasks);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载任务失败");
    } finally {
      if (silent) {
        setIsRefreshing(false);
      } else {
        setIsLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void fetchTasks();
  }, [fetchTasks]);

  useEffect(() => {
    if (!hasRunningTask) return;
    const timer = window.setInterval(() => {
      void fetchTasks(true);
    }, 5000);
    return () => window.clearInterval(timer);
  }, [fetchTasks, hasRunningTask]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError("");
    setSuccess("");

    const repo = repositoryUrl.trim();
    const nextBranch = branch.trim() || "main";

    if (!repo) {
      setCreateError("请输入仓库地址");
      return;
    }

    setIsCreating(true);
    try {
      await api.post("/api/v1/indexing/tasks", {
        repository_url: repo,
        branch: nextBranch,
      });
      setRepositoryUrl("");
      setSuccess("索引任务已创建");
      await fetchTasks(true);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "创建任务失败");
    } finally {
      setIsCreating(false);
    }
  };

  const handleRetry = async (taskId: string) => {
    setRetryingId(taskId);
    try {
      await api.post(`/api/v1/indexing/tasks/${taskId}/retry`, {});
      setSuccess("重试任务已创建");
      await fetchTasks(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "重试失败");
    } finally {
      setRetryingId(null);
    }
  };

  const inputClass =
    "w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors";

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-zinc-400" />
          <h2 className="text-sm font-medium text-zinc-200">索引任务</h2>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={onOpenSearch}
            className="px-3 py-1.5 rounded-md text-xs font-medium text-zinc-300 hover:text-zinc-50 hover:bg-zinc-800 disabled:opacity-50 transition-colors flex items-center gap-1.5 border border-zinc-700"
          >
            <Search className="w-3.5 h-3.5" />
            搜索
          </button>
          <button
            type="button"
            onClick={() => void fetchTasks(true)}
            disabled={isLoading || isRefreshing}
            className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 disabled:opacity-50 transition-colors"
            title="刷新任务"
          >
            <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin" : ""}`} />
          </button>
        </div>
      </div>

      <form onSubmit={handleCreate} className="space-y-3">
        <div>
          <label className="text-xs text-zinc-500 mb-1 block">仓库地址</label>
          <input
            type="url"
            value={repositoryUrl}
            onChange={(e) => {
              setRepositoryUrl(e.target.value);
              setCreateError("");
              setSuccess("");
            }}
            placeholder="https://github.com/owner/repo.git"
            className={inputClass}
            required
          />
        </div>
        <div>
          <label className="text-xs text-zinc-500 mb-1 block">分支</label>
          <input
            type="text"
            value={branch}
            onChange={(e) => {
              setBranch(e.target.value);
              setCreateError("");
              setSuccess("");
            }}
            placeholder="main"
            className={inputClass}
          />
        </div>

        {createError && <p className="text-red-400 text-xs">{createError}</p>}
        {success && <p className="text-green-400 text-xs">{success}</p>}

        <button
          type="submit"
          disabled={isCreating}
          className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors flex items-center justify-center gap-2"
        >
          {isCreating ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              创建中...
            </>
          ) : (
            <>
              <Plus className="w-4 h-4" />
              创建索引任务
            </>
          )}
        </button>
      </form>

      {error && <p className="text-red-400 text-xs">{error}</p>}

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800">
              <th className="px-4 py-3 text-left font-medium text-zinc-400">仓库地址</th>
              <th className="px-4 py-3 text-left font-medium text-zinc-400">分支</th>
              <th className="px-4 py-3 text-left font-medium text-zinc-400">状态</th>
              <th className="px-4 py-3 text-left font-medium text-zinc-400">进度</th>
              <th className="px-4 py-3 text-left font-medium text-zinc-400">创建时间</th>
              <th className="px-4 py-3 text-left font-medium text-zinc-400">操作</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 3 }).map((_, idx) => (
                <tr key={idx} className="border-t border-zinc-800 animate-pulse">
                  <td className="px-4 py-3"><div className="h-4 w-48 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-20 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-6 w-16 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-24 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-28 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-8 w-16 bg-zinc-800 rounded" /></td>
                </tr>
              ))
            ) : tasks.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-zinc-500">
                  暂无索引任务
                </td>
              </tr>
            ) : (
              tasks.map((task) => {
                const status = STATUS_META[task.status];
                const progress = getProgress(task);
                const totalFilesText = task.total_files > 0 ? task.total_files : "?";

                return (
                  <tr key={task.id} className="border-t border-zinc-800 hover:bg-zinc-950/50 transition-colors">
                    <td className="px-4 py-3">
                      <p className="text-sm text-zinc-100 max-w-xs truncate" title={task.repository_url}>
                        {task.repository_url}
                      </p>
                    </td>
                    <td className="px-4 py-3">
                      <div className="inline-flex items-center gap-1 text-xs text-zinc-500">
                        <GitBranch className="w-3 h-3" />
                        <span className="text-zinc-400">{task.branch}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded border text-xs ${status.className}`}>
                        {task.status === "running" ? (
                          <Loader2 className="w-3 h-3 animate-spin" />
                        ) : null}
                        {status.label}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      {task.status === "running" ? (
                        <div className="flex items-center gap-2">
                          <div className="flex-1 h-1.5 rounded-full bg-zinc-800 overflow-hidden max-w-[120px]">
                            <div
                              className="h-full bg-amber-400/80 transition-all"
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <span className="font-mono text-xs text-zinc-400">
                            {task.indexed_files}/{totalFilesText}
                          </span>
                        </div>
                      ) : (
                        <span className="font-mono text-xs text-zinc-400">
                          {task.indexed_files}/{totalFilesText}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-xs text-zinc-500">
                      {formatDate(task.created_at)}
                    </td>
                    <td className="px-4 py-3">
                      {task.status === "failed" && (
                        <button
                          type="button"
                          onClick={() => void handleRetry(task.id)}
                          disabled={retryingId === task.id}
                          className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium text-red-400 hover:text-red-300 hover:bg-red-500/10 disabled:opacity-50 transition-colors"
                          title="重试任务"
                        >
                          <RotateCcw className={`w-3 h-3 ${retryingId === task.id ? "animate-spin" : ""}`} />
                          重试
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
