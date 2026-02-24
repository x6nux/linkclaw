-- 002: 知识库全文搜索

-- 自动更新 search_vec 的触发器
CREATE OR REPLACE FUNCTION knowledge_docs_search_vec_update()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vec := to_tsvector('simple', 
        COALESCE(NEW.title, '') || ' ' || 
        COALESCE(array_to_string(NEW.tags, ' '), '') || ' ' || 
        COALESCE(NEW.content, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tsvectorupdate ON knowledge_docs;
CREATE TRIGGER tsvectorupdate
    BEFORE INSERT OR UPDATE ON knowledge_docs
    FOR EACH ROW EXECUTE FUNCTION knowledge_docs_search_vec_update();

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS knowledge_docs_search_idx ON knowledge_docs USING GIN(search_vec);
