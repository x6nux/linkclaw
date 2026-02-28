-- 019: Webhook 集成（事件订阅 + HMAC 签名 + 重试队列）

CREATE TABLE IF NOT EXISTS webhook_signing_keys (
    id         VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id VARCHAR(36) NOT NULL,
    name       VARCHAR(100) NOT NULL,
    key_type   VARCHAR(20) NOT NULL DEFAULT 'hmac_sha256' CHECK (key_type IN ('hmac_sha256','ed25519')),
    public_key TEXT,
    secret_key_enc BYTEA,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, name)
);

CREATE INDEX IF NOT EXISTS webhook_signing_keys_company_id_idx
    ON webhook_signing_keys(company_id);

CREATE TABLE IF NOT EXISTS webhooks (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    url             VARCHAR(500) NOT NULL,
    signing_key_id  VARCHAR(36),
    events          TEXT NOT NULL DEFAULT '[]',
    secret_header   VARCHAR(100) DEFAULT 'X-Webhook-Signature',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    timeout_seconds INT NOT NULL DEFAULT 10,
    retry_policy    JSONB DEFAULT '{"max_attempts":3,"backoff_base":2,"backoff_max":3600}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, name)
);

CREATE INDEX IF NOT EXISTS webhooks_company_id_idx ON webhooks(company_id);
CREATE INDEX IF NOT EXISTS webhooks_signing_key_id_idx ON webhooks(signing_key_id);

-- 注意：不使用外键约束，由应用层保证数据一致性（见 CLAUDE.md）

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id      VARCHAR(36) NOT NULL,
    company_id      VARCHAR(36) NOT NULL,
    event_type      VARCHAR(50) NOT NULL,
    payload         JSONB NOT NULL,
    signature       VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending','delivering','success','failed','retry_later')),
    http_status     INT,
    response_body   TEXT,
    attempt_count   INT NOT NULL DEFAULT 0,
    next_retry_at   TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS webhook_deliveries_webhook_id_idx ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS webhook_deliveries_status_next_retry_idx
    ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending','retry_later');
CREATE INDEX IF NOT EXISTS webhook_deliveries_company_id_idx ON webhook_deliveries(company_id);
