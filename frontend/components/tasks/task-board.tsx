"use client";

import { useTasks } from "@/hooks/use-tasks";
import { Task, TaskStatus } from "@/lib/types";
import { getPriorityColor, formatDate, cn } from "@/lib/utils";

const columns: { status: TaskStatus; label: string }[] = [
  { status: "pending", label: "待分配" },
  { status: "assigned", label: "已分配" },
  { status: "in_progress", label: "进行中" },
  { status: "done", label: "已完成" },
];

function TaskCard({ task }: { task: Task }) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-md p-3 hover:border-zinc-700 transition-colors">
      <p className="text-zinc-50 text-sm font-medium line-clamp-2">{task.title}</p>
      <div className="flex items-center gap-2 mt-2">
        <span className={cn("text-xs font-medium", getPriorityColor(task.priority))}>
          {task.priority}
        </span>
        {task.dueAt && (
          <span className="text-zinc-600 text-xs">{formatDate(task.dueAt)}</span>
        )}
      </div>
    </div>
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

export function TaskBoard() {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4">
      {columns.map((col) => (
        <Column key={col.status} {...col} />
      ))}
    </div>
  );
}
