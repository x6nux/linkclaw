-- 017: Agent Persona 自我学习与优化

-- 1. Persona 优化建议（AI 分析生成的改进建议）
CREATE TABLE IF NOT EXISTS persona_optimization_suggestions (
    id               VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       VARCHAR(36) NOT NULL,
    agent_id         VARCHAR(36) NOT NULL,
    suggestion_type  VARCHAR(50) NOT NULL, -- 'tone', 'structure', 'content', 'length'
    current_persona  TEXT NOT NULL,        -- 当前 persona 内容
    suggested_change TEXT NOT NULL,        -- 建议的变更
    reason           TEXT,                 -- 优化理由
    confidence       FLOAT DEFAULT 0.0,    -- AI 置信度 0-1
    priority         VARCHAR(20) NOT NULL DEFAULT 'medium',
    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'approved', 'rejected', 'applied')),
    applied_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS persona_suggestions_agent_id_idx
    ON persona_optimization_suggestions(agent_id);
CREATE INDEX IF NOT EXISTS persona_suggestions_company_id_idx
    ON persona_optimization_suggestions(company_id);

-- 2. Persona 变更历史（审计追踪）
CREATE TABLE IF NOT EXISTS persona_history (
    id            VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id    VARCHAR(36) NOT NULL,
    agent_id      VARCHAR(36) NOT NULL,
    old_persona   TEXT,
    new_persona   TEXT,
    change_reason TEXT,
    suggestion_id VARCHAR(36),
    changed_by    VARCHAR(36),          -- 操作者 ID（AI 变更为 NULL）
    change_type   VARCHAR(50) NOT NULL, -- 'manual', 'ai_suggested', 'auto_optimized'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS persona_history_agent_id_idx
    ON persona_history(agent_id);
CREATE INDEX IF NOT EXISTS persona_history_company_id_idx
    ON persona_history(company_id, created_at DESC);

-- 3. A/B 测试 Persona（对比效果）
CREATE TABLE IF NOT EXISTS ab_test_personas (
    id                       VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id               VARCHAR(36) NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    -- 原始 Agent（对照组）
    control_agent_id         VARCHAR(36) NOT NULL,
    control_persona          TEXT NOT NULL,
    -- 变异 Agent（实验组）
    variant_agent_id         VARCHAR(36) NOT NULL,
    variant_persona          TEXT NOT NULL,
    -- 测试状态
    status                   VARCHAR(20) NOT NULL DEFAULT 'running'
                             CHECK (status IN ('running', 'paused', 'completed', 'stopped')),
    start_time               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time                 TIMESTAMPTZ,
    -- 结果指标
    control_tasks_completed  INT DEFAULT 0,
    variant_tasks_completed  INT DEFAULT 0,
    control_avg_duration     INT, -- 秒
    variant_avg_duration     INT, -- 秒
    winner                   VARCHAR(20), -- 'control', 'variant', 'inconclusive'
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ab_test_company_id_idx
    ON ab_test_personas(company_id);
