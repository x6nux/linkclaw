package domain

import "time"

type SuggestionType string
type SuggestionStatus string
type SuggestionPriority string
type ChangeType string

const (
	SuggestionTypeTone      SuggestionType = "tone"
	SuggestionTypeStructure SuggestionType = "structure"
	SuggestionTypeContent   SuggestionType = "content"
	SuggestionTypeLength    SuggestionType = "length"

	SuggestionStatusPending  SuggestionStatus = "pending"
	SuggestionStatusApproved SuggestionStatus = "approved"
	SuggestionStatusRejected SuggestionStatus = "rejected"
	SuggestionStatusApplied  SuggestionStatus = "applied"

	SuggestionPriorityLow    SuggestionPriority = "low"
	SuggestionPriorityMedium SuggestionPriority = "medium"
	SuggestionPriorityHigh   SuggestionPriority = "high"

	ChangeTypeManual        ChangeType = "manual"
	ChangeTypeAISuggested   ChangeType = "ai_suggested"
	ChangeTypeAutoOptimized ChangeType = "auto_optimized"
)

type PersonaOptimizationSuggestion struct {
	ID              string             `gorm:"column:id"               json:"id"`
	CompanyID       string             `gorm:"column:company_id"       json:"company_id"`
	AgentID         string             `gorm:"column:agent_id"         json:"agent_id"`
	SuggestionType  SuggestionType     `gorm:"column:suggestion_type"  json:"suggestion_type"`
	Priority        SuggestionPriority `gorm:"column:priority"         json:"priority"`
	CurrentPersona  string             `gorm:"column:current_persona"  json:"current_persona"`
	SuggestedChange string             `gorm:"column:suggested_change" json:"suggested_change"`
	Reason          string             `gorm:"column:reason"           json:"reason"`
	Confidence      float64            `gorm:"column:confidence"       json:"confidence"`
	Status          SuggestionStatus   `gorm:"column:status"           json:"status"`
	AppliedAt       *time.Time         `gorm:"column:applied_at"       json:"applied_at"`
	CreatedAt       time.Time          `gorm:"column:created_at"       json:"created_at"`
	UpdatedAt       time.Time          `gorm:"column:updated_at"       json:"updated_at"`
}

type PersonaHistory struct {
	ID           string      `gorm:"column:id"            json:"id"`
	CompanyID    string      `gorm:"column:company_id"    json:"company_id"`
	AgentID      string      `gorm:"column:agent_id"      json:"agent_id"`
	OldPersona   string      `gorm:"column:old_persona"   json:"old_persona"`
	NewPersona   string      `gorm:"column:new_persona"   json:"new_persona"`
	ChangeReason string      `gorm:"column:change_reason" json:"change_reason"`
	SuggestionID *string     `gorm:"column:suggestion_id" json:"suggestion_id"`
	ChangeType   ChangeType  `gorm:"column:change_type"   json:"change_type"`
	ChangedBy    string      `gorm:"column:changed_by"    json:"changed_by"`
	CreatedAt    time.Time   `gorm:"column:created_at"    json:"created_at"`
}

type ABTestStatus string

const (
	ABTestStatusRunning   ABTestStatus = "running"
	ABTestStatusPaused    ABTestStatus = "paused"
	ABTestStatusCompleted ABTestStatus = "completed"
	ABTestStatusStopped   ABTestStatus = "stopped"
)

type ABTestPersona struct {
	ID                    string      `gorm:"column:id"                      json:"id"`
	CompanyID             string      `gorm:"column:company_id"              json:"company_id"`
	Name                  string      `gorm:"column:name"                    json:"name"`
	Description           string      `gorm:"column:description"             json:"description"`
	ControlAgentID        string      `gorm:"column:control_agent_id"        json:"control_agent_id"`
	ControlPersona        string      `gorm:"column:control_persona"         json:"control_persona"`
	VariantAgentID        string      `gorm:"column:variant_agent_id"        json:"variant_agent_id"`
	VariantPersona        string      `gorm:"column:variant_persona"         json:"variant_persona"`
	Status                ABTestStatus `gorm:"column:status"                json:"status"`
	StartTime             time.Time   `gorm:"column:start_time"              json:"start_time"`
	EndTime               *time.Time  `gorm:"column:end_time"                json:"end_time"`
	ControlTasksCompleted int         `gorm:"column:control_tasks_completed" json:"control_tasks_completed"`
	VariantTasksCompleted int         `gorm:"column:variant_tasks_completed" json:"variant_tasks_completed"`
	ControlAvgDuration    int         `gorm:"column:control_avg_duration"    json:"control_avg_duration"`
	VariantAvgDuration    int         `gorm:"column:variant_avg_duration"    json:"variant_avg_duration"`
	Winner                *string     `gorm:"column:winner"                  json:"winner"`
	CreatedAt             time.Time   `gorm:"column:created_at"              json:"created_at"`
	UpdatedAt             time.Time   `gorm:"column:updated_at"              json:"updated_at"`
}
