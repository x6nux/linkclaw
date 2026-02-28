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
	CompanyID       string             `gorm:"column:company_id"       json:"companyId"`
	AgentID         string             `gorm:"column:agent_id"         json:"agentId"`
	SuggestionType  SuggestionType     `gorm:"column:suggestion_type"  json:"suggestionType"`
	Priority        SuggestionPriority `gorm:"column:priority"         json:"priority"`
	CurrentPersona  string             `gorm:"column:current_persona"  json:"currentPersona"`
	SuggestedChange string             `gorm:"column:suggested_change" json:"suggestedChange"`
	Reason          string             `gorm:"column:reason"           json:"reason"`
	Confidence      float64            `gorm:"column:confidence"       json:"confidence"`
	Status          SuggestionStatus   `gorm:"column:status"           json:"status"`
	AppliedAt       *time.Time         `gorm:"column:applied_at"       json:"appliedAt"`
	CreatedAt       time.Time          `gorm:"column:created_at"       json:"createdAt"`
	UpdatedAt       time.Time          `gorm:"column:updated_at"       json:"updatedAt"`
}

type PersonaHistory struct {
	ID           string      `gorm:"column:id"            json:"id"`
	CompanyID    string      `gorm:"column:company_id"    json:"companyId"`
	AgentID      string      `gorm:"column:agent_id"      json:"agentId"`
	OldPersona   string      `gorm:"column:old_persona"   json:"oldPersona"`
	NewPersona   string      `gorm:"column:new_persona"   json:"newPersona"`
	ChangeReason string      `gorm:"column:change_reason" json:"changeReason"`
	SuggestionID *string     `gorm:"column:suggestion_id" json:"suggestionId"`
	ChangeType   ChangeType  `gorm:"column:change_type"   json:"changeType"`
	ChangedBy    string      `gorm:"column:changed_by"    json:"changedBy"`
	CreatedAt    time.Time   `gorm:"column:created_at"    json:"createdAt"`
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
	CompanyID             string      `gorm:"column:company_id"              json:"companyId"`
	Name                  string      `gorm:"column:name"                    json:"name"`
	Description           string      `gorm:"column:description"             json:"description"`
	ControlAgentID        string      `gorm:"column:control_agent_id"        json:"controlAgentId"`
	ControlPersona        string      `gorm:"column:control_persona"         json:"controlPersona"`
	VariantAgentID        string      `gorm:"column:variant_agent_id"        json:"variantAgentId"`
	VariantPersona        string      `gorm:"column:variant_persona"         json:"variantPersona"`
	Status                ABTestStatus `gorm:"column:status"                json:"status"`
	StartTime             time.Time   `gorm:"column:start_time"              json:"startTime"`
	EndTime               *time.Time  `gorm:"column:end_time"                json:"endTime"`
	ControlTasksCompleted int         `gorm:"column:control_tasks_completed" json:"controlTasksCompleted"`
	VariantTasksCompleted int         `gorm:"column:variant_tasks_completed" json:"variantTasksCompleted"`
	ControlAvgDuration    int         `gorm:"column:control_avg_duration"    json:"controlAvgDuration"`
	VariantAvgDuration    int         `gorm:"column:variant_avg_duration"    json:"variantAvgDuration"`
	Winner                *string     `gorm:"column:winner"                  json:"winner"`
	CreatedAt             time.Time   `gorm:"column:created_at"              json:"createdAt"`
	UpdatedAt             time.Time   `gorm:"column:updated_at"              json:"updatedAt"`
}
