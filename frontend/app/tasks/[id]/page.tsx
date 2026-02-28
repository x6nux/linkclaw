"use client";

import { Shell } from "@/components/layout/shell";
import { TaskDetailPanel } from "@/components/tasks/task-detail-panel";

export default function TaskDetailPage({ params }: { params: { id: string } }) {
  return (
    <Shell>
      <TaskDetailPanel taskId={params.id} />
    </Shell>
  );
}
