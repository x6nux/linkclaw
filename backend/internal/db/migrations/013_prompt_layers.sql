CREATE TABLE IF NOT EXISTS prompt_layers (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  VARCHAR(36) NOT NULL,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('department', 'position')),
    key         VARCHAR(100) NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_prompt_layer_unique ON prompt_layers(company_id, type, key);
CREATE INDEX IF NOT EXISTS idx_prompt_layer_company ON prompt_layers(company_id);
