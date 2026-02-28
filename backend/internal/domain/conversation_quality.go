package domain

import (
	"encoding/json"
	"time"
)

type EvaluatorType string

const (
	EvaluatorRule     EvaluatorType = "rule"
	EvaluatorLLMJudge EvaluatorType = "llm_judge"
)

type ConversationQualityScore struct {
	ID              string          `gorm:"column:id"               json:"id"`
	CompanyID       string          `gorm:"column:company_id"       json:"companyId"`
	TraceID         string          `gorm:"column:trace_id"         json:"traceId"`
	ScoredAgentID   *string         `gorm:"column:scored_agent_id"  json:"scoredAgentId"`
	EvaluatorType   EvaluatorType   `gorm:"column:evaluator_type"   json:"evaluatorType"`
	OverallScore    *float64        `gorm:"column:overall_score"    json:"overallScore"`
	DimensionScores json.RawMessage `gorm:"column:dimension_scores" json:"dimensionScores"`
	Feedback        *string         `gorm:"column:feedback"         json:"feedback"`
	CreatedAt       time.Time       `gorm:"column:created_at"       json:"createdAt"`
}
