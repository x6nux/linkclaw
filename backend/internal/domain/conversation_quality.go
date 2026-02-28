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
	CompanyID       string          `gorm:"column:company_id"       json:"company_id"`
	TraceID         string          `gorm:"column:trace_id"         json:"trace_id"`
	ScoredAgentID   *string         `gorm:"column:scored_agent_id"  json:"scored_agent_id"`
	EvaluatorType   EvaluatorType   `gorm:"column:evaluator_type"   json:"evaluator_type"`
	OverallScore    *float64        `gorm:"column:overall_score"    json:"overall_score"`
	DimensionScores json.RawMessage `gorm:"column:dimension_scores" json:"dimension_scores"`
	Feedback        *string         `gorm:"column:feedback"         json:"feedback"`
	CreatedAt       time.Time       `gorm:"column:created_at"       json:"created_at"`
}
