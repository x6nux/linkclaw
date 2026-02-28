"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Database, GitBranch, Loader2, Plus, RefreshCw } from "lucide-react";
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
  if (task.totalFiles <= 0) return 0;
  const value = Math.round((task.indexedFiles / task.totalFiles) * 100);
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

export function IndexTasks() {
  const [tasks, setTasks] = useState<IndexTask[]>([]);
  const [repositoryUrl, setRepositoryUrl] = useState("");
  const [branch, setBranch] = useState("main");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
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
          (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
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

  const inputClass =
    "w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors";

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-zinc-400" />
          <h2 className="text-sm font-medium text-zinc-200">索引任务</h2>
        </div>
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

      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, idx) => (
            <div
              key={idx}
              className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-2 animate-pulse"
            >
              <div className="h-4 w-3/4 bg-zinc-800 rounded" />
              <div className="h-3 w-1/3 bg-zinc-800 rounded" />
              <div className="h-2 w-full bg-zinc-800 rounded" />
            </div>
          ))
        ) : tasks.length === 0 ? (
          <div className="bg-zinc-950 border border-zinc-800 rounded-md p-6 text-center text-sm text-zinc-500">
            暂无索引任务
          </div>
        ) : (
          tasks.map((task) => {
            const status = STATUS_META[task.status];
            const progress = getProgress(task);
            const totalFilesText = task.totalFiles > 0 ? task.totalFiles : "?";

            return (
              <div key={task.id} className="bg-zinc-950 border border-zinc-800 rounded-md p-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="text-sm text-zinc-100 break-all">{task.repositoryUrl}</p>
                    <div className="mt-1 inline-flex items-center gap-1 text-xs text-zinc-500">
                      <GitBranch className="w-3 h-3" />
                      <span>{task.branch}</span>
                    </div>
                  </div>
                  <span
                    className={`inline-flex items-center gap-1 px-2 py-0.5 rounded border text-xs ${status.className}`}
                  >
                    {task.status === "running" ? (
                      <Loader2 className="w-3 h-3 animate-spin" />
                    ) : null}
                    {status.label}
                  </span>
                </div>

                {task.status === "running" ? (
                  <div className="mt-3">
                    <div className="flex items-center justify-between text-xs text-zinc-500">
                      <span>进度</span>
                      <span className="font-mono text-zinc-400">
                        {task.indexedFiles}/{totalFilesText} ({progress}%)
                      </span>
                    </div>
                    <div className="mt-1 h-1.5 rounded-full bg-zinc-800 overflow-hidden">
                      <div
                        className="h-full bg-amber-400/80 transition-all"
                        style={{ width: `${progress}%` }}
                      />
                    </div>
                  </div>
                ) : (
                  <p className="mt-2 text-xs text-zinc-500">
                    文件：
                    <span className="font-mono text-zinc-400">
                      {" "}
                      {task.indexedFiles}/{totalFilesText}
                    </span>
                  </p>
                )}

                {task.status === "failed" && task.errorMessage ? (
                  <p className="mt-2 text-xs text-red-400">{task.errorMessage}</p>
                ) : null}

                <div className="mt-2 text-xs text-zinc-500 flex flex-wrap gap-x-4 gap-y-1">
                  <span>创建：{formatDate(task.createdAt)}</span>
                  {task.startedAt ? <span>开始：{formatDate(task.startedAt)}</span> : null}
                  {task.completedAt ? <span>完成：{formatDate(task.completedAt)}</span> : null}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
