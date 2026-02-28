-- 016: 任务协作（评论、依赖、关注）

CREATE TABLE IF NOT EXISTS task_comments (
    id         VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id    VARCHAR(36) NOT NULL,
    company_id VARCHAR(36) NOT NULL,
    agent_id   VARCHAR(36) NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS task_comments_task_id_created_at_idx
    ON task_comments(task_id, created_at);

CREATE INDEX IF NOT EXISTS task_comments_company_id_idx
    ON task_comments(company_id);

CREATE TABLE IF NOT EXISTS task_dependencies (
    id            VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id       VARCHAR(36) NOT NULL,
    depends_on_id VARCHAR(36) NOT NULL,
    company_id    VARCHAR(36) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (task_id, depends_on_id),
    CHECK (task_id <> depends_on_id)
);

CREATE INDEX IF NOT EXISTS task_dependencies_depends_on_id_idx
    ON task_dependencies(depends_on_id);

CREATE TABLE IF NOT EXISTS task_watchers (
    task_id    VARCHAR(36) NOT NULL,
    agent_id   VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, agent_id)
);

CREATE INDEX IF NOT EXISTS task_watchers_agent_id_idx
    ON task_watchers(agent_id);

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT '[]';
