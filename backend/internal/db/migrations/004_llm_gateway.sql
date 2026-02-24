-- 004: LLM API 网关

-- LLM Provider 配置表
CREATE TABLE IF NOT EXISTS llm_providers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    provider_type   VARCHAR(20)  NOT NULL CHECK (provider_type IN ('openai', 'anthropic')),
    base_url        VARCHAR(500) NOT NULL,
    -- API Key 使用 AES-256-GCM 加密存储（应用层加密）
    api_key_enc     TEXT NOT NULL,
    model           VARCHAR(100) NOT NULL,
    weight          INT  NOT NULL DEFAULT 100,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    -- 健康状态
    error_count     INT  NOT NULL DEFAULT 0,
    last_error_at   TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    -- 可选限额
    max_rpm         INT,   -- 每分钟请求上限（用于令牌桶）
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Token 使用明细日志
-- 费用存为微美元（microdollars = USD × 1,000,000）避免浮点精度问题
CREATE TABLE IF NOT EXISTS llm_usage_logs (
    id                        UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id                UUID    NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    provider_id               UUID    REFERENCES llm_providers(id) ON DELETE SET NULL,
    agent_id                  UUID    REFERENCES agents(id) ON DELETE SET NULL,
    request_model             VARCHAR(100) NOT NULL,
    -- Anthropic token 字段
    input_tokens              INT     NOT NULL DEFAULT 0,
    output_tokens             INT     NOT NULL DEFAULT 0,
    cache_creation_tokens     INT     NOT NULL DEFAULT 0,  -- cache write
    cache_read_tokens         INT     NOT NULL DEFAULT 0,  -- cache hit（低价）
    -- OpenAI token 字段（复用 input/output，扩展 cached）
    cached_prompt_tokens      INT     NOT NULL DEFAULT 0,
    -- 费用
    cost_microdollars         BIGINT  NOT NULL DEFAULT 0,
    -- 请求元信息
    status                    VARCHAR(20) NOT NULL DEFAULT 'success'
                                  CHECK (status IN ('success', 'error', 'timeout', 'retried')),
    latency_ms                INT,
    retry_count               SMALLINT NOT NULL DEFAULT 0,
    error_msg                 TEXT,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS llm_providers_company_id_idx ON llm_providers(company_id);
CREATE INDEX IF NOT EXISTS llm_providers_active_idx     ON llm_providers(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS llm_usage_logs_company_idx   ON llm_usage_logs(company_id);
CREATE INDEX IF NOT EXISTS llm_usage_logs_provider_idx  ON llm_usage_logs(provider_id);
CREATE INDEX IF NOT EXISTS llm_usage_logs_agent_idx     ON llm_usage_logs(agent_id);
CREATE INDEX IF NOT EXISTS llm_usage_logs_created_idx   ON llm_usage_logs(created_at DESC);

DROP TRIGGER IF EXISTS llm_providers_updated_at ON llm_providers;
CREATE TRIGGER llm_providers_updated_at
    BEFORE UPDATE ON llm_providers FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
