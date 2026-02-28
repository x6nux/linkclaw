-- +migrate Up
ALTER TABLE companies ADD COLUMN embedding_base_url TEXT DEFAULT '';
ALTER TABLE companies ADD COLUMN embedding_model TEXT DEFAULT 'text-embedding-3-small';
ALTER TABLE companies ADD COLUMN embedding_api_key TEXT DEFAULT '';

-- +migrate Down
ALTER TABLE companies DROP COLUMN embedding_api_key;
ALTER TABLE companies DROP COLUMN embedding_model;
ALTER TABLE companies DROP COLUMN embedding_base_url;
