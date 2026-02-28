# LinkClaw 待执行任务清单

> 生成时间：2026-02-27
> 后端 P0 三大功能已全部完成，剩余为前端页面 + 集成测试 + 代码审查。

## 已完成的后端 P0 功能

| 功能模块 | DB 迁移 | Repo + Service | API + MCP |
|---------|---------|---------------|-----------|
| P0-1 可观测性 | linkclaw-78n ✓ | linkclaw-hn9 ✓ | linkclaw-vye ✓ |
| P0-2 任务协作 | linkclaw-o9o ✓ | linkclaw-w1i ✓ | linkclaw-pru ✓ |
| P0-3 组织架构 | linkclaw-0l4 ✓ | linkclaw-dhn ✓ | linkclaw-2hi ✓ |

## 待执行任务（5 个，均为 P0）

### 1. linkclaw-2q1 — P0-1: 可观测性 - 前端仪表盘页面

- **优先级**：P0
- **依赖**：linkclaw-vye ✓（后端已完成，无阻塞）
- **范围**：
  - 可观测性总览仪表盘（调用 GET /observability/overview）
  - Trace 列表 + 详情（GET /observability/traces, /traces/:id）
  - 预算策略管理（CRUD /observability/budget-policies）
  - 预算告警列表（GET /observability/budget-alerts, PATCH 状态变更）
  - 错误策略管理（GET/POST /observability/error-policies）
  - 质量评分列表（GET /observability/quality-scores）
  - Trace 评分（POST /observability/traces/:id/score）

### 2. linkclaw-crj — P0-2: 任务协作 - 前端任务协作页面

- **优先级**：P0
- **依赖**：linkclaw-pru ✓（后端已完成，无阻塞）
- **范围**：
  - 任务详情页（含评论、依赖、关注者）
  - 评论 CRUD（POST/DELETE /tasks/:id/comments）
  - 依赖管理（POST/DELETE /tasks/:id/dependencies）
  - 关注/取消关注（POST/DELETE /tasks/:id/watchers）
  - 标签编辑（PUT /tasks/:id/tags）

### 3. linkclaw-6n8 — P0-3: 组织架构 - 前端组织架构与审批页面

- **优先级**：P0
- **依赖**：linkclaw-2hi ✓（后端已完成，无阻塞）
- **范围**：
  - 组织架构图（GET /organization/chart）
  - 部门管理 CRUD（/organization/departments）
  - 部门分配 Agent（POST /organization/departments/:id/assign）
  - 设置汇报关系（PUT /organization/agents/:id/manager）
  - 审批列表 + 创建（GET/POST /organization/approvals）
  - 审批/驳回（POST /organization/approvals/:id/approve|reject）

### 4. linkclaw-1jb — P0 全功能集成测试

- **优先级**：P0
- **依赖**：无（可随时开始）
- **被阻塞**：linkclaw-skv（代码审查依赖本任务）
- **范围**：
  - 后端 P0 三大功能（可观测性、任务协作、组织架构）的端到端集成测试
  - MCP 工具调用测试
  - API 端点测试
  - 权限验证测试（ChairmanOnly 等）

### 5. linkclaw-skv — P0 代码审查

- **优先级**：P0
- **依赖**：linkclaw-1jb（需集成测试通过后才可开始）
- **范围**：
  - 全量 P0 代码审查
  - 安全审计（权限、注入等）
  - 代码风格一致性
  - 性能热点排查

## 建议执行顺序

```
┌─────────────────────────────────────────┐
│  可并行：3 个前端页面 + 集成测试         │
│                                         │
│  linkclaw-2q1 (可观测性前端) ──┐         │
│  linkclaw-crj (任务协作前端) ──┼── 完成  │
│  linkclaw-6n8 (组织架构前端) ──┘   后    │
│  linkclaw-1jb (集成测试)     ──────┐     │
│                                    ↓     │
│  linkclaw-skv (代码审查)     ← 最后执行  │
└─────────────────────────────────────────┘
```

## 后端 API 端点速查

### 可观测性（ChairmanOnly）
```
GET    /api/v1/observability/overview
GET    /api/v1/observability/traces
GET    /api/v1/observability/traces/:id
POST   /api/v1/observability/traces/:id/score
GET    /api/v1/observability/budget-policies
POST   /api/v1/observability/budget-policies
PUT    /api/v1/observability/budget-policies/:id
GET    /api/v1/observability/budget-alerts
PATCH  /api/v1/observability/budget-alerts/:id
GET    /api/v1/observability/error-policies
POST   /api/v1/observability/error-policies
GET    /api/v1/observability/quality-scores
```

### 任务协作（认证用户）
```
GET    /api/v1/tasks/:id/detail
POST   /api/v1/tasks/:id/comments
DELETE /api/v1/tasks/:id/comments/:commentId
POST   /api/v1/tasks/:id/dependencies
DELETE /api/v1/tasks/:id/dependencies/:depId
POST   /api/v1/tasks/:id/watchers
DELETE /api/v1/tasks/:id/watchers
PUT    /api/v1/tasks/:id/tags
```

### 组织架构
```
GET    /api/v1/organization/approvals          (认证用户)
POST   /api/v1/organization/approvals          (认证用户)
GET    /api/v1/organization/departments        (ChairmanOnly)
POST   /api/v1/organization/departments        (ChairmanOnly)
PUT    /api/v1/organization/departments/:id    (ChairmanOnly)
DELETE /api/v1/organization/departments/:id    (ChairmanOnly)
POST   /api/v1/organization/departments/:id/assign (ChairmanOnly)
PUT    /api/v1/organization/agents/:id/manager (ChairmanOnly)
GET    /api/v1/organization/chart              (ChairmanOnly)
POST   /api/v1/organization/approvals/:id/approve  (ChairmanOnly)
POST   /api/v1/organization/approvals/:id/reject   (ChairmanOnly)
```
