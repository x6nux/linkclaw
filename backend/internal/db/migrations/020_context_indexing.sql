-- 020: 代码上下文索引

-- 1. 代码块存储（分块后的代码片段）
CREATE TABLE IF NOT EXISTS code_chunks (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    file_path       VARCHAR(512) NOT NULL,
    chunk_index     INT NOT NULL,
    content         TEXT NOT NULL,
    start_line      INT NOT NULL,
    end_line        INT NOT NULL,
    language        VARCHAR(50),
    symbols         TEXT,
    embedding_id    VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS code_chunks_company_idx
    ON code_chunks(company_id);
CREATE INDEX IF NOT EXISTS code_chunks_file_idx
    ON code_chunks(file_path);
CREATE INDEX IF NOT EXISTS code_chunks_embedding_idx
    ON code_chunks(embedding_id);

-- 2. 索引任务状态（追踪索引进度）
CREATE TABLE IF NOT EXISTS index_tasks (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    repository_url  VARCHAR(512),
    branch          VARCHAR(100) DEFAULT 'main',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    total_files     INT DEFAULT 0,
    indexed_files   INT DEFAULT 0,
    error_message   TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS index_tasks_company_idx
    ON index_tasks(company_id);
CREATE INDEX IF NOT EXISTS index_tasks_status_idx
    ON index_tasks(status);
