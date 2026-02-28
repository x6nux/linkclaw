package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type QualityScoringService struct {
	repo repository.ObservabilityRepo
}

func NewQualityScoringService(repo repository.ObservabilityRepo) *QualityScoringService {
	return &QualityScoringService{repo: repo}
}

type dimensionScores struct {
	SpanSuccessRate *float64 `json:"span_success_rate"`
}

// ScoreConversation 使用规则引擎对 trace 内所有 span 的成功率计算整体质量分。
func (s *QualityScoringService) ScoreConversation(ctx context.Context, traceID string) (*domain.ConversationQualityScore, error) {
	tr, err := s.repo.GetTraceRunByID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}
	spans, err := s.repo.ListTraceSpansByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	if len(spans) == 0 {
		return nil, fmt.Errorf("no spans in trace %s", traceID)
	}
	var success int
	for _, sp := range spans {
		if sp.Status == domain.TraceStatusSuccess {
			success++
		}
	}
	rate := float64(success) / float64(len(spans))
	dims := dimensionScores{SpanSuccessRate: &rate}
	dimsJSON, _ := json.Marshal(dims)
	score := &domain.ConversationQualityScore{
		ID:              uuid.New().String(),
		CompanyID:       tr.CompanyID,
		TraceID:         traceID,
		ScoredAgentID:   tr.RootAgentID,
		EvaluatorType:   domain.EvaluatorRule,
		OverallScore:    &rate,
		DimensionScores: dimsJSON,
		CreatedAt:       time.Now(),
	}
	if err := s.repo.CreateQualityScore(ctx, score); err != nil {
		return nil, fmt.Errorf("score conversation: %w", err)
	}
	return score, nil
}

func (s *QualityScoringService) ListScores(ctx context.Context, q repository.QualityScoreQuery) ([]*domain.ConversationQualityScore, error) {
	return s.repo.ListQualityScores(ctx, q)
}
