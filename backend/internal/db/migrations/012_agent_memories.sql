-- 012_agent_memories.sql — Agent 记忆系统
-- pgvector 扩展可选：不可用时 embedding 列用 TEXT 存储，语义搜索降级
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pgvector extension not available, embedding features will be limited';
END
$$;

-- 检测 vector 类型是否可用，决定 embedding 列类型
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vector') THEN
        EXECUTE '
            CREATE TABLE IF NOT EXISTS agent_memories (
                id               VARCHAR(36) PRIMARY KEY,
                company_id       VARCHAR(36) NOT NULL,
                agent_id         VARCHAR(36) NOT NULL,
                content          TEXT NOT NULL,
                category         VARCHAR(100) NOT NULL DEFAULT ''general'',
                tags             TEXT NOT NULL DEFAULT ''[]'',
                importance       SMALLINT NOT NULL DEFAULT 2 CHECK (importance >= 0 AND importance <= 4),
                embedding        vector(1536),
                source           VARCHAR(20) NOT NULL DEFAULT ''manual''
                                     CHECK (source IN (''conversation'',''manual'',''system'')),
                access_count     INT NOT NULL DEFAULT 0,
                last_accessed_at TIMESTAMPTZ,
                created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
            )';
    ELSE
        CREATE TABLE IF NOT EXISTS agent_memories (
            id               VARCHAR(36) PRIMARY KEY,
            company_id       VARCHAR(36) NOT NULL,
            agent_id         VARCHAR(36) NOT NULL,
            content          TEXT NOT NULL,
            category         VARCHAR(100) NOT NULL DEFAULT 'general',
            tags             TEXT NOT NULL DEFAULT '[]',
            importance       SMALLINT NOT NULL DEFAULT 2 CHECK (importance >= 0 AND importance <= 4),
            embedding        TEXT,
            source           VARCHAR(20) NOT NULL DEFAULT 'manual'
                                 CHECK (source IN ('conversation','manual','system')),
            access_count     INT NOT NULL DEFAULT 0,
            last_accessed_at TIMESTAMPTZ,
            created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_mem_company ON agent_memories(company_id);
CREATE INDEX IF NOT EXISTS idx_mem_agent ON agent_memories(agent_id);
CREATE INDEX IF NOT EXISTS idx_mem_importance ON agent_memories(importance);
CREATE INDEX IF NOT EXISTS idx_mem_created ON agent_memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mem_pending_embed ON agent_memories(created_at) WHERE embedding IS NULL;

-- HNSW 索引仅在 pgvector 可用时创建
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vector') THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_mem_embedding ON agent_memories
            USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)';
    END IF;
END
$$;

CREATE OR REPLACE FUNCTION update_mem_updated_at() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at := NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS mem_updated_at ON agent_memories;
CREATE TRIGGER mem_updated_at BEFORE UPDATE ON agent_memories
    FOR EACH ROW EXECUTE FUNCTION update_mem_updated_at();
