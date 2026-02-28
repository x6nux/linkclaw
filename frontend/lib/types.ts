export interface Company {
  id: string;
  name: string;
  slug: string;
  description: string;
  systemPrompt: string;
  createdAt: string;
  updatedAt: string;
}

export type RoleType = "chairman" | "hr" | "employee";
export type Position =
  | "chairman" | "cto" | "cfo" | "coo" | "cmo"
  | "hr_director" | "hr_manager"
  | "product_manager" | "ux_designer"
  | "frontend_dev" | "backend_dev" | "fullstack_dev" | "mobile_dev"
  | "devops" | "qa_engineer" | "data_engineer"
  | "sales_manager" | "bd_manager" | "customer_success"
  | "marketing_manager" | "content_creator"
  | "accountant" | "financial_analyst";

export const POSITION_LABELS: Record<string, string> = {
  chairman: "董事长/CEO", cto: "技术总监", cfo: "财务总监", coo: "运营总监", cmo: "市场总监",
  hr_director: "HR总监", hr_manager: "HR经理",
  product_manager: "产品经理", ux_designer: "UI/UX设计师",
  frontend_dev: "前端工程师", backend_dev: "后端工程师", fullstack_dev: "全栈工程师",
  mobile_dev: "移动端工程师", devops: "运维工程师", qa_engineer: "测试工程师", data_engineer: "数据工程师",
  sales_manager: "销售经理", bd_manager: "商务拓展", customer_success: "客户成功",
  marketing_manager: "市场经理", content_creator: "内容运营",
  accountant: "会计", financial_analyst: "财务分析师",
};

export const POSITION_DEPARTMENTS: Record<string, string> = {
  chairman: "高管", cto: "高管", cfo: "高管", coo: "高管", cmo: "高管",
  hr_director: "人力", hr_manager: "人力",
  product_manager: "产品", ux_designer: "产品",
  frontend_dev: "工程", backend_dev: "工程", fullstack_dev: "工程",
  mobile_dev: "工程", devops: "工程", qa_engineer: "工程", data_engineer: "工程",
  sales_manager: "商务", bd_manager: "商务", customer_success: "商务",
  marketing_manager: "市场", content_creator: "市场",
  accountant: "财务", financial_analyst: "财务",
};

export interface Agent {
  id: string;
  companyId: string;
  name: string;
  role: string;
  roleType: RoleType;
  position: Position;
  model: string;
  initialized: boolean;
  isHuman: boolean;
  permissions: string[];
  persona: string;
  status: "online" | "busy" | "offline";
  apiKeyPrefix: string;
  lastSeenAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export type TaskStatus = "pending" | "assigned" | "in_progress" | "done" | "failed" | "cancelled";
export type TaskPriority = "low" | "medium" | "high" | "urgent";

export interface Task {
  id: string;
  companyId: string;
  parentId: string | null;
  title: string;
  description: string;
  priority: TaskPriority;
  status: TaskStatus;
  assigneeId: string | null;
  createdBy: string | null;
  dueAt: string | null;
  result: string | null;
  failReason: string | null;
  subtasks?: Task[];
  createdAt: string;
  updatedAt: string;
}

export interface TaskMeta {
  task_id: string;
  title: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee_id?: string;
  due_at?: string;
  result?: string;
}

export interface Channel {
  id: string;
  companyId: string;
  name: string;
  description: string;
  isDefault: boolean;
  createdAt: string;
}

export interface Message {
  id: string;
  companyId: string;
  senderId: string | null;
  senderName?: string;      // 前端 join 后填充
  channelId: string | null;
  receiverId: string | null;
  content: string;
  msgType: "text" | "system" | "task_update";
  taskMeta?: TaskMeta;      // task_update 时非空
  createdAt: string;
}

export interface KnowledgeDoc {
  id: string;
  companyId: string;
  title: string;
  content: string;
  tags: string[];
  authorId: string | null;
  createdAt: string;
  updatedAt: string;
}

// ===== Agent 记忆 =====

export type MemoryImportance = 0 | 1 | 2 | 3 | 4;
export type MemorySource = "conversation" | "manual" | "system";

export const IMPORTANCE_LABELS: Record<MemoryImportance, string> = {
  0: "核心",
  1: "重要",
  2: "普通",
  3: "琐碎",
  4: "临时",
};

export const IMPORTANCE_COLORS: Record<MemoryImportance, string> = {
  0: "bg-red-500/20 text-red-400",
  1: "bg-orange-500/20 text-orange-400",
  2: "bg-blue-500/20 text-blue-400",
  3: "bg-zinc-500/20 text-zinc-400",
  4: "bg-zinc-600/20 text-zinc-500",
};

export interface Memory {
  id: string;
  companyId: string;
  agentId: string;
  content: string;
  category: string;
  tags: string[];
  importance: MemoryImportance;
  source: MemorySource;
  accessCount: number;
  lastAccessedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

// ===== LLM Gateway =====

export type ProviderType = "openai" | "anthropic";
export type ProviderStatus = "healthy" | "degraded" | "down";

export interface LLMProvider {
  id: string;
  companyId: string;
  name: string;
  type: ProviderType;
  baseUrl: string;
  models: string[];
  weight: number;
  isActive: boolean;
  errorCount: number;
  lastErrorAt: string | null;
  lastUsedAt: string | null;
  maxRpm: number | null;
  createdAt: string;
  updatedAt: string;
  // 运行时字段
  status: ProviderStatus;
  apiKeyPrefix: string;
}

export interface LLMUsageStats {
  providerId: string;
  providerName: string;
  totalRequests: number;
  successRequests: number;
  inputTokens: number;
  outputTokens: number;
  cacheCreationTokens: number;
  cacheReadTokens: number;
  totalCostUsd: number;
}

export interface LLMDailyUsage {
  date: string;
  inputTokens: number;
  outputTokens: number;
  costUsd: number;
  requests: number;
}

export interface LLMRecentLog {
  id: string;
  companyId: string;
  providerId: string | null;
  agentId: string | null;
  requestModel: string;
  inputTokens: number;
  outputTokens: number;
  cacheCreationTokens: number;
  cacheReadTokens: number;
  costMicrodollars: number;
  status: string;
  latencyMs: number | null;
  retryCount: number;
  errorMsg: string | null;
  createdAt: string;
}

export interface LLMStatsResponse {
  providers: LLMUsageStats[];
  daily: LLMDailyUsage[];
  recent: LLMRecentLog[];
  models: Record<ProviderType, string[]>;
}

// ===== 系统设置 =====

export interface CompanySettings {
  publicDomain: string;
  agentWsUrl: string;
  mcpPublicUrl: string;
  nanoclawImage: string;
  openclawPluginUrl: string;
}

// ===== Agent 部署 =====

export type DeployType = "local_docker" | "ssh_docker";
export type AgentImageType = "nanoclaw" | "openclaw";
export type DeployStatus = "pending" | "running" | "stopped" | "failed";

export interface AgentDeployment {
  id: string;
  agentId: string;
  deployType: DeployType;
  agentImage: AgentImageType;
  containerName: string;
  sshHost: string;
  sshPort: number;
  sshUser: string;
  status: DeployStatus;
  errorMsg?: string;
  createdAt: string;
  updatedAt: string;
}

export interface DeployRequest {
  deployType: DeployType;
  agentImage: AgentImageType;
  apiKey: string;
  sshHost?: string;
  sshPort?: number;
  sshUser?: string;
  sshPassword?: string;
  sshKey?: string;
}

// ===== 分层提示词 =====

export interface PromptLayer {
  id: string;
  companyId: string;
  type: "department" | "position";
  key: string;
  content: string;
  updatedAt: string;
}

export interface PromptAgentBrief {
  id: string;
  name: string;
  position: string;
  persona: string;
}

export interface PromptListResponse {
  global: string;
  departments: Record<string, string>;
  positions: Record<string, string>;
  agents: PromptAgentBrief[];
}

// ===== 部门和职位常量（用于提示词导航） =====

export const DEPARTMENTS = ["工程", "产品", "人力资源", "市场", "商务", "财务"] as const;

export const DEPARTMENT_POSITIONS: Record<string, Position[]> = {
  高管: ["chairman", "cto", "cfo", "coo", "cmo"],
  人力资源: ["hr_director", "hr_manager"],
  产品: ["product_manager", "ux_designer"],
  工程: ["frontend_dev", "backend_dev", "fullstack_dev", "mobile_dev", "devops", "qa_engineer", "data_engineer"],
  商务: ["sales_manager", "bd_manager", "customer_success"],
  市场: ["marketing_manager", "content_creator"],
  财务: ["accountant", "financial_analyst"],
};

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  cursor?: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
}

export * from "./types/organization";
export * from "./types/task-collaboration";
export * from "./types/observability";
