package webhook

import (
	"context"
	"encoding/json"
	"log"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/service"
)

type Subscriber struct {
	webhookSvc *service.WebhookService
	unsubs     []func()
}

func NewSubscriber(webhookSvc *service.WebhookService) *Subscriber {
	return &Subscriber{webhookSvc: webhookSvc}
}

func (s *Subscriber) Start() {
	if s.webhookSvc == nil {
		return
	}

	h := func(e event.Event) { s.handleEvent(e) }
	for _, t := range []event.Type{
		event.AgentOnline,
		event.AgentOffline,
		event.AgentStatus,
		event.TaskCreated,
		event.TaskUpdated,
		event.MessageNew,
		event.ApprovalApproved,
		event.BudgetAlertCreated,
		event.ErrorAlertCreated,
	} {
		s.unsubs = append(s.unsubs, event.Global.Subscribe(t, h))
	}
	log.Println("webhook event subscriber started")
}

func (s *Subscriber) Stop() {
	for _, unsub := range s.unsubs {
		unsub()
	}
	s.unsubs = nil
}

func (s *Subscriber) handleEvent(e event.Event) {
	payload := domain.JSONMap{}
	if len(e.Payload) > 0 {
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return
		}
	}

	companyID, _ := payload["company_id"].(string)
	if companyID == "" {
		return
	}

	mappedEvents := s.mapEventTypes(e, payload)
	for _, eventType := range mappedEvents {
		body := clonePayload(payload)
		go func(evt domain.WebhookEventType, data domain.JSONMap) {
			if err := s.webhookSvc.TriggerWebhook(context.Background(), companyID, evt, data); err != nil {
				log.Printf("webhook trigger failed (type=%s): %v", evt, err)
			}
		}(eventType, body)
	}
}

func (s *Subscriber) mapEventTypes(e event.Event, payload domain.JSONMap) []domain.WebhookEventType {
	switch e.Type {
	case event.AgentOnline:
		return []domain.WebhookEventType{domain.WebhookEventAgentOnline}
	case event.AgentOffline:
		return []domain.WebhookEventType{domain.WebhookEventAgentOffline}
	case event.AgentStatus:
		status, _ := payload["status"].(string)
		if status == "online" {
			return []domain.WebhookEventType{domain.WebhookEventAgentOnline}
		}
		if status == "offline" {
			return []domain.WebhookEventType{domain.WebhookEventAgentOffline}
		}
		return nil
	case event.TaskCreated:
		return []domain.WebhookEventType{domain.WebhookEventTaskCreated}
	case event.TaskUpdated:
		out := []domain.WebhookEventType{domain.WebhookEventTaskUpdated}
		status, _ := payload["status"].(string)
		if status == "done" || status == "completed" {
			out = append(out, domain.WebhookEventTaskCompleted)
		}
		return out
	case event.MessageNew:
		return []domain.WebhookEventType{domain.WebhookEventMessageNew}
	case event.ApprovalApproved:
		return []domain.WebhookEventType{domain.WebhookEventApprovalEvent}
	case event.BudgetAlertCreated:
		return []domain.WebhookEventType{domain.WebhookEventBudgetAlert}
	case event.ErrorAlertCreated:
		return []domain.WebhookEventType{domain.WebhookEventErrorAlert}
	default:
		return nil
	}
}

func clonePayload(in domain.JSONMap) domain.JSONMap {
	out := make(domain.JSONMap, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
