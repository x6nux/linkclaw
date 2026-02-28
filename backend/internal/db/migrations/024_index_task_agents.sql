-- 024: 索引任务 Agent 授权

CREATE TABLE IF NOT EXISTS index_task_agents (
    id             VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    index_task_id  VARCHAR(36) NOT NULL,
    agent_id       VARCHAR(36) NOT NULL,
    company_id     VARCHAR(36) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS index_task_agents_task_id_idx
    ON index_task_agents(index_task_id);

CREATE INDEX IF NOT EXISTS index_task_agents_agent_id_idx
    ON index_task_agents(agent_id);

CREATE INDEX IF NOT EXISTS index_task_agents_company_id_idx
    ON index_task_agents(company_id);

CREATE INDEX IF NOT EXISTS index_task_agents_task_agent_company_idx
    ON index_task_agents(index_task_id, agent_id, company_id);
