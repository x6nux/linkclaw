"use client";

import useSWR, { mutate as mutateCache } from "swr";
import { api } from "@/lib/api";
import type {
  BudgetAlertLevel,
  BudgetAlertStatus,
  BudgetPeriod,
  BudgetScopeType,
  ConversationQualityScore,
  ErrorAlertScopeType,
  LLMBudgetAlert,
  LLMBudgetPolicy,
  LLMErrorAlertPolicy,
  TraceOverview,
  TraceRun,
  TraceSourceType,
  TraceStatus,
  TraceTree,
} from "@/lib/types";

const OBS_BASE = "/api/v1/observability";
const OVERVIEW_KEY = `${OBS_BASE}/overview`;
const TRACES_KEY = `${OBS_BASE}/traces`;
const BUDGET_POLICIES_KEY = `${OBS_BASE}/budget-policies`;
const BUDGET_ALERTS_KEY = `${OBS_BASE}/budget-alerts`;
const ERROR_POLICIES_KEY = `${OBS_BASE}/error-policies`;
const QUALITY_SCORES_KEY = `${OBS_BASE}/quality-scores`;

interface SingleResponse<T> {
  data: T;
}

interface ListResponse<T> {
  data: T[];
  total: number;
}

function buildUrl(path: string, params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") query.set(key, String(value));
  });
  const qs = query.toString();
  return `${path}${qs ? `?${qs}` : ""}`;
}

export interface TraceQueryParams {
  status?: TraceStatus;
  sourceType?: TraceSourceType;
  limit?: number;
  offset?: number;
}

export interface BudgetAlertQueryParams {
  status?: BudgetAlertStatus;
  level?: BudgetAlertLevel;
  limit?: number;
  offset?: number;
}

export interface QualityScoreQueryParams {
  limit?: number;
  offset?: number;
}

export interface BudgetPolicyPayload {
  scope_type: BudgetScopeType;
  scope_id?: string | null;
  period: BudgetPeriod;
  budget_microdollars: number;
  warn_ratio: number;
  critical_ratio: number;
  hard_limit_enabled: boolean;
  is_active?: boolean;
}

export interface UpdateBudgetPolicyPayload {
  budget_microdollars: number;
  warn_ratio: number;
  critical_ratio: number;
  hard_limit_enabled: boolean;
  is_active: boolean;
}

export interface ErrorPolicyPayload {
  scope_type: ErrorAlertScopeType;
  scope_id?: string | null;
  window_minutes: number;
  min_requests: number;
  error_rate_threshold: number;
  cooldown_minutes: number;
}

export function useTraceOverview(enabled = true) {
  const { data, error, isLoading, mutate } = useSWR(enabled ? OVERVIEW_KEY : null, (url) =>
    api.get<SingleResponse<TraceOverview>>(url)
  );
  return { overview: data?.data ?? null, isLoading, error, mutate };
}

export function useTraces(params: TraceQueryParams = {}) {
  const key = buildUrl(TRACES_KEY, {
    status: params.status,
    source_type: params.sourceType,
    limit: params.limit,
    offset: params.offset,
  });
  const { data, error, isLoading, mutate } = useSWR(key, (url) =>
    api.get<ListResponse<TraceRun>>(url)
  );
  return { traces: data?.data ?? [], total: data?.total ?? 0, isLoading, error, mutate };
}

export function useTrace(id: string | null | undefined) {
  const { data, error, isLoading, mutate } = useSWR(id ? `${TRACES_KEY}/${id}` : null, (url) =>
    api.get<SingleResponse<TraceTree>>(url)
  );
  return { trace: data?.data ?? null, isLoading, error, mutate };
}

export function useBudgetPolicies() {
  const { data, error, isLoading, mutate } = useSWR(BUDGET_POLICIES_KEY, (url) =>
    api.get<ListResponse<LLMBudgetPolicy>>(url)
  );
  return { policies: data?.data ?? [], total: data?.total ?? 0, isLoading, error, mutate };
}

export function useBudgetAlerts(params: BudgetAlertQueryParams = {}) {
  const key = buildUrl(BUDGET_ALERTS_KEY, {
    status: params.status,
    level: params.level,
    limit: params.limit,
    offset: params.offset,
  });
  const { data, error, isLoading, mutate } = useSWR(key, (url) =>
    api.get<ListResponse<LLMBudgetAlert>>(url)
  );
  return { alerts: data?.data ?? [], total: data?.total ?? 0, isLoading, error, mutate };
}

export function useErrorPolicies() {
  const { data, error, isLoading, mutate } = useSWR(ERROR_POLICIES_KEY, (url) =>
    api.get<ListResponse<LLMErrorAlertPolicy>>(url)
  );
  return { policies: data?.data ?? [], total: data?.total ?? 0, isLoading, error, mutate };
}

export function useQualityScores(params: QualityScoreQueryParams = {}) {
  const key = buildUrl(QUALITY_SCORES_KEY, { limit: params.limit, offset: params.offset });
  const { data, error, isLoading, mutate } = useSWR(key, (url) =>
    api.get<ListResponse<ConversationQualityScore>>(url)
  );
  return { scores: data?.data ?? [], total: data?.total ?? 0, isLoading, error, mutate };
}

export async function createBudgetPolicy(body: BudgetPolicyPayload) {
  const res = await api.post<SingleResponse<LLMBudgetPolicy>>(BUDGET_POLICIES_KEY, body);
  await mutateCache(BUDGET_POLICIES_KEY);
  return res.data;
}

export async function updateBudgetPolicy(id: string, body: UpdateBudgetPolicyPayload) {
  const res = await api.put<SingleResponse<LLMBudgetPolicy>>(`${BUDGET_POLICIES_KEY}/${id}`, body);
  await mutateCache(BUDGET_POLICIES_KEY);
  return res.data;
}

export async function patchBudgetAlert(id: string, status: BudgetAlertStatus) {
  await api.patch<{ ok: true }>(`${BUDGET_ALERTS_KEY}/${id}`, { status });
  await mutateCache((key) => typeof key === "string" && key.startsWith(BUDGET_ALERTS_KEY));
}

export async function createErrorPolicy(body: ErrorPolicyPayload) {
  const res = await api.post<SingleResponse<LLMErrorAlertPolicy>>(ERROR_POLICIES_KEY, body);
  await mutateCache(ERROR_POLICIES_KEY);
  return res.data;
}

export async function scoreTrace(id: string) {
  const res = await api.post<SingleResponse<ConversationQualityScore>>(`${TRACES_KEY}/${id}/score`, {});
  await mutateCache(`${TRACES_KEY}/${id}`);
  await mutateCache((key) => typeof key === "string" && key.startsWith(QUALITY_SCORES_KEY));
  return res.data;
}
