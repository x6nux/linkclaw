import type { Agent } from "@/lib/types";

export interface Department {
  id: string;
  companyId: string;
  name: string;
  slug: string;
  description: string;
  directorAgentId: string | null;
  parentDeptId: string | null;
  createdAt: string;
}

export type ApprovalRequestType =
  | "hire"
  | "fire"
  | "budget_override"
  | "task_escalation"
  | "custom";

export type ApprovalStatus =
  | "pending"
  | "approved"
  | "rejected"
  | "cancelled";

export interface ApprovalRequest {
  id: string;
  companyId: string;
  requesterId: string;
  approverId: string | null;
  requestType: ApprovalRequestType;
  status: ApprovalStatus;
  payload: unknown;
  reason: string;
  decisionReason: string | null;
  createdAt: string;
  decidedAt: string | null;
}

export interface OrgDept {
  department: Department;
  members: Agent[];
}

export interface OrgChart {
  departments: OrgDept[];
}
