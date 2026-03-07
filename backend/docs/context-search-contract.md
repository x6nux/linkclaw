# Context Search API 契约文档

> **版本**: v1.1
> **最后更新**: 2026-03-07
> **状态**: 已实现 ✅

## 实现状态

### 已完成阶段

| 阶段 | 内容 | 文件 | 状态 |
|------|------|------|------|
| 第一阶段 | 统一数据结构 | backend/internal/domain/context.go | ✅ |
| 第二阶段 | 重构 Service 层 | backend/internal/service/context_service.go | ✅ |
| 第三阶段 | 更新 API Handler | backend/internal/api/context_handler.go, backend/internal/mcp/tools_context.go | ✅ |
| 第四阶段 | 更新前端类型 | frontend/lib/types.ts | ✅ |
| 第五阶段 | 配置与灰度 | backend/internal/config/config.go | ✅ |

### 改动总结

**新增统一参数**:
- `max_results`: 最大返回数 (默认 10)
- `min_relevance`: 最低相关性阈值 0-1 (默认 0.3)
- `timeout_ms`: 超时时间 ms (普通搜索默认 30000，Agent 搜索默认 60000)
- `use_index`: 是否使用索引 (默认 true)
- `max_turns`: Agent 最大轮数 (默认 10)

**统一响应结构**:
```json
{
  "ok": true,
  "request_id": "...",
  "latency_ms": 150,
  "results": [...],           // 普通搜索
  "total": 10,
  "diagnostics": {...},
  "error": {"code": "...", "message": "..."}
}
```

**错误码 (11 种)**:
INVALID_REQUEST, UNAUTHORIZED, FORBIDDEN, NOT_FOUND, NO_DIRECTORIES, INDEX_NOT_READY, TIMEOUT, LLM_UNAVAILABLE, LLM_RATE_LIMITED, LLM_CONTEXT_TOO_LARGE, AGENT_MAX_TURNS, AGENT_TOOL_FAILED

---

## 1. 统一能力模型

### 1.1 三种搜索模式

| 模式 | 端点 | 特点 | 适用场景 |
|------|------|------|----------|
| **普通搜索** | `/context/search` | LLM 语义分析，一次性返回 | 快速查找相关代码文件 |
| **Agent 搜索** | `/context/agent-search` | 多轮工具调用，深度分析 | 复杂问题，需要理解代码逻辑 |
| **MCP 搜索** | `search_context` (Tool) | 同普通搜索，供 Agent 调用 | Agent 自主决策时调用 |

### 1.2 统一请求参数

```typescript
interface BaseSearchRequest {
  query: string;                    // 必填：自然语言查询
  directory_ids?: string[];         // 可选：限制搜索范围
  // --- 新增统一参数 ---
  max_results?: number;             // 可选：最大返回数 (默认 10)
  min_relevance?: number;           // 可选：最低相关性阈值 0-1 (默认 0.3)
  timeout_ms?: number;              // 可选：超时时间 (默认 30000)
  use_index?: boolean;              // 可选：是否使用索引 (默认 true)
}

interface AgentSearchRequest extends BaseSearchRequest {
  max_turns?: number;               // 可选：Agent 最大轮数 (默认 5)
  enable_tools?: string[];          // 可选：启用的工具列表
}
```

## 2. 统一返回结构

### 2.1 普通搜索 / MCP 搜索 响应

```typescript
interface ContextSearchResponse {
  // --- 元数据 ---
  ok: boolean;
  request_id: string;
  latency_ms: number;

  // --- 搜索结果 ---
  results: SearchResult[];
  total: number;

  // --- 诊断信息 ---
  diagnostics?: {
    directories_scanned: number;
    files_analyzed: number;
    index_used: boolean;
    fallback_reason?: string;      // 降级原因（如有）
  };

  // --- 错误处理 ---
  error?: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
}

interface SearchResult {
  file_path: string;
  language?: string;
  summary: string;                  // 文件摘要
  relevance: number;                // 0-1 相关性分数
  reason: string;                   // LLM 解释
  line_count?: number;
  directory_id: string;
  // --- 新增 ---
  content_hash?: string;            // 内容哈希（用于缓存）
  indexed_at?: string;              // 索引时间（如有）
}
```

### 2.2 Agent 搜索 响应

```typescript
interface AgentSearchResponse {
  // --- 元数据 ---
  ok: boolean;
  request_id: string;
  latency_ms: number;

  // --- 核心内容 ---
  answer: string;                   // Agent 自然语言回答
  files_read: string[];             // 已读取文件列表

  // --- 中间过程 ---
  trace?: {
    turns: number;
    tool_calls: ToolCallTrace[];
  };

  // --- 诊断信息 ---
  diagnostics?: {
    directories_scanned: number;
    files_listed: number;
    files_read: number;
  };

  // --- 错误处理 ---
  error?: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
}

interface ToolCallTrace {
  tool_name: string;
  arguments: Record<string, any>;
  result_summary: string;
  is_error: boolean;
}
```

## 3. 统一错误模型

### 3.1 错误码定义

```typescript
enum SearchErrorCode {
  // --- 通用错误 ---
  INVALID_REQUEST = "INVALID_REQUEST",     // 400: 参数校验失败
  UNAUTHORIZED = "UNAUTHORIZED",           // 401: 未认证
  FORBIDDEN = "FORBIDDEN",                 // 403: 权限不足
  NOT_FOUND = "NOT_FOUND",                 // 404: 目录不存在

  // --- 搜索特定错误 ---
  NO_DIRECTORIES = "NO_DIRECTORIES",       // 400: 没有可搜索的目录
  INDEX_NOT_READY = "INDEX_NOT_READY",     // 503: 索引未就绪
  TIMEOUT = "TIMEOUT",                     // 504: 搜索超时

  // --- LLM 相关错误 ---
  LLM_UNAVAILABLE = "LLM_UNAVAILABLE",     // 503: LLM 服务不可用
  LLM_RATE_LIMITED = "LLM_RATE_LIMITED",   // 429: LLM 限流
  LLM_CONTEXT_TOO_LARGE = "LLM_CONTEXT_TOO_LARGE", // 413: 上下文超限

  // --- Agent 特定错误 ---
  AGENT_MAX_TURNS = "AGENT_MAX_TURNS",     // 400: 超过最大轮数
  AGENT_TOOL_FAILED = "AGENT_TOOL_FAILED", // 500: 工具执行失败
}
```

### 3.2 错误响应格式

```json
{
  "ok": false,
  "request_id": "req_abc123",
  "latency_ms": 150,
  "error": {
    "code": "NO_DIRECTORIES",
    "message": "没有可搜索的目录。请先配置上下文目录。",
    "details": {
      "company_id": "xxx",
      "suggestion": "调用 POST /context/directories 创建目录"
    }
  }
}
```

## 4. 超时语义

### 4.1 默认超时时间

| 端点 | 默认超时 | 最大超时 | 行为 |
|------|----------|----------|------|
| `/context/search` | 30s | 120s | 超时返回 `TIMEOUT` 错误 |
| `/context/agent-search` | 60s | 300s | 超时终止 Agent 循环 |
| `search_context` (MCP) | 30s | 120s | 同普通搜索 |

### 4.2 超时控制实现

```go
// context_service.go
func (s *ContextService) Search(ctx context.Context, in SearchInput) (*SearchOutput, error) {
    // 从输入或默认值获取超时
    timeout := in.TimeoutMs
    if timeout <= 0 {
        timeout = 30000 // 默认 30s
    }
    if timeout > 120000 {
        timeout = 120000 // 最大 120s
    }

    // 创建带超时的上下文
    searchCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
    defer cancel()

    // 在 searchCtx 中执行搜索
    // ...
}
```

## 5. 灰度开关

### 5.1 功能开关配置

```go
// config.ContextSearchConfig
type ContextSearchConfig struct {
    // --- 索引相关 ---
    EnableIndexing       bool    // 是否启用索引功能
    IndexThreshold       int     // 文件数超过此值启用索引 (默认 100)

    // --- 降级策略 ---
    EnableLLMFallback    bool    // LLM 失败时降级到关键词匹配
    EnableIndexFallback  bool    // 索引失败时降级到全量扫描

    // --- 限流保护 ---
    MaxConcurrentSearches int    // 最大并发搜索数 (默认 10)
    RateLimitPerAgent    int     // 每个 Agent 每秒最大请求数 (默认 5)

    // --- 灰度发布 ---
    AgentSearchEnabled   bool    // 是否启用 Agent 搜索
    AgentSearchRatio     float64 // Agent 搜索灰度比例 (0.0-1.0)
}
```

### 5.2 降级策略

```
普通搜索降级链路:
1. 索引语义搜索 (首选)
   ↓ 索引未就绪
2. LLM 实时语义搜索
   ↓ LLM 失败/超时
3. 关键词匹配 (fallback)
   ↓ 无结果
4. 返回空结果集 + 诊断信息

Agent 搜索降级链路:
1. 完整 Agent 多轮搜索
   ↓ 超过最大轮数
2. 返回已收集结果 + 警告
   ↓ 工具连续失败 3 次
3. 降级到普通搜索
```

## 6. 兼容策略

### 6.1 API 版本控制

```
当前版本：v1 (无版本号，隐式)
未来扩展：/api/v1/context/v2/search

兼容原则:
- 新增可选字段：直接添加，不影响旧客户端
- 废弃字段：保留字段 + @deprecated 标记，至少兼容 2 个大版本
- 行为变更：通过新版本号或显式参数控制
```

### 6.2 字段兼容性

```typescript
// 新增字段必须满足:
// 1. 可选 (optional)
// 2. 有合理的默认值
// 3. 不改变现有字段语义

interface SearchResult {
  // 现有字段 (必须保持向后兼容)
  file_path: string;
  summary: string;
  relevance: number;
  reason: string;

  // 新增字段 (可选，有默认值)
  content_hash?: string;      // v1.1 新增
  indexed_at?: string;        // v1.1 新增
  chunks?: ChunkInfo[];       // v1.2 新增 (索引分块信息)
}
```

## 7. 改动方案

### 7.1 第一阶段：统一数据结构

**文件**: `backend/internal/domain/context.go`
- [ ] 新增 `SearchInput` 统一参数结构
- [ ] 新增 `SearchResponse` 统一响应结构
- [ ] 新增 `SearchErrorCode` 错误码枚举

### 7.2 第二阶段：重构 Service 层

**文件**: `backend/internal/service/context_service.go`
- [ ] `Search()` 方法支持统一参数
- [ ] 添加超时控制逻辑
- [ ] 添加降级策略实现

### 7.3 第三阶段：更新 API Handler

**文件**: `backend/internal/api/context_handler.go`
- [ ] `search()` 使用统一响应格式
- [ ] `agentSearch()` 使用统一响应格式
- [ ] 添加参数校验和错误处理

### 7.4 第四阶段：更新 MCP Handler

**文件**: `backend/internal/mcp/tools_context.go`
- [ ] `toolSearchContext()` 返回统一格式
- [ ] 与 HTTP 端点行为一致

### 7.5 第五阶段：更新前端类型

**文件**: `frontend/lib/types.ts`
- [ ] 更新 `ContextSearchResult`
- [ ] 新增 `SearchResponse` 类型
- [ ] 新增错误码常量

### 7.6 第六阶段：配置与灰度

**文件**: `backend/internal/config/config.go`
- [ ] 新增 `ContextSearchConfig`
- [ ] 添加灰度开关
- [ ] 添加限流配置
