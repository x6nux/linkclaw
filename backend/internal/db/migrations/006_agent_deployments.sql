-- 006: Agent 部署记录表
CREATE TABLE IF NOT EXISTS agent_deployments (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id      UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    -- 部署类型: local_docker | ssh_docker
    deploy_type   VARCHAR(30) NOT NULL,
    -- agent 镜像类型: nanoclaw | openclaw
    agent_image   VARCHAR(30) NOT NULL DEFAULT 'nanoclaw',
    -- 本地 Docker 容器名
    container_name VARCHAR(100),
    -- SSH 配置（仅 ssh_docker 时使用）
    ssh_host      VARCHAR(255),
    ssh_port      INTEGER DEFAULT 22,
    ssh_user      VARCHAR(100),
    ssh_password  TEXT,       -- 密码认证（可选）
    ssh_key       TEXT,       -- 私钥内容（可选）
    -- 部署状态: pending | running | stopped | failed
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_msg     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_deployments_agent_id ON agent_deployments(agent_id);
