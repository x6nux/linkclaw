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
	CompanyID          string          `gorm:"column:company_id"          json:"company_id"`
	ScopeType          BudgetScopeType `gorm:"column:scope_type"          json:"scope_type"`
	ScopeID            *string         `gorm:"column:scope_id"            json:"scope_id"`
	Period             BudgetPeriod    `gorm:"column:period"              json:"period"`
	BudgetMicrodollars int64           `gorm:"column:budget_microdollars" json:"budget_microdollars"`
	WarnRatio          float64         `gorm:"column:warn_ratio"          json:"warn_ratio"`
	CriticalRatio      float64         `gorm:"column:critical_ratio"      json:"critical_ratio"`
	HardLimitEnabled   bool            `gorm:"column:hard_limit_enabled"  json:"hard_limit_enabled"`
	IsActive           bool            `gorm:"column:is_active"           json:"is_active"`
	CreatedAt          time.Time       `gorm:"column:created_at"          json:"created_at"`
}

type LLMBudgetAlert struct {
	ID                      string            `gorm:"column:id"                        json:"id"`
	CompanyID               string            `gorm:"column:company_id"                json:"company_id"`
	PolicyID                string            `gorm:"column:policy_id"                 json:"policy_id"`
	ScopeType               BudgetScopeType   `gorm:"column:scope_type"                json:"scope_type"`
	ScopeID                 *string           `gorm:"column:scope_id"                  json:"scope_id"`
	PeriodStart             time.Time         `gorm:"column:period_start"              json:"period_start"`
	PeriodEnd               time.Time         `gorm:"column:period_end"                json:"period_end"`
	CurrentCostMicrodollars int64             `gorm:"column:current_cost_microdollars" json:"current_cost_microdollars"`
	Level                   BudgetAlertLevel  `gorm:"column:level"                     json:"level"`
	Status                  BudgetAlertStatus `gorm:"column:status"                    json:"status"`
	CreatedAt               time.Time         `gorm:"column:created_at"                json:"created_at"`
}

type LLMErrorAlertPolicy struct {
	ID                 string              `gorm:"column:id"                   json:"id"`
	CompanyID          string              `gorm:"column:company_id"           json:"company_id"`
	ScopeType          ErrorAlertScopeType `gorm:"column:scope_type"           json:"scope_type"`
	ScopeID            *string             `gorm:"column:scope_id"             json:"scope_id"`
	WindowMinutes      int                 `gorm:"column:window_minutes"       json:"window_minutes"`
	MinRequests        int                 `gorm:"column:min_requests"         json:"min_requests"`
	ErrorRateThreshold float64             `gorm:"column:error_rate_threshold" json:"error_rate_threshold"`
	CooldownMinutes    int                 `gorm:"column:cooldown_minutes"     json:"cooldown_minutes"`
	CreatedAt          time.Time           `gorm:"column:created_at"           json:"created_at"`
}
