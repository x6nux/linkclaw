package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
)

type approvalRepo struct {
	db *gorm.DB
}

func NewApprovalRepo(db *gorm.DB) ApprovalRepo {
	return &approvalRepo{db: db}
}

func (r *approvalRepo) Create(ctx context.Context, req *domain.ApprovalRequest) error {
	q := `INSERT INTO approval_requests
		(id, company_id, requester_id, approver_id, request_type, status, payload, reason, decision_reason, created_at, decided_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	result := r.db.WithContext(ctx).Exec(q,
		req.ID, req.CompanyID, req.RequesterID, req.ApproverID, string(req.RequestType),
		string(req.Status), req.Payload, req.Reason, req.DecisionReason, req.CreatedAt, req.DecidedAt,
	)
	if result.Error != nil {
		return fmt.Errorf("approval create: %w", result.Error)
	}
	return nil
}

func (r *approvalRepo) GetByID(ctx context.Context, id string) (*domain.ApprovalRequest, error) {
	var req domain.ApprovalRequest
	result := r.db.WithContext(ctx).Raw(`SELECT * FROM approval_requests WHERE id = $1`, id).Scan(&req)
	if result.Error != nil {
		return nil, fmt.Errorf("approval get: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &req, nil
}

func (r *approvalRepo) List(ctx context.Context, q ApprovalQuery) ([]*domain.ApprovalRequest, int, error) {
	var (
		conds []string
		args  []any
	)
	idx := 1

	if q.CompanyID != "" {
		conds = append(conds, fmt.Sprintf("company_id = $%d", idx))
		args = append(args, q.CompanyID)
		idx++
	}
	if q.RequesterID != "" {
		conds = append(conds, fmt.Sprintf("requester_id = $%d", idx))
		args = append(args, q.RequesterID)
		idx++
	}
	if q.ApproverID != "" {
		conds = append(conds, fmt.Sprintf("approver_id = $%d", idx))
		args = append(args, q.ApproverID)
		idx++
	}
	if q.Status != "" {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(q.Status))
		idx++
	}
	if q.RequestType != "" {
		conds = append(conds, fmt.Sprintf("request_type = $%d", idx))
		args = append(args, string(q.RequestType))
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM approval_requests`+where,
		args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("approval count: %w", err)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}

	listQ := `SELECT * FROM approval_requests` + where +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) +
		` OFFSET $` + fmt.Sprint(len(args)+2)
	listArgs := append(args, limit, q.Offset)

	var requests []*domain.ApprovalRequest
	if err := r.db.WithContext(ctx).Raw(listQ, listArgs...).Scan(&requests).Error; err != nil {
		return nil, 0, fmt.Errorf("approval list: %w", err)
	}
	return requests, int(total), nil
}

func (r *approvalRepo) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.ApprovalStatus,
	decisionReason string,
	decidedAt *time.Time,
) error {
	result := r.db.WithContext(ctx).Exec(
		`UPDATE approval_requests SET status = $1, decision_reason = $2, decided_at = $3 WHERE id = $4 AND status = 'pending'`,
		string(status), decisionReason, decidedAt, id,
	)
	if result.Error != nil {
		return fmt.Errorf("approval update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("approval request not in pending state")
	}
	return nil
}
