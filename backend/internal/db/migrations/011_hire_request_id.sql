-- 011: 招聘幂等键，防止重试导致重复创建 Agent
ALTER TABLE agents ADD COLUMN IF NOT EXISTS hire_request_id VARCHAR(100);
CREATE UNIQUE INDEX IF NOT EXISTS idx_agents_hire_request_id
    ON agents(hire_request_id) WHERE hire_request_id IS NOT NULL;
