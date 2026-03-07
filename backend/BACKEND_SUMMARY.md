# LinkClaw Backend 实现总结

## 1. 技术栈与启动入口

**启动入口**: `backend/cmd/server/main.go`

### 技术栈
- **语言**: Go 1.x
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL (主存储) + Redis (缓存/MCP 会话)
- **认证**: JWT (人类用户) + API Key (AI Agent)
- **密码哈希**: bcrypt

### 启动流程
1. 加载配置 → 2. 连接数据库 → 3. 执行迁移 → 4. 初始化 Repositories (17 个) → 5. 初始化 Services (15+ 个) → 6. 启动后台服务 (Webhook/Embedding/监控) → 7. 注册路由 → 8. 启动服务器

---

## 2. 分层架构

```
┌─────────────────────────────────────────────────────┐
│                    HTTP / MCP                        │
│  (api/*_handler.go, mcp/)                           │
├─────────────────────────────────────────────────────┤
│                   Service 层                         │
│  (service/) - 业务逻辑、事务编排、事件发布          │
├─────────────────────────────────────────────────────┤
│                 Repository 层                        │
│  (repository/) - 数据库 CRUD、原始 SQL              │
├─────────────────────────────────────────────────────┤
│                   Domain 层                          │
│  (domain/) - 领域模型、常量、业务规则               │
└─────────────────────────────────────────────────────┘
```

---

## 3. 核心业务模块

| 模块 | 主要功能 |
|------|---------|
| **Agent 管理** | 用户/Agent 创建、更新、删除、状态管理 |
| **任务系统** | 任务 CRUD、评论、依赖、 watcher、附件 |
| **消息/频道** | 频道消息、WebSocket 实时推送 |
| **知识库** | 文档存储、全文搜索 |
| **记忆系统** | 向量嵌入、语义搜索 |
| **部署管理** | Docker 本地/SSH 远程部署 |
| **组织架构** | 部门管理、审批流程 |
| **可观测性** | Trace 追踪、预算告警、错误监控 |
| **LLM Gateway** | 多 Provider 代理、计费统计 |
| **Context 搜索** | 文件系统语义搜索 |
| **跨公司 Partner** | 公司间 API 配对 |

---

## 4. 用户/认证现状

### 认证方式
1. **JWT** - 人类用户 (`is_human=true`)
2. **API Key** - AI Agent (`lc_` 前缀)
3. **重置密码** - `RESET_SECRET` 环境变量

### 创建用户/成员能力
**已完整实现** (`service/agent_service.go:49-129`)

```go
type CreateAgentInput struct {
    CompanyID  string
    Name       string
    Position   domain.Position      // 28 种预设职位
    RoleType   domain.RoleType      // chairman/hr/employee
    IsHuman    bool                 // 人类 vs AI
    Password   string               // 人类密码
    SeedAPIKey string               // AI 密钥
}
```

**职位体系**: 高管 (5) + 人力资源 (2) + 产品 (2) + 工程 (7) + 商务 (3) + 市场 (2) + 财务 (3) = **24 种职位**

**权限**:
- `RoleChairman`: 所有权限
- `RoleHR`: 招聘、任务管理、知识写入
- `Director`: 任务管理、Persona/知识写入

---

## 5. 外部依赖与数据存储

### 外部依赖
| 依赖 | 用途 |
|------|------|
| PostgreSQL | 主数据库 (27 张表) |
| Redis | MCP 会话、Embedding 队列 |
| Docker/SSH | Agent 部署 |
| LLM Providers | 通过 Gateway 统一管理 |

### 数据库设计原则
- **禁止外键约束** - 关联由应用层保证
- **JSONB 存储数组** - `StringList` 类型
- **敏感数据加密** - LLM API Key 等

---

## 6. 缺口与风险

### 缺口
1. 用户管理 API 缺少批量操作、高级搜索
2. 组织架构字段存在但 API 支持有限
3. 无个人资料编辑（头像、简介等）

### 风险
1. **安全**: 敏感操作检测硬编码路径，易绕过
2. **一致性**: 无外键，删除逻辑依赖 Service 层手动覆盖
3. **并发**: 任务状态机无锁保护，可能竞态
4. **技术债务**: `main.go` 过于集中，部分 Handler 跳过 Service 层

---

## 关键文件索引

| 类别 | 文件 |
|------|------|
| **启动** | `cmd/server/main.go` |
| **路由** | `api/router.go` |
| **认证** | `api/auth_handler.go`, `api/middleware.go` |
| **核心 Service** | `service/agent_service.go`, `service/task_service.go` |
| **领域模型** | `domain/agent.go`, `domain/task.go`, `domain/company.go` |
| **MCP** | `mcp/server.go`, `mcp/handler.go` |
| **数据库** | `db/migrate.go`, `db/migrations/*.sql` |
| **配置** | `config/config.go` |
