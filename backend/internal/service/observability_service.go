package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type TraceTree struct {
	Run   *domain.TraceRun    `json:"run"`
	Spans []*domain.TraceSpan `json:"spans"`
}

type ObservabilityService struct {
	repo repository.ObservabilityRepo
}

func NewObservabilityService(repo repository.ObservabilityRepo) *ObservabilityService {
	return &ObservabilityService{repo: repo}
}

func (s *ObservabilityService) StartTrace(ctx context.Context, companyID string, rootAgentID *string, sourceType domain.TraceSourceType, sourceRefID *string) (*domain.TraceRun, error) {
	now := time.Now()
	t := &domain.TraceRun{
		ID:          uuid.New().String(),
		CompanyID:   companyID,
		RootAgentID: rootAgentID,
		SourceType:  sourceType,
		SourceRefID: sourceRefID,
		Status:      domain.TraceStatusRunning,
		StartedAt:   now,
		CreatedAt:   now,
	}
	if err := s.repo.CreateTraceRun(ctx, t); err != nil {
		return nil, fmt.Errorf("start trace: %w", err)
	}
	return t, nil
}

func (s *ObservabilityService) StartSpan(ctx context.Context, traceID string, parentSpanID *string, agentID *string, spanType domain.SpanType, name string) (*domain.TraceSpan, error) {
	tr, err := s.repo.GetTraceRunByID(ctx, traceID)
	if err != nil || tr == nil {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}
	now := time.Now()
	sp := &domain.TraceSpan{
		ID:           uuid.New().String(),
		TraceID:      traceID,
		ParentSpanID: parentSpanID,
		CompanyID:    tr.CompanyID,
		AgentID:      agentID,
		SpanType:     spanType,
		Name:         name,
		Status:       domain.TraceStatusRunning,
		StartedAt:    now,
		CreatedAt:    now,
	}
	if err := s.repo.CreateTraceSpan(ctx, sp); err != nil {
		return nil, fmt.Errorf("start span: %w", err)
	}
	return sp, nil
}

func (s *ObservabilityService) EndSpan(ctx context.Context, spanID string, status domain.TraceStatus, inputTokens, outputTokens *int, cost *int64, errorMsg *string) error {
	sp, err := s.repo.GetTraceSpanByID(ctx, spanID)
	if err != nil || sp == nil {
		return fmt.Errorf("span not found: %s", spanID)
	}
	now := time.Now()
	dur := int(now.Sub(sp.StartedAt).Milliseconds())
	return s.repo.UpdateTraceSpan(ctx, spanID, status, &now, &dur, inputTokens, outputTokens, cost, errorMsg)
}

func (s *ObservabilityService) EndTrace(ctx context.Context, traceID string, status domain.TraceStatus, errorMsg *string) error {
	tr, err := s.repo.GetTraceRunByID(ctx, traceID)
	if err != nil || tr == nil {
		return fmt.Errorf("trace not found: %s", traceID)
	}
	now := time.Now()
	dur := int(now.Sub(tr.StartedAt).Milliseconds())
	spans, _ := s.repo.ListTraceSpansByTraceID(ctx, traceID)
	var totalCost int64
	var totalIn, totalOut int
	for _, sp := range spans {
		if sp.CostMicrodollars != nil {
			totalCost += *sp.CostMicrodollars
		}
		if sp.InputTokens != nil {
			totalIn += *sp.InputTokens
		}
		if sp.OutputTokens != nil {
			totalOut += *sp.OutputTokens
		}
	}
	if err := s.repo.UpdateTraceRunTotals(ctx, traceID, totalCost, totalIn, totalOut); err != nil {
		return err
	}
	return s.repo.UpdateTraceRunStatus(ctx, traceID, status, &now, &dur, errorMsg)
}

func (s *ObservabilityService) GetTraceTree(ctx context.Context, traceID string) (*TraceTree, error) {
	tr, err := s.repo.GetTraceRunByID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return nil, nil
	}
	spans, err := s.repo.ListTraceSpansByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	return &TraceTree{Run: tr, Spans: spans}, nil
}

func (s *ObservabilityService) ListTraces(ctx context.Context, q repository.TraceRunQuery) ([]*domain.TraceRun, int, error) {
	return s.repo.ListTraceRuns(ctx, q)
}
