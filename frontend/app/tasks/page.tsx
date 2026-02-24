import { Shell } from "@/components/layout/shell";
import { TaskBoard } from "@/components/tasks/task-board";

export default function TasksPage() {
  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">任务看板</h1>
          <p className="text-zinc-400 text-sm mt-1">追踪所有 Agent 的任务状态</p>
        </div>
        <TaskBoard />
      </div>
    </Shell>
  );
}
