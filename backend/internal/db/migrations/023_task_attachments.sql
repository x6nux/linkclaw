-- 023: 任务附件

CREATE TABLE IF NOT EXISTS task_attachments (
    id                VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id           VARCHAR(36) NOT NULL,
    company_id        VARCHAR(36) NOT NULL,
    filename          VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_size         BIGINT NOT NULL,
    mime_type         VARCHAR(255) NOT NULL,
    storage_path      VARCHAR(1024) NOT NULL,
    uploaded_by       VARCHAR(36),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS task_attachments_task_id_idx
    ON task_attachments(task_id);

CREATE INDEX IF NOT EXISTS task_attachments_company_id_idx
    ON task_attachments(company_id);
