-- 013: 可观测性 - 调用链追踪、预算策略、错误告警、对话质量

-- 调用链根节点
CREATE TABLE IF NOT EXISTS trace_runs (
    id                      VARCHAR(36) PRIMARY KEY,
    company_id              VARCHAR(36) NOT NULL,
    root_agent_id           VARCHAR(36),
    session_id              VARCHAR(36),
    source_type             VARCHAR(20) NOT NULL CHECK (source_type IN ('mcp','http','workflow','ws')),
    source_ref_id           VARCHAR(36),
    status                  VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running','success','error','timeout')),
    started_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at                TIMESTAMPTZ,
    duration_ms             INT,
    total_cost_microdollars BIGINT NOT NULL DEFAULT 0,
    total_input_tokens      INT NOT NULL DEFAULT 0,
    total_output_tokens     INT NOT NULL DEFAULT 0,
    error_msg               TEXT,
    metadata                JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 调用链 Span
CREATE TABLE IF NOT EXISTS trace_spans (
    id                VARCHAR(36) PRIMARY KEY,
    trace_id          VARCHAR(36) NOT NULL,
    parent_span_id    VARCHAR(36),
    company_id        VARCHAR(36) NOT NULL,
    agent_id          VARCHAR(36),
    span_type         VARCHAR(30) NOT NULL CHECK (span_type IN ('mcp_tool','llm_call','workflow_node','kb_retrieval','http_call','internal')),
    name              VARCHAR(160) NOT NULL,
    provider_id       VARCHAR(36),
    request_model     VARCHAR(120),
    status            VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running','success','error','timeout')),
    started_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at          TIMESTAMPTZ,
    duration_ms       INT,
    input_tokens      INT,
    output_tokens     INT,
    cost_microdollars BIGINT,
    error_msg         TEXT,
    attributes        JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 请求/响应回放（body 使用 AES-GCM 加密）
CREATE TABLE IF NOT EXISTS trace_replays (
    id                   VARCHAR(36) PRIMARY KEY,
    company_id           VARCHAR(36) NOT NULL,
    trace_id             VARCHAR(36) NOT NULL,
    span_id              VARCHAR(36),
    request_headers      JSONB,
    response_headers     JSONB,
    request_body_enc     BYTEA,
    response_body_enc    BYTEA,
    status_code          INT,
    is_stream            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 预算策略
CREATE TABLE IF NOT EXISTS llm_budget_policies (
    id                  VARCHAR(36) PRIMARY KEY,
    company_id          VARCHAR(36) NOT NULL,
    scope_type          VARCHAR(20) NOT NULL CHECK (scope_type IN ('company','agent','provider')),
    scope_id            VARCHAR(36),
    period              VARCHAR(20) NOT NULL CHECK (period IN ('daily','weekly','monthly')),
    budget_microdollars BIGINT NOT NULL,
    warn_ratio          NUMERIC(5,2) NOT NULL DEFAULT 0.80,
    critical_ratio      NUMERIC(5,2) NOT NULL DEFAULT 0.95,
    hard_limit_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 预算告警
CREATE TABLE IF NOT EXISTS llm_budget_alerts (
    id                        VARCHAR(36) PRIMARY KEY,
    company_id                VARCHAR(36) NOT NULL,
    policy_id                 VARCHAR(36) NOT NULL,
    scope_type                VARCHAR(20) NOT NULL,
    scope_id                  VARCHAR(36),
    period_start              TIMESTAMPTZ NOT NULL,
    period_end                TIMESTAMPTZ NOT NULL,
    current_cost_microdollars BIGINT NOT NULL DEFAULT 0,
    level                     VARCHAR(20) NOT NULL CHECK (level IN ('warn','critical','blocked')),
    status                    VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open','acked','resolved')),
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 错误率告警策略
CREATE TABLE IF NOT EXISTS llm_error_alert_policies (
    id                   VARCHAR(36) PRIMARY KEY,
    company_id           VARCHAR(36) NOT NULL,
    scope_type           VARCHAR(20) NOT NULL CHECK (scope_type IN ('company','provider','model','agent')),
    scope_id             VARCHAR(36),
    window_minutes       INT NOT NULL DEFAULT 5,
    min_requests         INT NOT NULL DEFAULT 10,
    error_rate_threshold NUMERIC(5,2) NOT NULL,
    cooldown_minutes     INT NOT NULL DEFAULT 30,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 对话质量评分
CREATE TABLE IF NOT EXISTS conversation_quality_scores (
    id               VARCHAR(36) PRIMARY KEY,
    company_id       VARCHAR(36) NOT NULL,
    trace_id         VARCHAR(36) NOT NULL,
    scored_agent_id  VARCHAR(36),
    evaluator_type   VARCHAR(20) NOT NULL CHECK (evaluator_type IN ('rule','llm_judge')),
    overall_score    NUMERIC(5,2),
    dimension_scores JSONB,
    feedback         TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS trace_runs_company_id_idx                ON trace_runs(company_id);
CREATE INDEX IF NOT EXISTS trace_runs_started_at_idx                ON trace_runs(started_at DESC);
CREATE INDEX IF NOT EXISTS trace_spans_trace_id_idx                 ON trace_spans(trace_id);
CREATE INDEX IF NOT EXISTS trace_spans_company_id_idx               ON trace_spans(company_id);
CREATE INDEX IF NOT EXISTS trace_replays_trace_id_idx               ON trace_replays(trace_id);
CREATE INDEX IF NOT EXISTS llm_budget_alerts_company_id_idx         ON llm_budget_alerts(company_id);
CREATE INDEX IF NOT EXISTS conversation_quality_scores_trace_id_idx ON conversation_quality_scores(trace_id);
