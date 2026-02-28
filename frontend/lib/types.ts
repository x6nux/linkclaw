export interface Company {
  id: string;
  name: string;
  slug: string;
  description: string;
  system_prompt: string;
  created_at: string;
  updated_at: string;
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
  company_id: string;
  name: string;
  role: string;
  role_type: RoleType;
  position: Position;
  model: string;
  initialized: boolean;
  is_human: boolean;
  permissions: string[];
  persona: string;
  status: "online" | "busy" | "offline";
  api_key_prefix: string;
  last_seen_at: string | null;
  created_at: string;
  updated_at: string;
}

export type TaskStatus = "pending" | "assigned" | "in_progress" | "done" | "failed" | "cancelled";
export type TaskPriority = "low" | "medium" | "high" | "urgent";

export interface Task {
  id: string;
  company_id: string;
  parent_id: string | null;
  title: string;
  description: string;
  priority: TaskPriority;
  status: TaskStatus;
  assignee_id: string | null;
  created_by: string | null;
  due_at: string | null;
  result: string | null;
  fail_reason: string | null;
  subtasks?: Task[];
  created_at: string;
  updated_at: string;
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
  company_id: string;
  name: string;
  description: string;
  is_default: boolean;
  created_at: string;
}

export interface Message {
  id: string;
  company_id: string;
  sender_id: string | null;
  sender_name?: string;      // 前端 join 后填充
  channel_id: string | null;
  receiver_id: string | null;
  content: string;
  msg_type: "text" | "system" | "task_update";
  task_meta?: TaskMeta;      // task_update 时非空
  created_at: string;
}

export interface KnowledgeDoc {
  id: string;
  company_id: string;
  title: string;
  content: string;
  tags: string[];
  author_id: string | null;
  created_at: string;
  updated_at: string;
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
  company_id: string;
  agent_id: string;
  content: string;
  category: string;
  tags: string[];
  importance: MemoryImportance;
  source: MemorySource;
  access_count: number;
  last_accessed_at: string | null;
  created_at: string;
  updated_at: string;
}

// ===== LLM Gateway =====

export type ProviderType = "openai" | "anthropic";
export type ProviderStatus = "healthy" | "degraded" | "down";

export interface LLMProvider {
  id: string;
  company_id: string;
  name: string;
  type: ProviderType;
  base_url: string;
  models: string[];
  weight: number;
  is_active: boolean;
  error_count: number;
  last_error_at: string | null;
  last_used_at: string | null;
  max_rpm: number | null;
  created_at: string;
  updated_at: string;
  // 运行时字段
  status: ProviderStatus;
  api_key_prefix: string;
}

export interface LLMUsageStats {
  provider_id: string;
  provider_name: string;
  total_requests: number;
  success_requests: number;
  input_tokens: number;
  output_tokens: number;
  cache_creation_tokens: number;
  cache_read_tokens: number;
  total_cost_usd: number;
}

export interface LLMDailyUsage {
  date: string;
  input_tokens: number;
  output_tokens: number;
  cost_usd: number;
  requests: number;
}

export interface LLMRecentLog {
  id: string;
  company_id: string;
  provider_id: string | null;
  agent_id: string | null;
  request_model: string;
  input_tokens: number;
  output_tokens: number;
  cache_creation_tokens: number;
  cache_read_tokens: number;
  cost_microdollars: number;
  status: string;
  latency_ms: number | null;
  retry_count: number;
  error_msg: string | null;
  created_at: string;
}

export interface LLMStatsResponse {
  providers: LLMUsageStats[];
  daily: LLMDailyUsage[];
  recent: LLMRecentLog[];
  models: Record<ProviderType, string[]>;
}

// ===== 系统设置 =====

export interface CompanySettings {
  public_domain: string;
  agent_ws_url: string;
  mcp_public_url: string;
  nanoclaw_image: string;
  openclaw_plugin_url: string;
  embedding_base_url: string;
  embedding_model: string;
  embedding_api_key: string;
}

// ===== Context Indexing =====

export type IndexStatus = "pending" | "running" | "completed" | "failed";

export interface IndexTask {
  id: string;
  company_id: string;
  repository_url: string;
  branch: string;
  status: IndexStatus;
  total_files: number;
  indexed_files: number;
  error_message: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
}

export interface SearchResult {
  file_path: string;
  chunk_index: number;
  content: string;
  start_line: number;
  end_line: number;
  language: string;
  symbols: string;
  score: number;
}

// ===== Agent 部署 =====

export type DeployType = "local_docker" | "ssh_docker";
export type AgentImageType = "nanoclaw" | "openclaw";
export type DeployStatus = "pending" | "running" | "stopped" | "failed";

export interface AgentDeployment {
  id: string;
  agent_id: string;
  deploy_type: DeployType;
  agent_image: AgentImageType;
  container_name: string;
  ssh_host: string;
  ssh_port: number;
  ssh_user: string;
  status: DeployStatus;
  error_msg?: string;
  created_at: string;
  updated_at: string;
}

export interface DeployRequest {
  deploy_type: DeployType;
  agent_image: AgentImageType;
  api_key: string;
  ssh_host?: string;
  ssh_port?: number;
  ssh_user?: string;
  ssh_password?: string;
  ssh_key?: string;
}

// ===== 分层提示词 =====

export interface PromptLayer {
  id: string;
  company_id: string;
  type: "department" | "position";
  key: string;
  content: string;
  updated_at: string;
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
