"use client";

import useSWR, { mutate } from "swr";
import { api } from "@/lib/api";
import type {
  ApprovalRequest,
  ApprovalRequestType,
  ApprovalStatus,
  Department,
  OrgChart,
} from "@/lib/types";

const DEPARTMENTS_KEY = "/api/v1/organization/departments";
const ORG_CHART_KEY = "/api/v1/organization/chart";
const APPROVALS_KEY_PREFIX = "/api/v1/organization/approvals";

interface DepartmentListResponse {
  data: Department[];
  total: number;
}

interface OrgChartResponse {
  data: OrgChart;
}

interface ApprovalsResponse {
  data: ApprovalRequest[];
  total: number;
}

export interface DepartmentPayload {
  name: string;
  slug: string;
  description: string;
  directorAgentId?: string | null;
  parentDeptId?: string | null;
}

export interface ApprovalCreatePayload {
  requestType: ApprovalRequestType;
  payload: unknown;
  reason: string;
}

export interface ApprovalQueryParams {
  status?: ApprovalStatus;
  requestType?: ApprovalRequestType;
  limit?: number;
  offset?: number;
}

const departmentsFetcher = (url: string) => api.get<DepartmentListResponse>(url);
const chartFetcher = (url: string) => api.get<OrgChartResponse>(url);
const approvalsFetcher = (url: string) => api.get<ApprovalsResponse>(url);

function normalizeOptional(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : null;
}

function approvalsKey(params: ApprovalQueryParams = {}) {
  const query = new URLSearchParams();
  if (params.status) query.set("status", params.status);
  if (params.requestType) query.set("request_type", params.requestType);
  if (params.limit !== undefined) query.set("limit", String(params.limit));
  if (params.offset !== undefined) query.set("offset", String(params.offset));
  const qs = query.toString();
  return `${APPROVALS_KEY_PREFIX}${qs ? `?${qs}` : ""}`;
}

async function refreshOrgData() {
  await Promise.all([mutate(DEPARTMENTS_KEY), mutate(ORG_CHART_KEY)]);
}

async function refreshApprovals() {
  await mutate((key) => typeof key === "string" && key.startsWith(APPROVALS_KEY_PREFIX));
}

export function useDepartments() {
  const { data, error, isLoading, mutate: revalidate } = useSWR(DEPARTMENTS_KEY, departmentsFetcher);
  return {
    departments: data?.data ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    mutate: revalidate,
  };
}

export function useOrgChart(enabled = true) {
  const { data, error, isLoading, mutate: revalidate } = useSWR(enabled ? ORG_CHART_KEY : null, chartFetcher);
  return {
    chart: data?.data,
    isLoading,
    error,
    mutate: revalidate,
  };
}

export function useApprovals(params: ApprovalQueryParams = {}) {
  const key = approvalsKey(params);
  const { data, error, isLoading, mutate: revalidate } = useSWR(key, approvalsFetcher);
  return {
    approvals: data?.data ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    mutate: revalidate,
  };
}

export async function createDepartment(body: DepartmentPayload) {
  const res = await api.post<{ data: Department }>(DEPARTMENTS_KEY, {
    name: body.name,
    slug: body.slug,
    description: body.description,
    director_agent_id: normalizeOptional(body.directorAgentId),
    parent_dept_id: normalizeOptional(body.parentDeptId),
  });
  await refreshOrgData();
  return res.data;
}

export async function updateDepartment(id: string, body: DepartmentPayload) {
  await api.put<{ ok: true }>(`${DEPARTMENTS_KEY}/${id}`, {
    name: body.name,
    slug: body.slug,
    description: body.description,
    director_agent_id: normalizeOptional(body.directorAgentId),
    parent_dept_id: normalizeOptional(body.parentDeptId),
  });
  await refreshOrgData();
}

export async function deleteDepartment(id: string) {
  await api.delete<{ ok: true }>(`${DEPARTMENTS_KEY}/${id}`);
  await refreshOrgData();
}

export async function assignAgent(departmentId: string, agentId: string) {
  await api.post<{ ok: true }>(`${DEPARTMENTS_KEY}/${departmentId}/assign`, {
    agent_id: agentId,
  });
  await refreshOrgData();
}

export async function setManager(agentId: string, managerId: string | null) {
  await api.put<{ ok: true }>(`/api/v1/organization/agents/${agentId}/manager`, {
    manager_id: normalizeOptional(managerId),
  });
  await refreshOrgData();
}

export async function createApproval(body: ApprovalCreatePayload) {
  const res = await api.post<{ data: ApprovalRequest }>(APPROVALS_KEY_PREFIX, {
    request_type: body.requestType,
    payload: body.payload ?? {},
    reason: body.reason,
  });
  await refreshApprovals();
  return res.data;
}

export async function approveRequest(id: string, decisionReason: string) {
  await api.post<{ ok: true }>(`${APPROVALS_KEY_PREFIX}/${id}/approve`, {
    decision_reason: decisionReason,
  });
  await refreshApprovals();
}

export async function rejectRequest(id: string, decisionReason: string) {
  await api.post<{ ok: true }>(`${APPROVALS_KEY_PREFIX}/${id}/reject`, {
    decision_reason: decisionReason,
  });
  await refreshApprovals();
}
