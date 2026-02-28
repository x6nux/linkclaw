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
    <Link href={`/tasks/${task.id}`} className="block">
      <div className="bg-zinc-900 border border-zinc-800 rounded-md p-3 hover:border-zinc-700 transition-colors">
        <p className="text-zinc-50 text-sm font-medium line-clamp-2">{task.title}</p>
        <div className="flex items-center gap-2 mt-2">
          <span className={cn("text-xs font-medium", getPriorityColor(task.priority))}>
            {task.priority}
          </span>
          {task.due_at && (
            <span className="text-zinc-600 text-xs">{formatDate(task.due_at)}</span>
          )}
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
        <h3 className="text-sm font-medium text-zinc-300">{label}</h3>
        <span className="text-xs text-zinc-600 bg-zinc-800 px-1.5 py-0.5 rounded">
          {isLoading ? "…" : tasks.length}
        </span>
      </div>
      <div className="space-y-2">
        {isLoading
          ? Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="h-16 bg-zinc-900 border border-zinc-800 rounded-md animate-pulse" />
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
        <div className="inline-flex items-center gap-1 bg-zinc-900 border border-zinc-800 rounded-md p-1">
          {viewOptions.map((option) => {
            const Icon = option.icon;
            const active = view === option.value;

            return (
              <button
                key={option.value}
                type="button"
                onClick={() => handleViewChange(option.value)}
                className={cn(
                  "inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded text-xs border transition-colors",
                  active
                    ? "bg-blue-500/10 text-blue-400 border-blue-500/30"
                    : "bg-zinc-950 border-transparent text-zinc-400 hover:text-zinc-200"
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
