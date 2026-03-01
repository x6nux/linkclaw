"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { LayoutGrid, Table as TableIcon } from "lucide-react";
import { useTasks } from "@/hooks/use-tasks";
import type { Task, TaskStatus } from "@/lib/types";
import { getPriorityColor, formatDate, cn } from "@/lib/utils";
import { TaskTable } from "@/components/tasks/task-table";

const columns: Array<{ status: TaskStatus; label: string }> = [
  { status: "pending", label: "待分配" },
  { status: "assigned", label: "已分配" },
  { status: "in_progress", label: "进行中" },
  { status: "done", label: "已完成" },
];

type TaskView = "kanban" | "table";

const viewOptions: Array<{ value: TaskView; label: string; icon: typeof LayoutGrid }> = [
  { value: "kanban", label: "Kanban", icon: LayoutGrid },
  { value: "table", label: "Table", icon: TableIcon },
];

const TASK_VIEW_STORAGE_KEY = "lc_tasks_view";

function TaskCard({ task }: { task: Task }) {
  return (
    <Link href={`/tasks/${task.id}`} className="block group">
      <div className="relative bg-zinc-900/80 backdrop-blur-sm border border-zinc-800/50 rounded-md p-3 hover:border-zinc-700/50 hover:bg-zinc-900 transition-all duration-200 hover-lift overflow-hidden">
        {/* Hover gradient accent */}
        <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-indigo-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-200" />

        {/* Content */}
        <div className="relative">
          <p className="text-zinc-50 text-sm font-medium line-clamp-2 group-hover:text-blue-400 transition-colors duration-200">{task.title}</p>
          <div className="flex items-center gap-2 mt-2">
            <span className={cn(
              "inline-flex items-center px-2 py-0.5 rounded text-xs font-medium transition-all duration-200",
              task.priority === "urgent" ? "bg-red-500/20 text-red-400" :
              task.priority === "high" ? "bg-orange-500/20 text-orange-400" :
              task.priority === "medium" ? "bg-blue-500/20 text-blue-400" :
              "bg-zinc-500/20 text-zinc-400"
            )}>
              {task.priority}
            </span>
            {task.due_at && (
              <span className={cn(
                "text-xs transition-colors",
                new Date(task.due_at) < new Date() ? "text-red-400" : "text-zinc-500"
              )}>
                {formatDate(task.due_at)}
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}

function Column({ status, label }: { status: TaskStatus; label: string }) {
  const { tasks, isLoading } = useTasks({ status });

  return (
    <div className="flex-1 min-w-64">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-medium text-zinc-300">{label}</h3>
          <div className={cn(
            "w-2 h-2 rounded-full",
            status === "pending" ? "bg-zinc-500" :
            status === "assigned" ? "bg-blue-500" :
            status === "in_progress" ? "bg-amber-500" :
            "bg-emerald-500"
          )} />
        </div>
        <span className={cn(
          "text-xs px-1.5 py-0.5 rounded transition-all duration-200",
          isLoading
            ? "text-zinc-500"
            : tasks.length > 0
              ? "bg-zinc-800/50 text-zinc-300"
              : "bg-zinc-800/30 text-zinc-500"
        )}>
          {isLoading ? (
            <span className="inline-block w-3 h-3 border-2 border-zinc-600 border-t-transparent rounded-full animate-spin" />
          ) : (
            tasks.length
          )}
        </span>
      </div>
      <div className="space-y-2">
        {isLoading
          ? Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="h-20 bg-zinc-900/50 border border-zinc-800/50 rounded-md animate-pulse skeleton" />
            ))
          : tasks.map((task) => <TaskCard key={task.id} task={task} />)}
      </div>
    </div>
  );
}

function KanbanBoard() {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4">
      {columns.map((col) => (
        <Column key={col.status} {...col} />
      ))}
    </div>
  );
}

export function TaskBoard() {
  const [view, setView] = useState<TaskView>("kanban");

  useEffect(() => {
    const savedView = localStorage.getItem(TASK_VIEW_STORAGE_KEY);
    if (savedView === "kanban" || savedView === "table") {
      setView(savedView);
    }
  }, []);

  const handleViewChange = (nextView: TaskView) => {
    setView(nextView);
    localStorage.setItem(TASK_VIEW_STORAGE_KEY, nextView);
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <div className="inline-flex items-center gap-1 bg-zinc-900/50 backdrop-blur-sm border border-zinc-800/50 rounded-md p-1">
          {viewOptions.map((option) => {
            const Icon = option.icon;
            const active = view === option.value;

            return (
              <button
                key={option.value}
                type="button"
                onClick={() => handleViewChange(option.value)}
                className={cn(
                  "inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded text-xs border transition-all duration-200 active:scale-95",
                  active
                    ? "bg-gradient-to-r from-blue-500/20 to-indigo-500/20 text-blue-400 border-blue-500/30 shadow-md shadow-blue-500/10"
                    : "bg-zinc-950/50 border-transparent text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/50"
                )}
              >
                <Icon className="w-3.5 h-3.5" />
                {option.label}
              </button>
            );
          })}
        </div>
      </div>

      {view === "kanban" ? <KanbanBoard /> : <TaskTable />}
    </div>
  );
}
