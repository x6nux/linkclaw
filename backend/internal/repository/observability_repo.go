package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type observabilityRepo struct{ db *gorm.DB }

func NewObservabilityRepo(db *gorm.DB) ObservabilityRepo {
	return &observabilityRepo{db: db}
}

// --- TraceRun ---

func (r *observabilityRepo) CreateTraceRun(ctx context.Context, t *domain.TraceRun) error {
	q := `INSERT INTO trace_runs
		(id, company_id, root_agent_id, session_id, source_type, source_ref_id, status, started_at, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	res := r.db.WithContext(ctx).Exec(q, t.ID, t.CompanyID, t.RootAgentID, t.SessionID,
		string(t.SourceType), t.SourceRefID, string(t.Status), t.StartedAt, t.Metadata)
	if res.Error != nil {
		return fmt.Errorf("trace_run create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) GetTraceRunByID(ctx context.Context, id string) (*domain.TraceRun, error) {
	var t domain.TraceRun
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM trace_runs WHERE id = $1`, id).Scan(&t)
	if res.Error != nil {
		return nil, fmt.Errorf("trace_run get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &t, nil
}

func (r *observabilityRepo) ListTraceRuns(ctx context.Context, q TraceRunQuery) ([]*domain.TraceRun, int, error) {
	where := []string{"company_id = $1"}
	args := []interface{}{q.CompanyID}
	idx := 2
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(q.Status))
		idx++
	}
	if q.SourceType != "" {
		where = append(where, fmt.Sprintf("source_type = $%d", idx))
		args = append(args, string(q.SourceType))
		idx++
	}
	clause := strings.Join(where, " AND ")
	var total int64
	if err := r.db.WithContext(ctx).Raw(
		fmt.Sprintf("SELECT COUNT(*) FROM trace_runs WHERE %s", clause), args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	listArgs := append(args, limit, q.Offset)
	listQ := fmt.Sprintf("SELECT * FROM trace_runs WHERE %s ORDER BY started_at DESC LIMIT $%d OFFSET $%d",
		clause, idx, idx+1)
	var rows []*domain.TraceRun
	if err := r.db.WithContext(ctx).Raw(listQ, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("trace_run list: %w", err)
	}
	return rows, int(total), nil
}

func (r *observabilityRepo) UpdateTraceRunStatus(ctx context.Context, id string, status domain.TraceStatus, endedAt *time.Time, durationMs *int, errorMsg *string) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE trace_runs SET status=$1, ended_at=$2, duration_ms=$3, error_msg=$4 WHERE id=$5`,
		string(status), endedAt, durationMs, errorMsg, id)
	return res.Error
}

func (r *observabilityRepo) UpdateTraceRunTotals(ctx context.Context, id string, cost int64, inputTokens, outputTokens int) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE trace_runs SET total_cost_microdollars=$1, total_input_tokens=$2, total_output_tokens=$3 WHERE id=$4`,
		cost, inputTokens, outputTokens, id)
	return res.Error
}

// --- TraceSpan ---

func (r *observabilityRepo) CreateTraceSpan(ctx context.Context, s *domain.TraceSpan) error {
	q := `INSERT INTO trace_spans
		(id, trace_id, parent_span_id, company_id, agent_id, span_type, name, provider_id, request_model, status, started_at, attributes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	res := r.db.WithContext(ctx).Exec(q, s.ID, s.TraceID, s.ParentSpanID, s.CompanyID,
		s.AgentID, string(s.SpanType), s.Name, s.ProviderID, s.RequestModel,
		string(s.Status), s.StartedAt, s.Attributes)
	if res.Error != nil {
		return fmt.Errorf("trace_span create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) GetTraceSpanByID(ctx context.Context, id string) (*domain.TraceSpan, error) {
	var s domain.TraceSpan
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM trace_spans WHERE id = $1`, id).Scan(&s)
	if res.Error != nil {
		return nil, fmt.Errorf("trace_span get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &s, nil
}

func (r *observabilityRepo) ListTraceSpansByTraceID(ctx context.Context, traceID string) ([]*domain.TraceSpan, error) {
	var spans []*domain.TraceSpan
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM trace_spans WHERE trace_id = $1 ORDER BY started_at`, traceID,
	).Scan(&spans).Error; err != nil {
		return nil, fmt.Errorf("trace_span list: %w", err)
	}
	return spans, nil
}

func (r *observabilityRepo) UpdateTraceSpan(ctx context.Context, id string, status domain.TraceStatus, endedAt *time.Time, durationMs *int, inputTokens, outputTokens *int, cost *int64, errorMsg *string) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE trace_spans SET status=$1, ended_at=$2, duration_ms=$3, input_tokens=$4, output_tokens=$5, cost_microdollars=$6, error_msg=$7 WHERE id=$8`,
		string(status), endedAt, durationMs, inputTokens, outputTokens, cost, errorMsg, id)
	return res.Error
}

// --- TraceReplay ---

func (r *observabilityRepo) CreateTraceReplay(ctx context.Context, rp *domain.TraceReplay) error {
	q := `INSERT INTO trace_replays
		(id, company_id, trace_id, span_id, request_headers, response_headers, request_body_enc, response_body_enc, status_code, is_stream)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	res := r.db.WithContext(ctx).Exec(q, rp.ID, rp.CompanyID, rp.TraceID, rp.SpanID,
		rp.RequestHeaders, rp.ResponseHeaders, rp.RequestBodyEnc, rp.ResponseBodyEnc, rp.StatusCode, rp.IsStream)
	if res.Error != nil {
		return fmt.Errorf("trace_replay create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) GetTraceReplayBySpanID(ctx context.Context, spanID string) (*domain.TraceReplay, error) {
	var rp domain.TraceReplay
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM trace_replays WHERE span_id = $1 LIMIT 1`, spanID).Scan(&rp)
	if res.Error != nil {
		return nil, fmt.Errorf("trace_replay get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &rp, nil
}

// --- LLMBudgetPolicy ---

func (r *observabilityRepo) CreateBudgetPolicy(ctx context.Context, p *domain.LLMBudgetPolicy) error {
	q := `INSERT INTO llm_budget_policies
		(id, company_id, scope_type, scope_id, period, budget_microdollars, warn_ratio, critical_ratio, hard_limit_enabled, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	res := r.db.WithContext(ctx).Exec(q, p.ID, p.CompanyID, string(p.ScopeType), p.ScopeID,
		string(p.Period), p.BudgetMicrodollars, p.WarnRatio, p.CriticalRatio, p.HardLimitEnabled, p.IsActive)
	if res.Error != nil {
		return fmt.Errorf("budget_policy create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) UpdateBudgetPolicy(ctx context.Context, p *domain.LLMBudgetPolicy) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE llm_budget_policies SET budget_microdollars=$1, warn_ratio=$2, critical_ratio=$3, hard_limit_enabled=$4, is_active=$5 WHERE id=$6`,
		p.BudgetMicrodollars, p.WarnRatio, p.CriticalRatio, p.HardLimitEnabled, p.IsActive, p.ID)
	return res.Error
}

func (r *observabilityRepo) GetBudgetPolicyByID(ctx context.Context, id string) (*domain.LLMBudgetPolicy, error) {
	var p domain.LLMBudgetPolicy
	res := r.db.WithContext(ctx).Raw(`SELECT * FROM llm_budget_policies WHERE id = $1`, id).Scan(&p)
	if res.Error != nil {
		return nil, fmt.Errorf("budget_policy get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &p, nil
}

func (r *observabilityRepo) ListBudgetPolicies(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error) {
	var policies []*domain.LLMBudgetPolicy
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM llm_budget_policies WHERE company_id = $1 ORDER BY created_at DESC`, companyID,
	).Scan(&policies).Error; err != nil {
		return nil, fmt.Errorf("budget_policy list: %w", err)
	}
	return policies, nil
}

func (r *observabilityRepo) ListActiveBudgetPolicies(ctx context.Context, companyID string) ([]*domain.LLMBudgetPolicy, error) {
	q := `SELECT * FROM llm_budget_policies WHERE is_active = TRUE`
	args := []interface{}{}
	if companyID != "" {
		q += ` AND company_id = $1`
		args = append(args, companyID)
	}
	q += ` ORDER BY created_at`
	var policies []*domain.LLMBudgetPolicy
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&policies).Error; err != nil {
		return nil, fmt.Errorf("budget_policy list_active: %w", err)
	}
	return policies, nil
}

// --- LLMBudgetAlert ---

func (r *observabilityRepo) CreateBudgetAlert(ctx context.Context, a *domain.LLMBudgetAlert) error {
	q := `INSERT INTO llm_budget_alerts
		(id, company_id, policy_id, scope_type, scope_id, period_start, period_end, current_cost_microdollars, level, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	res := r.db.WithContext(ctx).Exec(q, a.ID, a.CompanyID, a.PolicyID, string(a.ScopeType), a.ScopeID,
		a.PeriodStart, a.PeriodEnd, a.CurrentCostMicrodollars, string(a.Level), string(a.Status))
	if res.Error != nil {
		return fmt.Errorf("budget_alert create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) UpdateBudgetAlert(ctx context.Context, id string, status domain.BudgetAlertStatus) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE llm_budget_alerts SET status = $1 WHERE id = $2`, string(status), id)
	return res.Error
}

func (r *observabilityRepo) ListBudgetAlerts(ctx context.Context, q BudgetAlertQuery) ([]*domain.LLMBudgetAlert, error) {
	where := []string{"company_id = $1"}
	args := []interface{}{q.CompanyID}
	idx := 2
	if q.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(q.Status))
		idx++
	}
	if q.Level != "" {
		where = append(where, fmt.Sprintf("level = $%d", idx))
		args = append(args, string(q.Level))
		idx++
	}
	clause := strings.Join(where, " AND ")
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	listArgs := append(args, limit, q.Offset)
	listQ := fmt.Sprintf("SELECT * FROM llm_budget_alerts WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		clause, idx, idx+1)
	var alerts []*domain.LLMBudgetAlert
	if err := r.db.WithContext(ctx).Raw(listQ, listArgs...).Scan(&alerts).Error; err != nil {
		return nil, fmt.Errorf("budget_alert list: %w", err)
	}
	return alerts, nil
}

// --- LLMErrorAlertPolicy ---

func (r *observabilityRepo) CreateErrorAlertPolicy(ctx context.Context, p *domain.LLMErrorAlertPolicy) error {
	q := `INSERT INTO llm_error_alert_policies
		(id, company_id, scope_type, scope_id, window_minutes, min_requests, error_rate_threshold, cooldown_minutes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	res := r.db.WithContext(ctx).Exec(q, p.ID, p.CompanyID, string(p.ScopeType), p.ScopeID,
		p.WindowMinutes, p.MinRequests, p.ErrorRateThreshold, p.CooldownMinutes)
	if res.Error != nil {
		return fmt.Errorf("error_alert_policy create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) UpdateErrorAlertPolicy(ctx context.Context, p *domain.LLMErrorAlertPolicy) error {
	res := r.db.WithContext(ctx).Exec(
		`UPDATE llm_error_alert_policies SET window_minutes=$1, min_requests=$2, error_rate_threshold=$3, cooldown_minutes=$4 WHERE id=$5`,
		p.WindowMinutes, p.MinRequests, p.ErrorRateThreshold, p.CooldownMinutes, p.ID)
	return res.Error
}

func (r *observabilityRepo) ListErrorAlertPolicies(ctx context.Context, companyID string) ([]*domain.LLMErrorAlertPolicy, error) {
	q := `SELECT * FROM llm_error_alert_policies`
	args := []interface{}{}
	if companyID != "" {
		q += ` WHERE company_id = $1`
		args = append(args, companyID)
	}
	q += ` ORDER BY created_at`
	var policies []*domain.LLMErrorAlertPolicy
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&policies).Error; err != nil {
		return nil, fmt.Errorf("error_alert_policy list: %w", err)
	}
	return policies, nil
}

// --- ConversationQualityScore ---

func (r *observabilityRepo) CreateQualityScore(ctx context.Context, s *domain.ConversationQualityScore) error {
	q := `INSERT INTO conversation_quality_scores
		(id, company_id, trace_id, scored_agent_id, evaluator_type, overall_score, dimension_scores, feedback)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	res := r.db.WithContext(ctx).Exec(q, s.ID, s.CompanyID, s.TraceID, s.ScoredAgentID,
		string(s.EvaluatorType), s.OverallScore, s.DimensionScores, s.Feedback)
	if res.Error != nil {
		return fmt.Errorf("quality_score create: %w", res.Error)
	}
	return nil
}

func (r *observabilityRepo) ListQualityScores(ctx context.Context, q QualityScoreQuery) ([]*domain.ConversationQualityScore, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	var scores []*domain.ConversationQualityScore
	if err := r.db.WithContext(ctx).Raw(
		`SELECT * FROM conversation_quality_scores WHERE company_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		q.CompanyID, limit, q.Offset,
	).Scan(&scores).Error; err != nil {
		return nil, fmt.Errorf("quality_score list: %w", err)
	}
	return scores, nil
}

func (r *observabilityRepo) GetQualityScoreByTraceID(ctx context.Context, traceID string) (*domain.ConversationQualityScore, error) {
	var s domain.ConversationQualityScore
	res := r.db.WithContext(ctx).Raw(
		`SELECT * FROM conversation_quality_scores WHERE trace_id = $1 LIMIT 1`, traceID,
	).Scan(&s)
	if res.Error != nil {
		return nil, fmt.Errorf("quality_score get: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, nil
	}
	return &s, nil
}

func (r *observabilityRepo) GetTraceOverview(ctx context.Context, companyID string) (*TraceOverview, error) {
	var overview TraceOverview
	if err := r.db.WithContext(ctx).Raw(
		`SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(AVG(duration_ms), 0) AS avg_latency_ms,
			COALESCE(SUM(total_cost_microdollars), 0) AS total_cost_microdollars
		FROM trace_runs
		WHERE company_id = $1`,
		companyID,
	).Scan(&overview).Error; err != nil {
		return nil, fmt.Errorf("trace_overview get: %w", err)
	}
	return &overview, nil
}
