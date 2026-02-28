package event

import "encoding/json"

// Type 事件类型
type Type string

const (
	AgentOnline  Type = "agent.online"
	AgentOffline Type = "agent.offline"
	AgentStatus  Type = "agent.status"
	TaskCreated  Type = "task.created"
	TaskUpdated  Type = "task.updated"
	MessageNew         Type = "message.new"
	AgentInitialized   Type = "agent.initialized"
	BudgetAlertCreated Type = "llm.budget_alert.created"
	ErrorAlertCreated  Type = "llm.error_alert.created"
	ApprovalApproved   Type = "approval.approved"
)

// Event 是平台内部事件的通用结构
type Event struct {
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// AgentInitializedPayload agent 初始化完成事件 payload
type AgentInitializedPayload struct {
	AgentID   string `json:"agent_id"`
	CompanyID string `json:"company_id"`
}

// AgentStatusPayload agent 状态变更事件 payload
type AgentStatusPayload struct {
	AgentID   string `json:"agent_id"`
	CompanyID string `json:"company_id"`
	Status    string `json:"status"`
}

// TaskCreatedPayload 任务创建事件 payload
type TaskCreatedPayload struct {
	TaskID     string  `json:"task_id"`
	CompanyID  string  `json:"company_id"`
	Title      string  `json:"title"`
	AssigneeID *string `json:"assignee_id,omitempty"`
}

// TaskUpdatedPayload 任务变更事件 payload
type TaskUpdatedPayload struct {
	TaskID     string  `json:"task_id"`
	CompanyID  string  `json:"company_id"`
	Status     string  `json:"status"`
	Title      string  `json:"title"`
	AssigneeID *string `json:"assignee_id,omitempty"`
}

// MessageNewPayload 新消息事件 payload（含完整内容，前端无需二次 fetch）
type MessageNewPayload struct {
	MessageID   string  `json:"message_id"`
	CompanyID   string  `json:"company_id"`
	ChannelID   *string `json:"channel_id,omitempty"`
	ChannelName *string `json:"channel_name,omitempty"`
	ReceiverID  *string `json:"receiver_id,omitempty"`
	SenderID    *string `json:"sender_id,omitempty"`
	MsgType     string  `json:"msg_type"`
	Content     string  `json:"content"`
	CreatedAt   string  `json:"created_at"`
}

func NewEvent(t Type, payload interface{}) Event {
	b, _ := json.Marshal(payload)
	return Event{Type: t, Payload: b}
}

type BudgetAlertPayload struct {
	AlertID   string `json:"alert_id"`
	CompanyID string `json:"company_id"`
	PolicyID  string `json:"policy_id"`
	Level     string `json:"level"`
}

type ApprovalApprovedPayload struct {
	RequestID   string `json:"request_id"`
	CompanyID   string `json:"company_id"`
	RequestType string `json:"request_type"`
	RequesterID string `json:"requester_id"`
}

type ErrorAlertPayload struct {
	PolicyID  string  `json:"policy_id"`
	CompanyID string  `json:"company_id"`
	ErrorRate float64 `json:"error_rate"`
	TotalReqs int64   `json:"total_reqs"`
}
