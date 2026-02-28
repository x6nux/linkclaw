package domain

import (
	"encoding/json"
	"time"
)

type TraceSourceType string

const (
	TraceSourceMCP      TraceSourceType = "mcp"
	TraceSourceHTTP     TraceSourceType = "http"
	TraceSourceWorkflow TraceSourceType = "workflow"
	TraceSourceWS       TraceSourceType = "ws"
)

type TraceStatus string

const (
	TraceStatusRunning TraceStatus = "running"
	TraceStatusSuccess TraceStatus = "success"
	TraceStatusError   TraceStatus = "error"
	TraceStatusTimeout TraceStatus = "timeout"
)

type SpanType string

const (
	SpanTypeMCPTool      SpanType = "mcp_tool"
	SpanTypeLLMCall      SpanType = "llm_call"
	SpanTypeWorkflowNode SpanType = "workflow_node"
	SpanTypeKBRetrieval  SpanType = "kb_retrieval"
	SpanTypeHTTPCall     SpanType = "http_call"
	SpanTypeInternal     SpanType = "internal"
)

type TraceRun struct {
	ID                    string          `gorm:"column:id"                      json:"id"`
	CompanyID             string          `gorm:"column:company_id"              json:"companyId"`
	RootAgentID           *string         `gorm:"column:root_agent_id"           json:"rootAgentId"`
	SessionID             *string         `gorm:"column:session_id"              json:"sessionId"`
	SourceType            TraceSourceType `gorm:"column:source_type"             json:"sourceType"`
	SourceRefID           *string         `gorm:"column:source_ref_id"           json:"sourceRefId"`
	Status                TraceStatus     `gorm:"column:status"                  json:"status"`
	StartedAt             time.Time       `gorm:"column:started_at"              json:"startedAt"`
	EndedAt               *time.Time      `gorm:"column:ended_at"                json:"endedAt"`
	DurationMs            *int            `gorm:"column:duration_ms"             json:"durationMs"`
	TotalCostMicrodollars int64           `gorm:"column:total_cost_microdollars" json:"totalCostMicrodollars"`
	TotalInputTokens      int             `gorm:"column:total_input_tokens"      json:"totalInputTokens"`
	TotalOutputTokens     int             `gorm:"column:total_output_tokens"     json:"totalOutputTokens"`
	ErrorMsg              *string         `gorm:"column:error_msg"               json:"errorMsg"`
	Metadata              json.RawMessage `gorm:"column:metadata"                json:"metadata"`
	CreatedAt             time.Time       `gorm:"column:created_at"              json:"createdAt"`
}

type TraceSpan struct {
	ID               string          `gorm:"column:id"                json:"id"`
	TraceID          string          `gorm:"column:trace_id"          json:"traceId"`
	ParentSpanID     *string         `gorm:"column:parent_span_id"    json:"parentSpanId"`
	CompanyID        string          `gorm:"column:company_id"        json:"companyId"`
	AgentID          *string         `gorm:"column:agent_id"          json:"agentId"`
	SpanType         SpanType        `gorm:"column:span_type"         json:"spanType"`
	Name             string          `gorm:"column:name"              json:"name"`
	ProviderID       *string         `gorm:"column:provider_id"       json:"providerId"`
	RequestModel     *string         `gorm:"column:request_model"     json:"requestModel"`
	Status           TraceStatus     `gorm:"column:status"            json:"status"`
	StartedAt        time.Time       `gorm:"column:started_at"        json:"startedAt"`
	EndedAt          *time.Time      `gorm:"column:ended_at"          json:"endedAt"`
	DurationMs       *int            `gorm:"column:duration_ms"       json:"durationMs"`
	InputTokens      *int            `gorm:"column:input_tokens"      json:"inputTokens"`
	OutputTokens     *int            `gorm:"column:output_tokens"     json:"outputTokens"`
	CostMicrodollars *int64          `gorm:"column:cost_microdollars" json:"costMicrodollars"`
	ErrorMsg         *string         `gorm:"column:error_msg"         json:"errorMsg"`
	Attributes       json.RawMessage `gorm:"column:attributes"        json:"attributes"`
	CreatedAt        time.Time       `gorm:"column:created_at"        json:"createdAt"`
}

type TraceReplay struct {
	ID              string          `gorm:"column:id"                json:"id"`
	CompanyID       string          `gorm:"column:company_id"        json:"companyId"`
	TraceID         string          `gorm:"column:trace_id"          json:"traceId"`
	SpanID          *string         `gorm:"column:span_id"           json:"spanId"`
	RequestHeaders  json.RawMessage `gorm:"column:request_headers"   json:"requestHeaders"`
	ResponseHeaders json.RawMessage `gorm:"column:response_headers"  json:"responseHeaders"`
	RequestBodyEnc  []byte          `gorm:"column:request_body_enc"  json:"-"`
	ResponseBodyEnc []byte          `gorm:"column:response_body_enc" json:"-"`
	StatusCode      *int            `gorm:"column:status_code"       json:"statusCode"`
	IsStream        bool            `gorm:"column:is_stream"         json:"isStream"`
	CreatedAt       time.Time       `gorm:"column:created_at"        json:"createdAt"`
}
