-- 026: Partner API Keys (公司间配对密钥管理)

CREATE TABLE IF NOT EXISTS partner_api_keys (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,         -- 密钥生成方（发送方）
    partner_slug    VARCHAR(100) NOT NULL,        -- 配对公司 slug（接收方）
    partner_id      VARCHAR(36),                  -- 配对公司 ID（冗余）
    name            VARCHAR(100),                 -- 密钥名称/备注
    key_hash        VARCHAR(64) NOT NULL,         -- SHA256 hash
    key_prefix      VARCHAR(20) NOT NULL,         -- 前缀用于显示
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, partner_slug)             -- 每对公司只有一个有效密钥
);

CREATE INDEX IF NOT EXISTS idx_partner_api_keys_company ON partner_api_keys(company_id);
CREATE INDEX IF NOT EXISTS idx_partner_api_keys_partner ON partner_api_keys(partner_slug);
CREATE INDEX IF NOT EXISTS idx_partner_api_keys_lookup ON partner_api_keys(company_id, partner_slug, key_hash);
