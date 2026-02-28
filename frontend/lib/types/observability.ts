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
  companyId: string;
  rootAgentId: string | null;
  sessionId: string | null;
  sourceType: TraceSourceType;
  sourceRefId: string | null;
  status: TraceStatus;
  startedAt: string;
  endedAt: string | null;
  durationMs: number | null;
  totalCostMicrodollars: number;
  totalInputTokens: number;
  totalOutputTokens: number;
  errorMsg: string | null;
  createdAt: string;
}

export interface TraceSpan {
  id: string;
  traceId: string;
  parentSpanId: string | null;
  companyId: string;
  agentId: string | null;
  spanType: SpanType;
  name: string;
  providerId: string | null;
  requestModel: string | null;
  status: TraceStatus;
  startedAt: string;
  endedAt: string | null;
  durationMs: number | null;
  inputTokens: number | null;
  outputTokens: number | null;
  costMicrodollars: number | null;
  errorMsg: string | null;
  attributes: unknown;
  createdAt: string;
}

export interface TraceTree {
  run: TraceRun;
  spans: TraceSpan[];
}

export interface TraceOverview {
  total: number;
  successCount: number;
  avgLatencyMs: number;
  totalCostMicrodollars: number;
}

export type BudgetScopeType = "company" | "agent" | "provider";
export type BudgetPeriod = "daily" | "weekly" | "monthly";
export type BudgetAlertLevel = "warn" | "critical" | "blocked";
export type BudgetAlertStatus = "open" | "acked" | "resolved";
export type ErrorAlertScopeType = "company" | "provider" | "model" | "agent";

export interface LLMBudgetPolicy {
  id: string;
  companyId: string;
  scopeType: BudgetScopeType;
  scopeId: string | null;
  period: BudgetPeriod;
  budgetMicrodollars: number;
  warnRatio: number;
  criticalRatio: number;
  hardLimitEnabled: boolean;
  isActive: boolean;
  createdAt: string;
}

export interface LLMBudgetAlert {
  id: string;
  companyId: string;
  policyId: string;
  scopeType: BudgetScopeType;
  scopeId: string | null;
  periodStart: string;
  periodEnd: string;
  currentCostMicrodollars: number;
  level: BudgetAlertLevel;
  status: BudgetAlertStatus;
  createdAt: string;
}

export interface LLMErrorAlertPolicy {
  id: string;
  companyId: string;
  scopeType: ErrorAlertScopeType;
  scopeId: string | null;
  windowMinutes: number;
  minRequests: number;
  errorRateThreshold: number;
  cooldownMinutes: number;
  createdAt: string;
}

export interface ConversationQualityScore {
  id: string;
  companyId: string;
  traceId: string;
  scoredAgentId: string | null;
  evaluatorType: "rule" | "llm_judge";
  overallScore: number | null;
  dimensionScores: unknown;
  feedback: string | null;
  createdAt: string;
}
