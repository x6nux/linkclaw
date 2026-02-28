export type TraceStatus = "running" | "success" | "error" | "timeout";
export type TraceSourceType = "mcp" | "http" | "workflow" | "ws";
export type SpanType =
  | "mcp_tool"
  | "llm_call"
  | "workflow_node"
  | "kb_retrieval"
  | "http_call"
  | "internal";

export interface TraceRun {
  id: string;
  company_id: string;
  root_agent_id: string | null;
  session_id: string | null;
  source_type: TraceSourceType;
  source_ref_id: string | null;
  status: TraceStatus;
  started_at: string;
  ended_at: string | null;
  duration_ms: number | null;
  total_cost_microdollars: number;
  total_input_tokens: number;
  total_output_tokens: number;
  error_msg: string | null;
  created_at: string;
}

export interface TraceSpan {
  id: string;
  trace_id: string;
  parent_span_id: string | null;
  company_id: string;
  agent_id: string | null;
  span_type: SpanType;
  name: string;
  provider_id: string | null;
  request_model: string | null;
  status: TraceStatus;
  started_at: string;
  ended_at: string | null;
  duration_ms: number | null;
  input_tokens: number | null;
  output_tokens: number | null;
  cost_microdollars: number | null;
  error_msg: string | null;
  attributes: unknown;
  created_at: string;
}

export interface TraceTree {
  run: TraceRun;
  spans: TraceSpan[];
}

export interface TraceOverview {
  total: number;
  success_count: number;
  avg_latency_ms: number;
  total_cost_microdollars: number;
}

export type BudgetScopeType = "company" | "agent" | "provider";
export type BudgetPeriod = "daily" | "weekly" | "monthly";
export type BudgetAlertLevel = "warn" | "critical" | "blocked";
export type BudgetAlertStatus = "open" | "acked" | "resolved";
export type ErrorAlertScopeType = "company" | "provider" | "model" | "agent";

export interface LLMBudgetPolicy {
  id: string;
  company_id: string;
  scope_type: BudgetScopeType;
  scope_id: string | null;
  period: BudgetPeriod;
  budget_microdollars: number;
  warn_ratio: number;
  critical_ratio: number;
  hard_limit_enabled: boolean;
  is_active: boolean;
  created_at: string;
}

export interface LLMBudgetAlert {
  id: string;
  company_id: string;
  policy_id: string;
  scope_type: BudgetScopeType;
  scope_id: string | null;
  period_start: string;
  period_end: string;
  current_cost_microdollars: number;
  level: BudgetAlertLevel;
  status: BudgetAlertStatus;
  created_at: string;
}

export interface LLMErrorAlertPolicy {
  id: string;
  company_id: string;
  scope_type: ErrorAlertScopeType;
  scope_id: string | null;
  window_minutes: number;
  min_requests: number;
  error_rate_threshold: number;
  cooldown_minutes: number;
  created_at: string;
}

export interface ConversationQualityScore {
  id: string;
  company_id: string;
  trace_id: string;
  scored_agent_id: string | null;
  evaluator_type: "rule" | "llm_judge";
  overall_score: number | null;
  dimension_scores: unknown;
  feedback: string | null;
  created_at: string;
}
