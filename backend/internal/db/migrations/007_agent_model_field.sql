-- 007: Agent 增加 model 字段（部署时使用的 LLM 模型）
ALTER TABLE agents ADD COLUMN IF NOT EXISTS model VARCHAR(100) NOT NULL DEFAULT '';
