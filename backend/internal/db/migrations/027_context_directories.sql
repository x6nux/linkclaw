-- 027: 文件系统上下文目录

-- 上下文目录配置
CREATE TABLE IF NOT EXISTS context_directories (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    name            VARCHAR(100) NOT NULL,
    path            VARCHAR(512) NOT NULL,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    file_patterns   TEXT,                    -- 允许的文件模式，逗号分隔，如 "*.go,*.ts,*.md"
    exclude_patterns TEXT,                   -- 排除的文件模式，逗号分隔，如 "node_modules/*,.git/*"
    max_file_size   INT DEFAULT 1048576,     -- 最大文件大小（字节），默认 1MB
    last_indexed_at TIMESTAMPTZ,
    file_count      INT DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS context_directories_company_idx
    ON context_directories(company_id);
CREATE INDEX IF NOT EXISTS context_directories_active_idx
    ON context_directories(company_id, is_active);

-- 文件总结缓存（加速搜索）
CREATE TABLE IF NOT EXISTS context_file_summaries (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    directory_id    VARCHAR(36) NOT NULL,
    file_path       VARCHAR(1024) NOT NULL,
    content_hash    VARCHAR(64) NOT NULL,    -- 文件内容 SHA256，用于检测变更
    summary         TEXT NOT NULL,           -- LLM 生成的文件总结
    language        VARCHAR(50),
    line_count      INT,
    summarized_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS context_file_summaries_dir_idx
    ON context_file_summaries(directory_id);
CREATE INDEX IF NOT EXISTS context_file_summaries_file_idx
    ON context_file_summaries(directory_id, file_path);
CREATE INDEX IF NOT EXISTS context_file_summaries_hash_idx
    ON context_file_summaries(content_hash);

-- 搜索历史（用于优化和审计）
CREATE TABLE IF NOT EXISTS context_search_logs (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    agent_id        VARCHAR(36),
    query           TEXT NOT NULL,
    directory_ids   TEXT,                    -- 搜索的目录 ID 列表（JSON 数组）
    results_count   INT,
    latency_ms      INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS context_search_logs_company_idx
    ON context_search_logs(company_id);
CREATE INDEX IF NOT EXISTS context_search_logs_created_idx
    ON context_search_logs(created_at DESC);
