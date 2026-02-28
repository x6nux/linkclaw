package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

type BudgetWatcher struct {
	repo repository.ObservabilityRepo
	db   *gorm.DB
	stop chan struct{}
}

func NewBudgetWatcher(repo repository.ObservabilityRepo, db *gorm.DB) *BudgetWatcher {
	return &BudgetWatcher{repo: repo, db: db, stop: make(chan struct{})}
}

func (w *BudgetWatcher) Start() { go w.run() }
func (w *BudgetWatcher) Stop()  { close(w.stop) }

func (w *BudgetWatcher) run() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := w.check(context.Background()); err != nil {
				log.Printf("budget_watcher error: %v", err)
			}
		case <-w.stop:
			return
		}
	}
}

func periodBounds(p domain.BudgetPeriod) (start, end time.Time) {
	now := time.Now().UTC()
	switch p {
	case domain.BudgetPeriodDaily:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 1)
	case domain.BudgetPeriodWeekly:
		offset := int(now.Weekday())
		start = time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 0, 7)
	default: // monthly
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0)
	}
	return
}

func (w *BudgetWatcher) aggregateCost(ctx context.Context, companyID string, scopeType domain.BudgetScopeType, scopeID *string, start, end time.Time) (int64, error) {
	q := `SELECT COALESCE(SUM(cost_microdollars), 0) FROM llm_usage_logs WHERE company_id = $1 AND created_at >= $2 AND created_at < $3`
	args := []interface{}{companyID, start, end}
	idx := 4
	if scopeType == domain.BudgetScopeAgent && scopeID != nil {
		q += fmt.Sprintf(" AND agent_id = $%d", idx)
		args = append(args, *scopeID)
	} else if scopeType == domain.BudgetScopeProvider && scopeID != nil {
		q += fmt.Sprintf(" AND provider_id = $%d", idx)
		args = append(args, *scopeID)
	}
	var total int64
	if err := w.db.WithContext(ctx).Raw(q, args...).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (w *BudgetWatcher) check(ctx context.Context) error {
	policies, err := w.repo.ListActiveBudgetPolicies(ctx, "")
	if err != nil {
		return err
	}
	for _, p := range policies {
		start, end := periodBounds(p.Period)
		current, err := w.aggregateCost(ctx, p.CompanyID, p.ScopeType, p.ScopeID, start, end)
		if err != nil {
			continue
		}
		ratio := float64(current) / float64(p.BudgetMicrodollars)
		var level domain.BudgetAlertLevel
		switch {
		case ratio >= p.CriticalRatio:
			level = domain.AlertLevelCritical
		case ratio >= p.WarnRatio:
			level = domain.AlertLevelWarn
		default:
			continue
		}
		alert := &domain.LLMBudgetAlert{
			ID:                      uuid.New().String(),
			CompanyID:               p.CompanyID,
			PolicyID:                p.ID,
			ScopeType:               p.ScopeType,
			ScopeID:                 p.ScopeID,
			PeriodStart:             start,
			PeriodEnd:               end,
			CurrentCostMicrodollars: current,
			Level:                   level,
			Status:                  domain.AlertStatusOpen,
		}
		if err := w.repo.CreateBudgetAlert(ctx, alert); err != nil {
			continue
		}
		event.Global.Publish(event.NewEvent(event.BudgetAlertCreated, event.BudgetAlertPayload{
			AlertID:   alert.ID,
			CompanyID: p.CompanyID,
			PolicyID:  p.ID,
			Level:     string(level),
		}))
	}
	return nil
}
