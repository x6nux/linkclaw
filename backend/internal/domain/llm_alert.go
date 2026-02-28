package domain

import "time"

type BudgetScopeType string

const (
	BudgetScopeCompany  BudgetScopeType = "company"
	BudgetScopeAgent    BudgetScopeType = "agent"
	BudgetScopeProvider BudgetScopeType = "provider"
)

type BudgetPeriod string

const (
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

type BudgetAlertLevel string

const (
	AlertLevelWarn     BudgetAlertLevel = "warn"
	AlertLevelCritical BudgetAlertLevel = "critical"
	AlertLevelBlocked  BudgetAlertLevel = "blocked"
)

type BudgetAlertStatus string

const (
	AlertStatusOpen     BudgetAlertStatus = "open"
	AlertStatusAcked    BudgetAlertStatus = "acked"
	AlertStatusResolved BudgetAlertStatus = "resolved"
)

type ErrorAlertScopeType string

const (
	ErrorScopeCompany  ErrorAlertScopeType = "company"
	ErrorScopeProvider ErrorAlertScopeType = "provider"
	ErrorScopeModel    ErrorAlertScopeType = "model"
	ErrorScopeAgent    ErrorAlertScopeType = "agent"
)

type LLMBudgetPolicy struct {
	ID                 string          `gorm:"column:id"                  json:"id"`
	CompanyID          string          `gorm:"column:company_id"          json:"companyId"`
	ScopeType          BudgetScopeType `gorm:"column:scope_type"          json:"scopeType"`
	ScopeID            *string         `gorm:"column:scope_id"            json:"scopeId"`
	Period             BudgetPeriod    `gorm:"column:period"              json:"period"`
	BudgetMicrodollars int64           `gorm:"column:budget_microdollars" json:"budgetMicrodollars"`
	WarnRatio          float64         `gorm:"column:warn_ratio"          json:"warnRatio"`
	CriticalRatio      float64         `gorm:"column:critical_ratio"      json:"criticalRatio"`
	HardLimitEnabled   bool            `gorm:"column:hard_limit_enabled"  json:"hardLimitEnabled"`
	IsActive           bool            `gorm:"column:is_active"           json:"isActive"`
	CreatedAt          time.Time       `gorm:"column:created_at"          json:"createdAt"`
}

type LLMBudgetAlert struct {
	ID                      string            `gorm:"column:id"                        json:"id"`
	CompanyID               string            `gorm:"column:company_id"                json:"companyId"`
	PolicyID                string            `gorm:"column:policy_id"                 json:"policyId"`
	ScopeType               BudgetScopeType   `gorm:"column:scope_type"                json:"scopeType"`
	ScopeID                 *string           `gorm:"column:scope_id"                  json:"scopeId"`
	PeriodStart             time.Time         `gorm:"column:period_start"              json:"periodStart"`
	PeriodEnd               time.Time         `gorm:"column:period_end"                json:"periodEnd"`
	CurrentCostMicrodollars int64             `gorm:"column:current_cost_microdollars" json:"currentCostMicrodollars"`
	Level                   BudgetAlertLevel  `gorm:"column:level"                     json:"level"`
	Status                  BudgetAlertStatus `gorm:"column:status"                    json:"status"`
	CreatedAt               time.Time         `gorm:"column:created_at"                json:"createdAt"`
}

type LLMErrorAlertPolicy struct {
	ID                 string              `gorm:"column:id"                   json:"id"`
	CompanyID          string              `gorm:"column:company_id"           json:"companyId"`
	ScopeType          ErrorAlertScopeType `gorm:"column:scope_type"           json:"scopeType"`
	ScopeID            *string             `gorm:"column:scope_id"             json:"scopeId"`
	WindowMinutes      int                 `gorm:"column:window_minutes"       json:"windowMinutes"`
	MinRequests        int                 `gorm:"column:min_requests"         json:"minRequests"`
	ErrorRateThreshold float64             `gorm:"column:error_rate_threshold" json:"errorRateThreshold"`
	CooldownMinutes    int                 `gorm:"column:cooldown_minutes"     json:"cooldownMinutes"`
	CreatedAt          time.Time           `gorm:"column:created_at"           json:"createdAt"`
}
