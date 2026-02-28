package domain

import (
	"encoding/json"
	"time"
)

type ApprovalRequestType string

const (
	ApprovalHire           ApprovalRequestType = "hire"
	ApprovalFire           ApprovalRequestType = "fire"
	ApprovalBudgetOverride ApprovalRequestType = "budget_override"
	ApprovalTaskEscalation ApprovalRequestType = "task_escalation"
	ApprovalCustom         ApprovalRequestType = "custom"
)

type ApprovalStatus string

const (
	ApprovalPending   ApprovalStatus = "pending"
	ApprovalApproved  ApprovalStatus = "approved"
	ApprovalRejected  ApprovalStatus = "rejected"
	ApprovalCancelled ApprovalStatus = "cancelled"
)

type ApprovalRequest struct {
	ID             string              `gorm:"column:id"                 json:"id"`
	CompanyID      string              `gorm:"column:company_id"         json:"companyId"`
	RequesterID    string              `gorm:"column:requester_id"       json:"requesterId"`
	ApproverID     *string             `gorm:"column:approver_id"        json:"approverId"`
	RequestType    ApprovalRequestType `gorm:"column:request_type"       json:"requestType"`
	Status         ApprovalStatus      `gorm:"column:status"             json:"status"`
	Payload        json.RawMessage     `gorm:"column:payload;type:jsonb" json:"payload"`
	Reason         string              `gorm:"column:reason"             json:"reason"`
	DecisionReason string              `gorm:"column:decision_reason"    json:"decisionReason"`
	CreatedAt      time.Time           `gorm:"column:created_at"         json:"createdAt"`
	DecidedAt      *time.Time          `gorm:"column:decided_at"         json:"decidedAt"`
}
