package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

type ErrorAlertWatcher struct {
	repo     repository.ObservabilityRepo
	db       *gorm.DB
	cooldown sync.Map // policyID -> time.Time
	stop     chan struct{}
}

func NewErrorAlertWatcher(repo repository.ObservabilityRepo, db *gorm.DB) *ErrorAlertWatcher {
	return &ErrorAlertWatcher{repo: repo, db: db, stop: make(chan struct{})}
}

func (w *ErrorAlertWatcher) Start() { go w.run() }
func (w *ErrorAlertWatcher) Stop()  { close(w.stop) }

func (w *ErrorAlertWatcher) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := w.check(context.Background()); err != nil {
				log.Printf("error_alert_watcher: %v", err)
			}
		case <-w.stop:
			return
		}
	}
}

type errRateRow struct {
	Total  int64   `gorm:"column:total"`
	Errors int64   `gorm:"column:errors"`
	Rate   float64 `gorm:"column:rate"`
}

func (w *ErrorAlertWatcher) errorRate(ctx context.Context, companyID string, scopeType domain.ErrorAlertScopeType, scopeID *string, windowMinutes int) (rate float64, total int64, err error) {
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	q := `SELECT COUNT(*) AS total,
		SUM(CASE WHEN status IN ('error','timeout') THEN 1 ELSE 0 END) AS errors,
		CASE WHEN COUNT(*) = 0 THEN 0.0
			ELSE CAST(SUM(CASE WHEN status IN ('error','timeout') THEN 1 ELSE 0 END) AS float8)/COUNT(*)
		END AS rate
		FROM llm_usage_logs WHERE company_id = $1 AND created_at >= $2`
	args := []interface{}{companyID, since}
	idx := 3
	if scopeType == domain.ErrorScopeAgent && scopeID != nil {
		q += fmt.Sprintf(" AND agent_id = $%d", idx)
		args = append(args, *scopeID)
	} else if scopeType == domain.ErrorScopeProvider && scopeID != nil {
		q += fmt.Sprintf(" AND provider_id = $%d", idx)
		args = append(args, *scopeID)
	} else if scopeType == domain.ErrorScopeModel && scopeID != nil {
		q += fmt.Sprintf(" AND request_model = $%d", idx)
		args = append(args, *scopeID)
	}
	var r errRateRow
	if err = w.db.WithContext(ctx).Raw(q, args...).Scan(&r).Error; err != nil {
		return
	}
	return r.Rate, r.Total, nil
}

func (w *ErrorAlertWatcher) check(ctx context.Context) error {
	policies, err := w.repo.ListErrorAlertPolicies(ctx, "")
	if err != nil {
		return err
	}
	now := time.Now()
	for _, p := range policies {
		if last, ok := w.cooldown.Load(p.ID); ok {
			if now.Sub(last.(time.Time)) < time.Duration(p.CooldownMinutes)*time.Minute {
				continue
			}
		}
		rate, total, err := w.errorRate(ctx, p.CompanyID, p.ScopeType, p.ScopeID, p.WindowMinutes)
		if err != nil || total < int64(p.MinRequests) || rate < p.ErrorRateThreshold {
			continue
		}
		w.cooldown.Store(p.ID, now)
		event.Global.Publish(event.NewEvent(event.ErrorAlertCreated, event.ErrorAlertPayload{
			PolicyID:  p.ID,
			CompanyID: p.CompanyID,
			ErrorRate: rate,
			TotalReqs: total,
		}))
	}
	return nil
}
