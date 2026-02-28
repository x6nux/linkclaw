import type { Agent } from "@/lib/types";

export interface Department {
  id: string;
  company_id: string;
  name: string;
  slug: string;
  description: string;
  director_agent_id: string | null;
  parent_dept_id: string | null;
  created_at: string;
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
  company_id: string;
  requester_id: string;
  approver_id: string | null;
  request_type: ApprovalRequestType;
  status: ApprovalStatus;
  payload: unknown;
  reason: string;
  decision_reason: string | null;
  created_at: string;
  decided_at: string | null;
}

export interface OrgDept {
  department: Department;
  members: Agent[];
}

export interface OrgChart {
  departments: OrgDept[];
}
