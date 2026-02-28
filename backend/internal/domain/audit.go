package domain

import (
	"encoding/json"
	"time"
)

type AuditAction string

const (
	AuditActionCreate         AuditAction = "create"
	AuditActionRead           AuditAction = "read"
	AuditActionUpdate         AuditAction = "update"
	AuditActionDelete         AuditAction = "delete"
	AuditActionExport         AuditAction = "export"
	AuditActionLogin          AuditAction = "login"
	AuditActionLogout         AuditAction = "logout"
	AuditActionPartnerAPICall AuditAction = "partner_api_call"
)

type AuditLog struct {
	ID           string          `gorm:"column:id"            json:"id"`
	CompanyID    string          `gorm:"column:company_id"    json:"company_id"`
	AgentID      string          `gorm:"column:agent_id"      json:"agent_id"`
	AgentName    string          `gorm:"column:agent_name"    json:"agent_name"`
	Action       AuditAction     `gorm:"column:action"        json:"action"`
	ResourceType string          `gorm:"column:resource_type" json:"resource_type"`
	ResourceID   *string         `gorm:"column:resource_id"   json:"resource_id,omitempty"`
	IPAddress    *string         `gorm:"column:ip_address"    json:"ip_address,omitempty"`
	UserAgent    *string         `gorm:"column:user_agent"    json:"user_agent,omitempty"`
	RequestID    *string         `gorm:"column:request_id"    json:"request_id,omitempty"`
	Details      json.RawMessage `gorm:"column:details"       json:"details,omitempty"`
	CreatedAt    time.Time       `gorm:"column:created_at"    json:"created_at"`
}

type PartnerAPICall struct {
	ID              string    `gorm:"column:id"                json:"id"`
	CompanyID       string    `gorm:"column:company_id"        json:"company_id"`
	FromCompanySlug string    `gorm:"column:from_company_slug" json:"from_company_slug"`
	FromCompanyID   *string   `gorm:"column:from_company_id"   json:"from_company_id,omitempty"`
	Endpoint        string    `gorm:"column:endpoint"          json:"endpoint"`
	Method          string    `gorm:"column:method"            json:"method"`
	RequestBody     *string   `gorm:"column:request_body"      json:"request_body,omitempty"`
	ResponseStatus  int       `gorm:"column:response_status"   json:"response_status"`
	ResponseBody    *string   `gorm:"column:response_body"     json:"response_body,omitempty"`
	ErrorCode       *string   `gorm:"column:error_code"        json:"error_code,omitempty"`
	DurationMs      int       `gorm:"column:duration_ms"       json:"duration_ms"`
	CreatedAt       time.Time `gorm:"column:created_at"        json:"created_at"`
}

type SensitiveAction string

const (
	SensitiveActionViewPayroll      SensitiveAction = "view_payroll"
	SensitiveActionEditPayroll      SensitiveAction = "edit_payroll"
	SensitiveActionViewContracts    SensitiveAction = "view_contracts"
	SensitiveActionEditContracts    SensitiveAction = "edit_contracts"
	SensitiveActionViewPersonalInfo SensitiveAction = "view_personal_info"
	SensitiveActionEditPersonalInfo SensitiveAction = "edit_personal_info"
	SensitiveActionDeleteAgent      SensitiveAction = "delete_agent"
	SensitiveActionExportAllData    SensitiveAction = "export_all_data"
)

type SensitiveAccessLog struct {
	ID             string          `gorm:"column:id"               json:"id"`
	CompanyID      string          `gorm:"column:company_id"       json:"company_id"`
	AgentID        string          `gorm:"column:agent_id"         json:"agent_id"`
	AgentName      string          `gorm:"column:agent_name"       json:"agent_name"`
	Action         SensitiveAction `gorm:"column:action"           json:"action"`
	TargetResource string          `gorm:"column:target_resource"  json:"target_resource"`
	TargetAgentID  *string         `gorm:"column:target_agent_id"  json:"target_agent_id,omitempty"`
	IPAddress      *string         `gorm:"column:ip_address"       json:"ip_address,omitempty"`
	UserAgent      *string         `gorm:"column:user_agent"       json:"user_agent,omitempty"`
	RequestID      *string         `gorm:"column:request_id"       json:"request_id,omitempty"`
	Justification  *string         `gorm:"column:justification"    json:"justification,omitempty"`
	ApprovalID     *string         `gorm:"column:approval_id"      json:"approval_id,omitempty"`
	CreatedAt      time.Time       `gorm:"column:created_at"       json:"created_at"`
}
