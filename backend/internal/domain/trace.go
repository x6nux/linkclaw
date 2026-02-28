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
	CompanyID             string          `gorm:"column:company_id"              json:"company_id"`
	RootAgentID           *string         `gorm:"column:root_agent_id"           json:"root_agent_id"`
	SessionID             *string         `gorm:"column:session_id"              json:"session_id"`
	SourceType            TraceSourceType `gorm:"column:source_type"             json:"source_type"`
	SourceRefID           *string         `gorm:"column:source_ref_id"           json:"source_ref_id"`
	Status                TraceStatus     `gorm:"column:status"                  json:"status"`
	StartedAt             time.Time       `gorm:"column:started_at"              json:"started_at"`
	EndedAt               *time.Time      `gorm:"column:ended_at"                json:"ended_at"`
	DurationMs            *int            `gorm:"column:duration_ms"             json:"duration_ms"`
	TotalCostMicrodollars int64           `gorm:"column:total_cost_microdollars" json:"total_cost_microdollars"`
	TotalInputTokens      int             `gorm:"column:total_input_tokens"      json:"total_input_tokens"`
	TotalOutputTokens     int             `gorm:"column:total_output_tokens"     json:"total_output_tokens"`
	ErrorMsg              *string         `gorm:"column:error_msg"               json:"error_msg"`
	Metadata              json.RawMessage `gorm:"column:metadata"                json:"metadata"`
	CreatedAt             time.Time       `gorm:"column:created_at"              json:"created_at"`
}

type TraceSpan struct {
	ID               string          `gorm:"column:id"                json:"id"`
	TraceID          string          `gorm:"column:trace_id"          json:"trace_id"`
	ParentSpanID     *string         `gorm:"column:parent_span_id"    json:"parent_span_id"`
	CompanyID        string          `gorm:"column:company_id"        json:"company_id"`
	AgentID          *string         `gorm:"column:agent_id"          json:"agent_id"`
	SpanType         SpanType        `gorm:"column:span_type"         json:"span_type"`
	Name             string          `gorm:"column:name"              json:"name"`
	ProviderID       *string         `gorm:"column:provider_id"       json:"provider_id"`
	RequestModel     *string         `gorm:"column:request_model"     json:"request_model"`
	Status           TraceStatus     `gorm:"column:status"            json:"status"`
	StartedAt        time.Time       `gorm:"column:started_at"        json:"started_at"`
	EndedAt          *time.Time      `gorm:"column:ended_at"          json:"ended_at"`
	DurationMs       *int            `gorm:"column:duration_ms"       json:"duration_ms"`
	InputTokens      *int            `gorm:"column:input_tokens"      json:"input_tokens"`
	OutputTokens     *int            `gorm:"column:output_tokens"     json:"output_tokens"`
	CostMicrodollars *int64          `gorm:"column:cost_microdollars" json:"cost_microdollars"`
	ErrorMsg         *string         `gorm:"column:error_msg"         json:"error_msg"`
	Attributes       json.RawMessage `gorm:"column:attributes"        json:"attributes"`
	CreatedAt        time.Time       `gorm:"column:created_at"        json:"created_at"`
}

type TraceReplay struct {
	ID              string          `gorm:"column:id"                json:"id"`
	CompanyID       string          `gorm:"column:company_id"        json:"company_id"`
	TraceID         string          `gorm:"column:trace_id"          json:"trace_id"`
	SpanID          *string         `gorm:"column:span_id"           json:"span_id"`
	RequestHeaders  json.RawMessage `gorm:"column:request_headers"   json:"request_headers"`
	ResponseHeaders json.RawMessage `gorm:"column:response_headers"  json:"response_headers"`
	RequestBodyEnc  []byte          `gorm:"column:request_body_enc"  json:"-"`
	ResponseBodyEnc []byte          `gorm:"column:response_body_enc" json:"-"`
	StatusCode      *int            `gorm:"column:status_code"       json:"status_code"`
	IsStream        bool            `gorm:"column:is_stream"         json:"is_stream"`
	CreatedAt       time.Time       `gorm:"column:created_at"        json:"created_at"`
}
