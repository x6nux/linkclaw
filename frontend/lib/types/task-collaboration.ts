import type { Task } from "@/lib/types";

export interface TaskComment {
  id: string;
  taskId: string;
  companyId: string;
  agentId: string;
  content: string;
  createdAt: string;
}

export interface TaskDependency {
  id: string;
  taskId: string;
  dependsOnId: string;
  companyId: string;
  createdAt: string;
}

export interface TaskWatcher {
  taskId: string;
  agentId: string;
  createdAt: string;
}

export interface TaskDetail extends Task {
  tags: string[];
  comments: TaskComment[];
  dependencies: TaskDependency[];
  watchers: TaskWatcher[];
}
