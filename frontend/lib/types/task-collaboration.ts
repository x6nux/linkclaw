import type { Task } from "@/lib/types";

export interface TaskComment {
  id: string;
  task_id: string;
  company_id: string;
  agent_id: string;
  content: string;
  created_at: string;
}

export interface TaskDependency {
  id: string;
  task_id: string;
  depends_on_id: string;
  company_id: string;
  created_at: string;
}

export interface TaskWatcher {
  task_id: string;
  agent_id: string;
  created_at: string;
}

export interface TaskDetail extends Task {
  tags: string[];
  comments: TaskComment[];
  dependencies: TaskDependency[];
  watchers: TaskWatcher[];
}
