-- 001: 初始化核心表

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS companies (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    system_prompt TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    role            VARCHAR(255) NOT NULL,
    -- role_type 决定权限等级: chairman(用户本人) | hr(可创建Agent) | employee(普通)
    role_type       VARCHAR(20)  NOT NULL DEFAULT 'employee'
                        CHECK (role_type IN ('chairman', 'hr', 'employee')),
    -- position 是具体职位，用于生成身份提示词和分类显示
    -- 高管: chairman|cto|cfo|coo|cmo
    -- 人力: hr_director|hr_manager
    -- 产品: product_manager|ux_designer
    -- 工程: frontend_dev|backend_dev|fullstack_dev|mobile_dev|devops|qa_engineer|data_engineer
    -- 商务: sales_manager|bd_manager|customer_success
    -- 市场: marketing_manager|content_creator
    -- 财务: accountant|financial_analyst
    position        VARCHAR(50)  NOT NULL DEFAULT 'employee',
    -- model: 部署时使用的 LLM 模型（如 glm-4.7）
    model           VARCHAR(100) NOT NULL DEFAULT '',
    -- is_human=true 时走 JWT 登录（董事长），否则走 MCP API Key
    is_human        BOOLEAN      NOT NULL DEFAULT FALSE,
    permissions     TEXT         NOT NULL DEFAULT '[]',
    persona         TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'offline'
                        CHECK (status IN ('online', 'busy', 'offline')),
    -- AI Agent 认证：api_key_hash + api_key_prefix
    api_key_hash    VARCHAR(64)  UNIQUE,
    api_key_prefix  VARCHAR(20),
    -- 人类用户（董事长）认证：password_hash
    password_hash   VARCHAR(255),
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT agents_auth_check CHECK (
        (is_human = TRUE  AND password_hash IS NOT NULL) OR
        (is_human = FALSE AND api_key_hash  IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS sessions (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id         UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    session_token    VARCHAR(255) NOT NULL UNIQUE,
    client_info      JSONB,
    protocol_version VARCHAR(50),
    connected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tasks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES tasks(id) ON DELETE SET NULL,
    title       VARCHAR(500) NOT NULL,
    description TEXT,
    priority    VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'in_progress', 'done', 'failed', 'cancelled')),
    assignee_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    created_by  UUID REFERENCES agents(id) ON DELETE SET NULL,
    due_at      TIMESTAMPTZ,
    result      TEXT,
    fail_reason TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- channels 表：管理群聊频道（DM 不入此表，动态生成）
CREATE TABLE IF NOT EXISTS channels (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,  -- e.g. "general", "engineering", "hr"
    description TEXT,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, name)
);

CREATE TABLE IF NOT EXISTS messages (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    sender_id   UUID REFERENCES agents(id) ON DELETE SET NULL,
    -- 群聊: channel_id 非空, receiver_id 为空
    -- DM:   channel_id 为空, receiver_id 非空
    channel_id  UUID REFERENCES channels(id) ON DELETE CASCADE,
    receiver_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    content     TEXT NOT NULL,
    msg_type    VARCHAR(50) NOT NULL DEFAULT 'text'
                    CHECK (msg_type IN ('text', 'system', 'task_update')),
    -- task_update 消息附带任务快照（id, title, status, progress 等）
    task_meta   JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT messages_target_check CHECK (
        (channel_id IS NOT NULL AND receiver_id IS NULL) OR
        (channel_id IS NULL     AND receiver_id IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS knowledge_docs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    title       VARCHAR(500) NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    tags        TEXT,
    author_id   UUID REFERENCES agents(id) ON DELETE SET NULL,
    search_vec  TSVECTOR,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
