"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { Link2, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { addDependency, removeDependency } from "@/hooks/use-task-detail";
import type { TaskDependency } from "@/lib/types";
import { formatDate } from "@/lib/utils";

interface TaskDependencyListProps {
  taskId: string;
  dependencies: TaskDependency[];
  isLoading: boolean;
}

export function TaskDependencyList({ taskId, dependencies, isLoading }: TaskDependencyListProps) {
  const [dependsOnId, setDependsOnId] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [removingId, setRemovingId] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const value = dependsOnId.trim();
    if (!value) return;

    setSubmitting(true);
    try {
      await addDependency(taskId, value);
      setDependsOnId("");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemove = async (depId: string) => {
    setRemovingId(depId);
    try {
      await removeDependency(taskId, depId);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setRemovingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-2">
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          value={dependsOnId}
          onChange={(event) => setDependsOnId(event.target.value)}
          placeholder="输入依赖任务 ID"
          className="flex-1 px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-md text-zinc-50 placeholder-zinc-400 text-sm focus:outline-none focus:border-zinc-700"
        />
        <button
          type="submit"
          disabled={submitting || !dependsOnId.trim()}
          className="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 text-zinc-50 rounded-md text-xs transition-colors"
        >
          {submitting ? "添加中..." : "添加"}
        </button>
      </form>

      <div className="space-y-2">
        {dependencies.length === 0 ? (
          <p className="text-zinc-400 text-sm">暂无依赖任务</p>
        ) : (
          dependencies.map((dep) => (
            <div key={dep.id} className="bg-zinc-950 border border-zinc-800 rounded-md p-3">
              <div className="flex items-center justify-between gap-2">
                <Link
                  href={`/tasks/${dep.depends_on_id}`}
                  className="inline-flex items-center gap-2 text-sm text-zinc-50 hover:text-zinc-300 transition-colors"
                >
                  <Link2 className="w-3.5 h-3.5 text-zinc-400" />
                  {dep.depends_on_id}
                </Link>
                <button
                  type="button"
                  onClick={() => handleRemove(dep.depends_on_id)}
                  disabled={removingId === dep.depends_on_id}
                  className="inline-flex items-center gap-1 px-2 py-1 text-zinc-400 hover:text-zinc-50 disabled:opacity-50 rounded transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span className="text-xs">{removingId === dep.id ? "移除中..." : "移除"}</span>
                </button>
              </div>
              <p className="text-xs text-zinc-400 mt-2">
                创建时间：{formatDate(dep.created_at)}
              </p>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
