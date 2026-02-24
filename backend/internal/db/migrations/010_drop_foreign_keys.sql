-- 010: 删除所有外键约束，改用应用层保证一致性
-- messages 表
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_company_id_fkey;
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_sender_id_fkey;
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_channel_id_fkey;
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_receiver_id_fkey;

-- message_reads 表
ALTER TABLE message_reads DROP CONSTRAINT IF EXISTS message_reads_message_id_fkey;
ALTER TABLE message_reads DROP CONSTRAINT IF EXISTS message_reads_agent_id_fkey;

-- tasks 表
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_company_id_fkey;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_parent_id_fkey;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_assignee_id_fkey;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_created_by_fkey;

-- agents 表
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_company_id_fkey;

-- sessions 表
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_agent_id_fkey;

-- channels 表
ALTER TABLE channels DROP CONSTRAINT IF EXISTS channels_company_id_fkey;

-- knowledge_docs 表
ALTER TABLE knowledge_docs DROP CONSTRAINT IF EXISTS knowledge_docs_company_id_fkey;
ALTER TABLE knowledge_docs DROP CONSTRAINT IF EXISTS knowledge_docs_author_id_fkey;

-- agent_deployments 表
ALTER TABLE agent_deployments DROP CONSTRAINT IF EXISTS agent_deployments_agent_id_fkey;
