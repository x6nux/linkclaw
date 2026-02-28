package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) AuditRepo {
	return &auditRepo{db: db}
}

func (r *auditRepo) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	q := `INSERT INTO audit_logs
		(company_id, agent_id, agent_name, action, resource_type, resource_id, ip_address, user_agent, request_id, details)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	res := r.db.WithContext(ctx).Exec(q,
		log.CompanyID, log.AgentID, log.AgentName, string(log.Action), log.ResourceType, log.ResourceID,
		log.IPAddress, log.UserAgent, log.RequestID, log.Details,
	)
	if res.Error != nil {
		return fmt.Errorf("audit_log create: %w", res.Error)
	}
	return nil
}

func (r *auditRepo) CreatePartnerAPICall(ctx context.Context, call *domain.PartnerAPICall) error {
	q := `INSERT INTO partner_api_calls
		(company_id, from_company_slug, from_company_id, endpoint, method, request_body, response_status, response_body, error_code, duration_ms)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	res := r.db.WithContext(ctx).Exec(q,
		call.CompanyID, call.FromCompanySlug, call.FromCompanyID, call.Endpoint, call.Method,
		call.RequestBody, call.ResponseStatus, call.ResponseBody, call.ErrorCode, call.DurationMs,
	)
	if res.Error != nil {
		return fmt.Errorf("partner_api_call create: %w", res.Error)
	}
	return nil
}

func (r *auditRepo) CreateSensitiveAccessLog(ctx context.Context, log *domain.SensitiveAccessLog) error {
	q := `INSERT INTO sensitive_access_logs
		(company_id, agent_id, agent_name, action, target_resource, target_agent_id, ip_address, user_agent, request_id, justification, approval_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	res := r.db.WithContext(ctx).Exec(q,
		log.CompanyID, log.AgentID, log.AgentName, string(log.Action), log.TargetResource, log.TargetAgentID,
		log.IPAddress, log.UserAgent, log.RequestID, log.Justification, log.ApprovalID,
	)
	if res.Error != nil {
		return fmt.Errorf("sensitive_access_log create: %w", res.Error)
	}
	return nil
}

func (r *auditRepo) ListAuditLogs(ctx context.Context, q AuditQuery) ([]*domain.AuditLog, int, error) {
	var (
		conds = []string{"company_id = $1"}
		args  = []any{q.CompanyID}
	)
	idx := 2

	if q.ResourceType != "" {
		conds = append(conds, fmt.Sprintf("resource_type = $%d", idx))
		args = append(args, q.ResourceType)
		idx++
	}
	if q.Action != "" {
		conds = append(conds, fmt.Sprintf("action = $%d", idx))
		args = append(args, q.Action)
		idx++
	}
	if q.AgentID != "" {
		conds = append(conds, fmt.Sprintf("agent_id = $%d", idx))
		args = append(args, q.AgentID)
		idx++
	}

	where := " WHERE " + strings.Join(conds, " AND ")

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM audit_logs`+where,
		args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("audit_log count: %w", err)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}

	listQ := `SELECT * FROM audit_logs` + where +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(idx) +
		` OFFSET $` + fmt.Sprint(idx+1)
	listArgs := append(args, limit, q.Offset)

	var logs []*domain.AuditLog
	if err := r.db.WithContext(ctx).Raw(listQ, listArgs...).Scan(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("audit_log list: %w", err)
	}
	return logs, int(total), nil
}
