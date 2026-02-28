-- 018: 审计日志与敏感访问追踪

CREATE TABLE IF NOT EXISTS audit_logs (
    id            VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id    VARCHAR(36) NOT NULL,
    agent_id      VARCHAR(36) NOT NULL,
    agent_name    VARCHAR(100) NOT NULL,
    action        VARCHAR(50) NOT NULL CHECK (action IN ('create','read','update','delete','export','login','logout','partner_api_call')),
    resource_type VARCHAR(100) NOT NULL,
    resource_id   VARCHAR(36),
    ip_address    INET,
    user_agent    TEXT,
    request_id    VARCHAR(36),
    details       JSONB DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS audit_logs_company_id_idx ON audit_logs(company_id);
CREATE INDEX IF NOT EXISTS audit_logs_agent_id_idx ON audit_logs(agent_id);
CREATE INDEX IF NOT EXISTS audit_logs_action_idx ON audit_logs(action);
CREATE INDEX IF NOT EXISTS audit_logs_resource_idx ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS audit_logs_created_at_idx ON audit_logs(created_at DESC);

CREATE TABLE IF NOT EXISTS partner_api_calls (
    id                VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id        VARCHAR(36) NOT NULL,
    from_company_slug VARCHAR(100) NOT NULL,
    from_company_id   VARCHAR(36),
    endpoint          VARCHAR(200) NOT NULL,
    method            VARCHAR(10) NOT NULL CHECK (method IN ('GET','POST','PUT','DELETE','PATCH')),
    request_body      TEXT,
    response_status   INT NOT NULL,
    response_body     TEXT,
    error_code        VARCHAR(50),
    duration_ms       INT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS partner_api_calls_company_id_idx ON partner_api_calls(company_id);
CREATE INDEX IF NOT EXISTS partner_api_calls_from_slug_idx ON partner_api_calls(from_company_slug);
CREATE INDEX IF NOT EXISTS partner_api_calls_created_at_idx ON partner_api_calls(created_at DESC);

CREATE TABLE IF NOT EXISTS sensitive_access_logs (
    id               VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       VARCHAR(36) NOT NULL,
    agent_id         VARCHAR(36) NOT NULL,
    agent_name       VARCHAR(100) NOT NULL,
    action           VARCHAR(50) NOT NULL CHECK (action IN ('view_payroll','edit_payroll','view_contracts','edit_contracts','view_personal_info','edit_personal_info','delete_agent','export_all_data')),
    target_resource  VARCHAR(100) NOT NULL,
    target_agent_id  VARCHAR(36),
    ip_address       INET,
    user_agent       TEXT,
    request_id       VARCHAR(36),
    justification    TEXT,
    approval_id      VARCHAR(36),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS sensitive_access_logs_company_id_idx ON sensitive_access_logs(company_id);
CREATE INDEX IF NOT EXISTS sensitive_access_logs_agent_id_idx ON sensitive_access_logs(agent_id);
CREATE INDEX IF NOT EXISTS sensitive_access_logs_action_idx ON sensitive_access_logs(action);
CREATE INDEX IF NOT EXISTS sensitive_access_logs_target_agent_idx ON sensitive_access_logs(target_agent_id);
CREATE INDEX IF NOT EXISTS sensitive_access_logs_created_at_idx ON sensitive_access_logs(created_at DESC);
