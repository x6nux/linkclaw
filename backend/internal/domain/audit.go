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
	CompanyID    string          `gorm:"column:company_id"    json:"companyId"`
	AgentID      string          `gorm:"column:agent_id"      json:"agentId"`
	AgentName    string          `gorm:"column:agent_name"    json:"agentName"`
	Action       AuditAction     `gorm:"column:action"        json:"action"`
	ResourceType string          `gorm:"column:resource_type" json:"resourceType"`
	ResourceID   *string         `gorm:"column:resource_id"   json:"resourceId,omitempty"`
	IPAddress    *string         `gorm:"column:ip_address"    json:"ipAddress,omitempty"`
	UserAgent    *string         `gorm:"column:user_agent"    json:"userAgent,omitempty"`
	RequestID    *string         `gorm:"column:request_id"    json:"requestId,omitempty"`
	Details      json.RawMessage `gorm:"column:details"       json:"details,omitempty"`
	CreatedAt    time.Time       `gorm:"column:created_at"    json:"createdAt"`
}

type PartnerAPICall struct {
	ID              string    `gorm:"column:id"                json:"id"`
	CompanyID       string    `gorm:"column:company_id"        json:"companyId"`
	FromCompanySlug string    `gorm:"column:from_company_slug" json:"fromCompanySlug"`
	FromCompanyID   *string   `gorm:"column:from_company_id"   json:"fromCompanyId,omitempty"`
	Endpoint        string    `gorm:"column:endpoint"          json:"endpoint"`
	Method          string    `gorm:"column:method"            json:"method"`
	RequestBody     *string   `gorm:"column:request_body"      json:"requestBody,omitempty"`
	ResponseStatus  int       `gorm:"column:response_status"   json:"responseStatus"`
	ResponseBody    *string   `gorm:"column:response_body"     json:"responseBody,omitempty"`
	ErrorCode       *string   `gorm:"column:error_code"        json:"errorCode,omitempty"`
	DurationMs      int       `gorm:"column:duration_ms"       json:"durationMs"`
	CreatedAt       time.Time `gorm:"column:created_at"        json:"createdAt"`
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
	CompanyID      string          `gorm:"column:company_id"       json:"companyId"`
	AgentID        string          `gorm:"column:agent_id"         json:"agentId"`
	AgentName      string          `gorm:"column:agent_name"       json:"agentName"`
	Action         SensitiveAction `gorm:"column:action"           json:"action"`
	TargetResource string          `gorm:"column:target_resource"  json:"targetResource"`
	TargetAgentID  *string         `gorm:"column:target_agent_id"  json:"targetAgentId,omitempty"`
	IPAddress      *string         `gorm:"column:ip_address"       json:"ipAddress,omitempty"`
	UserAgent      *string         `gorm:"column:user_agent"       json:"userAgent,omitempty"`
	RequestID      *string         `gorm:"column:request_id"       json:"requestId,omitempty"`
	Justification  *string         `gorm:"column:justification"    json:"justification,omitempty"`
	ApprovalID     *string         `gorm:"column:approval_id"      json:"approvalId,omitempty"`
	CreatedAt      time.Time       `gorm:"column:created_at"       json:"createdAt"`
}
