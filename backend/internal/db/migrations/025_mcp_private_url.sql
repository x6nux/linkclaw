-- 添加 MCP 内部通信 URL 字段
ALTER TABLE companies ADD COLUMN mcp_private_url TEXT DEFAULT NULL;
