package service

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

const (
	defaultWebhookTimeoutSeconds = 10
	defaultWebhookQueueInterval  = 5 * time.Second
	defaultWebhookQueueBatch     = 50
)

type WebhookService struct {
	webhookRepo    repository.WebhookRepo
	httpClient     *http.Client
	queueInterval  time.Duration
	queueBatchSize int
}

func NewWebhookService(webhookRepo repository.WebhookRepo, queueInterval time.Duration) *WebhookService {
	if queueInterval <= 0 {
		queueInterval = defaultWebhookQueueInterval
	}
	return &WebhookService{
		webhookRepo:    webhookRepo,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		queueInterval:  queueInterval,
		queueBatchSize: defaultWebhookQueueBatch,
	}
}

type CreateWebhookInput struct {
	CompanyID      string
	Name           string
	URL            string
	SigningKeyID   *string
	Events         []domain.WebhookEventType
	SecretHeader   string
	IsActive       *bool
	TimeoutSeconds int
	RetryPolicy    *domain.RetryPolicy
}

type UpdateWebhookInput struct {
	CompanyID      string
	Name           string
	URL            string
	SigningKeyID   *string
	Events         []domain.WebhookEventType
	SecretHeader   string
	IsActive       *bool
	TimeoutSeconds int
	RetryPolicy    *domain.RetryPolicy
}

type CreateSigningKeyInput struct {
	CompanyID string
	Name      string
	KeyType   domain.WebhookSigningKeyType
	PublicKey string
	SecretKey string
	IsActive  *bool
}

func (s *WebhookService) CreateWebhook(ctx context.Context, in CreateWebhookInput) (*domain.Webhook, error) {
	if strings.TrimSpace(in.CompanyID) == "" {
		return nil, fmt.Errorf("company id is required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := validateWebhookURL(in.URL); err != nil {
		return nil, err
	}
	if len(in.Events) == 0 {
		return nil, fmt.Errorf("at least one event is required")
	}

	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}

	timeoutSeconds := in.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultWebhookTimeoutSeconds
	}

	policy := normalizeRetryPolicy(in.RetryPolicy)
	secretHeader := strings.TrimSpace(in.SecretHeader)
	if secretHeader == "" {
		secretHeader = "X-Webhook-Signature"
	}

	w := &domain.Webhook{
		ID:             uuid.New().String(),
		CompanyID:      in.CompanyID,
		Name:           strings.TrimSpace(in.Name),
		URL:            strings.TrimSpace(in.URL),
		SigningKeyID:   in.SigningKeyID,
		Events:         toEventStringList(in.Events),
		SecretHeader:   secretHeader,
		IsActive:       active,
		TimeoutSeconds: timeoutSeconds,
		RetryPolicy:    retryPolicyToJSONMap(policy),
	}

	if err := s.webhookRepo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}
	created, err := s.webhookRepo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, fmt.Errorf("get created webhook: %w", err)
	}
	return created, nil
}

func (s *WebhookService) UpdateWebhook(ctx context.Context, id string, in UpdateWebhookInput) (*domain.Webhook, error) {
	w, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get webhook: %w", err)
	}
	if w == nil {
		return nil, fmt.Errorf("webhook not found")
	}
	if in.CompanyID != "" && w.CompanyID != in.CompanyID {
		return nil, fmt.Errorf("forbidden")
	}

	if strings.TrimSpace(in.Name) != "" {
		w.Name = strings.TrimSpace(in.Name)
	}
	if strings.TrimSpace(in.URL) != "" {
		if err := validateWebhookURL(in.URL); err != nil {
			return nil, err
		}
		w.URL = strings.TrimSpace(in.URL)
	}
	if in.SigningKeyID != nil {
		w.SigningKeyID = in.SigningKeyID
	}
	if len(in.Events) > 0 {
		w.Events = toEventStringList(in.Events)
	}
	if strings.TrimSpace(in.SecretHeader) != "" {
		w.SecretHeader = strings.TrimSpace(in.SecretHeader)
	}
	if in.IsActive != nil {
		w.IsActive = *in.IsActive
	}
	if in.TimeoutSeconds > 0 {
		w.TimeoutSeconds = in.TimeoutSeconds
	}
	if in.RetryPolicy != nil {
		w.RetryPolicy = retryPolicyToJSONMap(normalizeRetryPolicy(in.RetryPolicy))
	}

	if err := s.webhookRepo.Update(ctx, w); err != nil {
		return nil, fmt.Errorf("update webhook: %w", err)
	}
	updated, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get updated webhook: %w", err)
	}
	return updated, nil
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, companyID, id string) error {
	w, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get webhook: %w", err)
	}
	if w == nil {
		return fmt.Errorf("webhook not found")
	}
	if companyID != "" && w.CompanyID != companyID {
		return fmt.Errorf("forbidden")
	}
	if err := s.webhookRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	return nil
}

func (s *WebhookService) GetWebhook(ctx context.Context, companyID, id string) (*domain.Webhook, error) {
	w, err := s.webhookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get webhook: %w", err)
	}
	if w == nil {
		return nil, nil
	}
	if companyID != "" && w.CompanyID != companyID {
		return nil, fmt.Errorf("forbidden")
	}
	return w, nil
}

func (s *WebhookService) ListWebhooks(ctx context.Context, companyID string) ([]*domain.Webhook, error) {
	list, err := s.webhookRepo.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	return list, nil
}

func (s *WebhookService) CreateSigningKey(ctx context.Context, in CreateSigningKeyInput) (*domain.WebhookSigningKey, error) {
	if strings.TrimSpace(in.CompanyID) == "" {
		return nil, fmt.Errorf("company id is required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}

	keyType := in.KeyType
	if keyType == "" {
		keyType = domain.WebhookSigningKeyHMACSHA256
	}
	if keyType != domain.WebhookSigningKeyHMACSHA256 && keyType != domain.WebhookSigningKeyEd25519 {
		return nil, fmt.Errorf("invalid key type")
	}

	secret := strings.TrimSpace(in.SecretKey)
	public := strings.TrimSpace(in.PublicKey)
	if keyType == domain.WebhookSigningKeyHMACSHA256 && secret == "" {
		return nil, fmt.Errorf("secret key is required for hmac")
	}
	if keyType == domain.WebhookSigningKeyEd25519 && secret == "" && public == "" {
		return nil, fmt.Errorf("secret key or public key is required for ed25519")
	}

	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}

	k := &domain.WebhookSigningKey{
		ID:           uuid.New().String(),
		CompanyID:    in.CompanyID,
		Name:         strings.TrimSpace(in.Name),
		KeyType:      keyType,
		PublicKey:    public,
		SecretKeyEnc: []byte(secret),
		IsActive:     active,
		CreatedAt:    time.Now(),
	}

	if err := s.webhookRepo.CreateSigningKey(ctx, k); err != nil {
		return nil, fmt.Errorf("create signing key: %w", err)
	}
	return k, nil
}

func (s *WebhookService) ListSigningKeys(ctx context.Context, companyID string) ([]*domain.WebhookSigningKey, error) {
	list, err := s.webhookRepo.ListSigningKeys(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list signing keys: %w", err)
	}
	return list, nil
}

func (s *WebhookService) DeleteSigningKey(ctx context.Context, companyID, id string) error {
	k, err := s.webhookRepo.GetSigningKeyByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get signing key: %w", err)
	}
	if k == nil {
		return fmt.Errorf("signing key not found")
	}
	if companyID != "" && k.CompanyID != companyID {
		return fmt.Errorf("forbidden")
	}
	if err := s.webhookRepo.DeleteSigningKey(ctx, id); err != nil {
		return fmt.Errorf("delete signing key: %w", err)
	}
	return nil
}

func (s *WebhookService) GetDelivery(ctx context.Context, companyID, id string) (*domain.WebhookDelivery, error) {
	d, err := s.webhookRepo.GetDeliveryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get delivery: %w", err)
	}
	if d == nil {
		return nil, nil
	}
	if companyID != "" && d.CompanyID != companyID {
		return nil, fmt.Errorf("forbidden")
	}
	return d, nil
}

func (s *WebhookService) TriggerWebhook(ctx context.Context, companyID string, eventType domain.WebhookEventType, payload domain.JSONMap) error {
	if payload == nil {
		payload = domain.JSONMap{}
	}

	webhooks, err := s.webhookRepo.ListActiveByEvent(ctx, companyID, eventType)
	if err != nil {
		return fmt.Errorf("list active webhooks: %w", err)
	}
	if len(webhooks) == 0 {
		return nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	for _, w := range webhooks {
		delivery := &domain.WebhookDelivery{
			ID:           uuid.New().String(),
			WebhookID:    w.ID,
			CompanyID:    companyID,
			EventType:    string(eventType),
			Payload:      payload,
			Status:       domain.WebhookDeliveryStatusPending,
			AttemptCount: 0,
			CreatedAt:    time.Now(),
		}

		if w.SigningKeyID != nil && *w.SigningKeyID != "" {
			k, keyErr := s.webhookRepo.GetSigningKeyByID(ctx, *w.SigningKeyID)
			if keyErr != nil {
				return fmt.Errorf("get signing key: %w", keyErr)
			}
			if k != nil && k.IsActive {
				sig, sigErr := s.signPayload(payloadBytes, k)
				if sigErr != nil {
					return fmt.Errorf("sign payload: %w", sigErr)
				}
				delivery.Signature = sig
			}
		}

		if err := s.webhookRepo.CreateDelivery(ctx, delivery); err != nil {
			return fmt.Errorf("create delivery: %w", err)
		}
	}
	return nil
}

func (s *WebhookService) ProcessDeliveryQueue(ctx context.Context) {
	ticker := time.NewTicker(s.queueInterval)
	defer ticker.Stop()
	log.Println("webhook delivery queue worker started")

	for {
		if err := s.processDeliveryBatch(ctx); err != nil {
			log.Printf("webhook queue process error: %v", err)
		}

		select {
		case <-ctx.Done():
			log.Println("webhook delivery queue worker stopped")
			return
		case <-ticker.C:
		}
	}
}

func (s *WebhookService) processDeliveryBatch(ctx context.Context) error {
	deliveries, err := s.webhookRepo.ListPendingDeliveries(ctx, s.queueBatchSize)
	if err != nil {
		return fmt.Errorf("list pending deliveries: %w", err)
	}
	for _, d := range deliveries {
		if err := s.processSingleDelivery(ctx, d); err != nil {
			log.Printf("webhook process delivery failed (id=%s): %v", d.ID, err)
		}
	}
	return nil
}

func (s *WebhookService) processSingleDelivery(ctx context.Context, d *domain.WebhookDelivery) error {
	w, err := s.webhookRepo.GetByID(ctx, d.WebhookID)
	if err != nil {
		return fmt.Errorf("get webhook: %w", err)
	}
	if w == nil || !w.IsActive {
		msg := "webhook not found or inactive"
		d.Status = domain.WebhookDeliveryStatusFailed
		d.AttemptCount++
		d.ResponseBody = &msg
		if updateErr := s.webhookRepo.UpdateDelivery(ctx, d); updateErr != nil {
			return fmt.Errorf("mark delivery failed: %w", updateErr)
		}
		return nil
	}

	payloadBytes, err := json.Marshal(d.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	signature := d.Signature
	if w.SigningKeyID != nil && *w.SigningKeyID != "" {
		k, keyErr := s.webhookRepo.GetSigningKeyByID(ctx, *w.SigningKeyID)
		if keyErr != nil {
			return fmt.Errorf("get signing key: %w", keyErr)
		}
		if k != nil && k.IsActive {
			sig, sigErr := s.signPayload(payloadBytes, k)
			if sigErr != nil {
				return fmt.Errorf("sign payload: %w", sigErr)
			}
			signature = sig
		}
	}

	d.Status = domain.WebhookDeliveryStatusDelivering
	d.Signature = signature
	d.AttemptCount++
	d.NextRetryAt = nil
	if err := s.webhookRepo.UpdateDelivery(ctx, d); err != nil {
		return fmt.Errorf("mark delivery delivering: %w", err)
	}

	statusCode, respBody, sendErr := s.sendWebhookRequest(ctx, w, d.EventType, d.ID, payloadBytes, signature)
	if sendErr == nil && statusCode >= 200 && statusCode < 300 {
		now := time.Now()
		d.Status = domain.WebhookDeliveryStatusSuccess
		d.HTTPStatus = &statusCode
		d.ResponseBody = &respBody
		d.DeliveredAt = &now
		d.NextRetryAt = nil
		if err := s.webhookRepo.UpdateDelivery(ctx, d); err != nil {
			return fmt.Errorf("mark delivery success: %w", err)
		}
		return nil
	}

	policy := retryPolicyFromJSONMap(w.RetryPolicy)
	if d.AttemptCount >= policy.MaxAttempts {
		d.Status = domain.WebhookDeliveryStatusFailed
		d.NextRetryAt = nil
	} else {
		next := time.Now().Add(calcBackoff(d.AttemptCount, policy))
		d.Status = domain.WebhookDeliveryStatusRetryLater
		d.NextRetryAt = &next
	}

	if sendErr != nil {
		respBody = sendErr.Error()
	}
	if statusCode > 0 {
		d.HTTPStatus = &statusCode
	}
	d.ResponseBody = &respBody
	if err := s.webhookRepo.UpdateDelivery(ctx, d); err != nil {
		return fmt.Errorf("update failed delivery: %w", err)
	}
	return nil
}

func (s *WebhookService) sendWebhookRequest(
	ctx context.Context,
	w *domain.Webhook,
	eventType string,
	deliveryID string,
	payload []byte,
	signature string,
) (statusCode int, responseBody string, err error) {
	timeout := w.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultWebhookTimeoutSeconds
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, w.URL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-Delivery-ID", deliveryID)
	if signature != "" {
		head := w.SecretHeader
		if head == "" {
			head = "X-Webhook-Signature"
		}
		req.Header.Set(head, signature)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		return resp.StatusCode, "", fmt.Errorf("read webhook response: %w", readErr)
	}
	return resp.StatusCode, string(body), nil
}

func (s *WebhookService) signPayload(payload []byte, k *domain.WebhookSigningKey) (string, error) {
	switch k.KeyType {
	case domain.WebhookSigningKeyEd25519:
		return s.generateEd25519Signature(payload, k.SecretKeyEnc)
	case domain.WebhookSigningKeyHMACSHA256, "":
		return s.generateHMACSignature(payload, k.SecretKeyEnc), nil
	default:
		return "", fmt.Errorf("unsupported signing key type: %s", k.KeyType)
	}
}

func (s *WebhookService) generateHMACSignature(payload []byte, secret []byte) string {
	mac := hmac.New(sha256.New, normalizeKeyBytes(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *WebhookService) generateEd25519Signature(payload []byte, secret []byte) (string, error) {
	key := normalizeKeyBytes(secret)
	var privateKey ed25519.PrivateKey
	switch len(key) {
	case ed25519.SeedSize:
		privateKey = ed25519.NewKeyFromSeed(key)
	case ed25519.PrivateKeySize:
		privateKey = ed25519.PrivateKey(key)
	default:
		return "", fmt.Errorf("invalid ed25519 private key size")
	}
	sig := ed25519.Sign(privateKey, payload)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (s *WebhookService) verifySignature(payload []byte, signature string, k *domain.WebhookSigningKey) bool {
	if k == nil || signature == "" {
		return false
	}

	sig := strings.TrimSpace(signature)
	sig = strings.TrimPrefix(sig, "sha256=")

	switch k.KeyType {
	case domain.WebhookSigningKeyEd25519:
		pub := normalizeKeyBytes([]byte(k.PublicKey))
		if len(pub) != ed25519.PublicKeySize {
			return false
		}
		sigBytes := decodeMaybeBase64OrHex(sig)
		if len(sigBytes) != ed25519.SignatureSize {
			return false
		}
		return ed25519.Verify(ed25519.PublicKey(pub), payload, sigBytes)
	case domain.WebhookSigningKeyHMACSHA256, "":
		expected := s.generateHMACSignature(payload, k.SecretKeyEnc)
		return hmac.Equal([]byte(expected), []byte(sig))
	default:
		return false
	}
}

func validateWebhookURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return fmt.Errorf("invalid webhook url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("webhook url must be http or https")
	}

	// SSRF 防护：禁止内网/本地地址
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("webhook url missing host")
	}
	hostLower := strings.ToLower(host)

	// 禁止 localhost
	if hostLower == "localhost" || hostLower == "127.0.0.1" || hostLower == "[::1]" || hostLower == "::1" {
		return fmt.Errorf("webhook url cannot point to localhost")
	}

	// 禁止私有 IP (10.0.0.0/8)
	if strings.HasPrefix(hostLower, "10.") {
		return fmt.Errorf("webhook url cannot point to private IP (10.x.x.x)")
	}
	// 禁止私有 IP (172.16.0.0/12)
	if matched, _ := regexp.MatchString(`^172\.(1[6-9]|2[0-9]|3[01])\.`, hostLower); matched {
		return fmt.Errorf("webhook url cannot point to private IP (172.16-31.x.x)")
	}
	// 禁止私有 IP (192.168.0.0/16)
	if strings.HasPrefix(hostLower, "192.168.") {
		return fmt.Errorf("webhook url cannot point to private IP (192.168.x.x)")
	}
	// 禁止回环地址 (127.0.0.0/8)
	if strings.HasPrefix(hostLower, "127.") {
		return fmt.Errorf("webhook url cannot point to loopback address (127.x.x.x)")
	}
	// 禁止链路本地地址 (169.254.0.0/16)
	if strings.HasPrefix(hostLower, "169.254.") {
		return fmt.Errorf("webhook url cannot point to link-local address (169.254.x.x)")
	}

	return nil
}

func toEventStringList(events []domain.WebhookEventType) domain.StringList {
	out := make(domain.StringList, 0, len(events))
	for _, evt := range events {
		e := strings.TrimSpace(string(evt))
		if e == "" {
			continue
		}
		out = append(out, e)
	}
	if len(out) == 0 {
		return domain.StringList{}
	}
	return out
}

func normalizeRetryPolicy(in *domain.RetryPolicy) domain.RetryPolicy {
	p := domain.RetryPolicy{MaxAttempts: 3, BackoffBase: 2, BackoffMax: 3600}
	if in == nil {
		return p
	}
	if in.MaxAttempts > 0 {
		p.MaxAttempts = in.MaxAttempts
	}
	if in.BackoffBase > 1 {
		p.BackoffBase = in.BackoffBase
	}
	if in.BackoffMax > 0 {
		p.BackoffMax = in.BackoffMax
	}
	return p
}

func retryPolicyToJSONMap(p domain.RetryPolicy) domain.JSONMap {
	return domain.JSONMap{
		"max_attempts": p.MaxAttempts,
		"backoff_base": p.BackoffBase,
		"backoff_max":  p.BackoffMax,
	}
}

func retryPolicyFromJSONMap(m domain.JSONMap) domain.RetryPolicy {
	out := normalizeRetryPolicy(nil)
	if m == nil {
		return out
	}
	if v, ok := readIntFromAny(m["max_attempts"]); ok && v > 0 {
		out.MaxAttempts = v
	}
	if v, ok := readIntFromAny(m["backoff_base"]); ok && v > 1 {
		out.BackoffBase = v
	}
	if v, ok := readIntFromAny(m["backoff_max"]); ok && v > 0 {
		out.BackoffMax = v
	}
	return out
}

func calcBackoff(attempt int, p domain.RetryPolicy) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	seconds := math.Pow(float64(p.BackoffBase), float64(attempt-1))
	if seconds > float64(p.BackoffMax) {
		seconds = float64(p.BackoffMax)
	}
	if seconds < 1 {
		seconds = 1
	}
	return time.Duration(seconds) * time.Second
}

func readIntFromAny(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}

func normalizeKeyBytes(raw []byte) []byte {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return []byte{}
	}
	if decoded, err := base64.StdEncoding.DecodeString(string(trimmed)); err == nil && len(decoded) > 0 {
		return decoded
	}
	if decoded, err := hex.DecodeString(string(trimmed)); err == nil && len(decoded) > 0 {
		return decoded
	}
	return trimmed
}

func decodeMaybeBase64OrHex(v string) []byte {
	if b, err := base64.StdEncoding.DecodeString(v); err == nil {
		return b
	}
	if b, err := hex.DecodeString(v); err == nil {
		return b
	}
	return []byte(v)
}
