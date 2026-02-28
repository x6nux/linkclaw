-- 015: 组织结构与审批流（Human-in-the-Loop）

CREATE TABLE IF NOT EXISTS departments (
    id                VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id        VARCHAR(36) NOT NULL,
    name              VARCHAR(100) NOT NULL,
    slug              VARCHAR(50) NOT NULL,
    description       TEXT DEFAULT '',
    director_agent_id VARCHAR(36),
    parent_dept_id    VARCHAR(36),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, slug)
);

CREATE INDEX IF NOT EXISTS departments_company_id_idx ON departments(company_id);

ALTER TABLE agents
  ADD COLUMN IF NOT EXISTS department_id VARCHAR(36),
  ADD COLUMN IF NOT EXISTS manager_id VARCHAR(36);

CREATE TABLE IF NOT EXISTS approval_requests (
    id              VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      VARCHAR(36) NOT NULL,
    requester_id    VARCHAR(36) NOT NULL,
    approver_id     VARCHAR(36),
    request_type    VARCHAR(50) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    payload         JSONB DEFAULT '{}'::jsonb,
    reason          TEXT DEFAULT '',
    decision_reason TEXT DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS approval_requests_company_status_idx
    ON approval_requests(company_id, status);

CREATE INDEX IF NOT EXISTS approval_requests_approver_id_idx
    ON approval_requests(approver_id);
