"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useTasks } from "@/hooks/use-tasks";
import type { Task, TaskPriority, TaskStatus } from "@/lib/types";
import { cn, formatDate, getPriorityColor, getStatusColor } from "@/lib/utils";

type SortKey = "id" | "title" | "status" | "priority" | "assignee_id" | "due_at" | "created_at";
type SortDirection = "asc" | "desc";

interface SortConfig {
  key: SortKey;
  direction: SortDirection;
}

const columns: Array<{ key: SortKey; label: string }> = [
  { key: "id", label: "ID" },
  { key: "title", label: "Title" },
  { key: "status", label: "Status" },
  { key: "priority", label: "Priority" },
  { key: "assignee_id", label: "Assignee" },
  { key: "due_at", label: "Due Date" },
  { key: "created_at", label: "Created At" },
];

const statusFilters: Array<{ value: "all" | TaskStatus; label: string }> = [
  { value: "all", label: "All" },
  { value: "pending", label: "Pending" },
  { value: "assigned", label: "Assigned" },
  { value: "in_progress", label: "In Progress" },
  { value: "done", label: "Done" },
  { value: "failed", label: "Failed" },
  { value: "cancelled", label: "Cancelled" },
];

const statusOrder: Record<TaskStatus, number> = {
  pending: 0,
  assigned: 1,
  in_progress: 2,
  done: 3,
  failed: 4,
  cancelled: 5,
};

const priorityOrder: Record<TaskPriority, number> = {
  low: 0,
  medium: 1,
  high: 2,
  urgent: 3,
};

function compareText(a: string, b: string) {
  return a.localeCompare(b, "zh-CN", { numeric: true, sensitivity: "base" });
}

function compareDate(a: string | null, b: string | null) {
  const aTime = a ? new Date(a).getTime() : Number.NEGATIVE_INFINITY;
  const bTime = b ? new Date(b).getTime() : Number.NEGATIVE_INFINITY;

  if (aTime === bTime) return 0;
  return aTime > bTime ? 1 : -1;
}

function compareTasks(a: Task, b: Task, key: SortKey) {
  switch (key) {
    case "id":
      return compareText(a.id, b.id);
    case "title":
      return compareText(a.title, b.title);
    case "status":
      return statusOrder[a.status] - statusOrder[b.status];
    case "priority":
      return priorityOrder[a.priority] - priorityOrder[b.priority];
    case "assignee_id":
      return compareText(a.assignee_id ?? "", b.assignee_id ?? "");
    case "due_at":
      return compareDate(a.due_at, b.due_at);
    case "created_at":
      return compareDate(a.created_at, b.created_at);
    default:
      return 0;
  }
}

export function TaskTable() {
  const [statusFilter, setStatusFilter] = useState<"all" | TaskStatus>("all");
  const [sortConfig, setSortConfig] = useState<SortConfig>({
    key: "created_at",
    direction: "desc",
  });

  const taskParams = statusFilter === "all" ? undefined : { status: statusFilter };
  const { tasks, isLoading } = useTasks(taskParams);

  const sortedTasks = useMemo(() => {
    const direction = sortConfig.direction === "asc" ? 1 : -1;
    return [...tasks].sort((a, b) => direction * compareTasks(a, b, sortConfig.key));
  }, [tasks, sortConfig]);

  const handleSort = (key: SortKey) => {
    setSortConfig((prev) => {
      if (prev.key === key) {
        return { key, direction: prev.direction === "asc" ? "desc" : "asc" };
      }

      return {
        key,
        direction: key === "created_at" || key === "due_at" ? "desc" : "asc",
      };
    });
  };

  const sortIcon = (key: SortKey) => {
    if (sortConfig.key !== key) {
      return <ArrowUpDown className="w-3.5 h-3.5 text-zinc-600" />;
    }

    return sortConfig.direction === "asc" ? (
      <ArrowUp className="w-3.5 h-3.5 text-zinc-300" />
    ) : (
      <ArrowDown className="w-3.5 h-3.5 text-zinc-300" />
    );
  };

  return (
    <div className="bg-zinc-900/80 backdrop-blur-sm border border-zinc-800/50 rounded-lg overflow-hidden">
      <div className="p-4 border-b border-zinc-800/50">
        <div className="flex flex-wrap gap-2">
          {statusFilters.map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setStatusFilter(option.value)}
              className={cn(
                "px-3 py-1.5 rounded-md text-xs border transition-all duration-200 active:scale-95",
                statusFilter === option.value
                  ? "bg-gradient-to-r from-blue-500/20 to-indigo-500/20 text-blue-400 border-blue-500/30 shadow-md shadow-blue-500/10"
                  : "bg-zinc-950/50 text-zinc-400 border-zinc-800/50 hover:text-zinc-200 hover:bg-zinc-800/50 hover:border-zinc-700/50"
              )}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800/50 bg-zinc-950/30">
              {columns.map((column) => (
                <th key={column.key} className="px-4 py-3 text-left">
                  <button
                    type="button"
                    onClick={() => handleSort(column.key)}
                    className="inline-flex items-center gap-1.5 font-medium text-zinc-400 hover:text-zinc-200 transition-all duration-200 hover:underline"
                  >
                    {column.label}
                    {sortIcon(column.key)}
                  </button>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800/50">
                  {columns.map((column, j) => (
                    <td key={column.key} className="px-4 py-3">
                      <div className={cn(
                        "h-4 rounded skeleton",
                        j === 0 ? "w-16" : j === 1 ? "w-48" : "w-20"
                      )} />
                    </td>
                  ))}
                </tr>
              ))
            ) : sortedTasks.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-10 text-center text-zinc-500">
                  <div className="flex flex-col items-center gap-2">
                    <div className="w-12 h-12 rounded-full bg-zinc-800/50 flex items-center justify-center">
                      <ArrowUpDown className="w-5 h-5 text-zinc-600" />
                    </div>
                    <p>No tasks found</p>
                  </div>
                </td>
              </tr>
            ) : (
              sortedTasks.map((task) => (
                <tr
                  key={task.id}
                  className="group border-t border-zinc-800/50 hover:bg-zinc-800/30 transition-all duration-200 cursor-pointer"
                >
                  <td className="px-4 py-3">
                    <Link
                      href={`/tasks/${task.id}`}
                      className="inline-flex items-center gap-1.5 text-blue-400 hover:text-blue-300 font-mono text-xs transition-colors"
                    >
                      <span className="w-1 h-1 rounded-full bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity" />
                      {task.id}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-zinc-100 max-w-xs truncate group-hover:text-zinc-50 transition-colors" title={task.title}>
                    {task.title}
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex items-center gap-1.5 text-zinc-300">
                      <span className={cn("w-2 h-2 rounded-full shadow-sm", getStatusColor(task.status), statusGlowClass(task.status))} />
                      <span className="capitalize">{task.status}</span>
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={cn(
                      "inline-flex items-center px-2 py-0.5 rounded text-xs font-medium transition-all",
                      task.priority === "urgent" ? "bg-red-500/20 text-red-400" :
                      task.priority === "high" ? "bg-orange-500/20 text-orange-400" :
                      task.priority === "medium" ? "bg-blue-500/20 text-blue-400" :
                      "bg-zinc-500/20 text-zinc-400"
                    )}>
                      {task.priority}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-zinc-400 font-mono text-xs group-hover:text-zinc-300 transition-colors">
                    {task.assignee_id ?? "—"}
                  </td>
                  <td className="px-4 py-3 text-zinc-400 group-hover:text-zinc-300 transition-colors">{formatDate(task.due_at)}</td>
                  <td className="px-4 py-3 text-zinc-500 group-hover:text-zinc-400 transition-colors">{formatDate(task.created_at)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function statusGlowClass(status: TaskStatus): string {
  switch (status) {
    case "pending": return "";
    case "assigned": return "status-glow-busy";
    case "in_progress": return "status-glow-busy";
    case "done": return "status-glow-online";
    case "failed": return "status-glow-offline";
    default: return "";
  }
}
