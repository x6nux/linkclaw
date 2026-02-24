-- 将 model VARCHAR 改为 models TEXT（JSON 数组字符串），保留原数据
ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS models TEXT NOT NULL DEFAULT '[]';

-- 仅在旧 model 列存在时迁移数据
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'llm_providers' AND column_name = 'model'
    ) THEN
        UPDATE llm_providers SET models = '["' || model || '"]' WHERE model <> '';
        ALTER TABLE llm_providers DROP COLUMN model;
    END IF;
END $$;
