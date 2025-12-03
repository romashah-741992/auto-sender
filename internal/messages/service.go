package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RedisCache is implemented by internal/redis.Client
type RedisCache interface {
	CacheSentMessage(ctx context.Context, id int64, externalID string, sentAt time.Time) error
}

type Service struct {
	repo       Repository
	cache      RedisCache
	webhookURL string
	authKey    string
	httpClient *http.Client
}

func NewService(repo Repository, cache RedisCache, webhookURL, authKey string) *Service {
	var client *http.Client
	if webhookURL != "" {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &Service{
		repo:       repo,
		cache:      cache,
		webhookURL: webhookURL,
		authKey:    authKey,
		httpClient: client,
	}
}

func (s *Service) SendPendingMessages(ctx context.Context, limit int) error {
	msgs, err := s.repo.GetPendingMessages(ctx, limit)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		log.Println("[service] no pending messages")
		return nil
	}

	for _, m := range msgs {
		if len(m.Content) > MaxContentLength {
			log.Printf("[service] message %d too long (%d chars), marking failed\n", m.ID, len(m.Content))
			_ = s.repo.MarkAsFailed(ctx, m.ID, "content too long")
			continue
		}

		if err := s.sendMessage(ctx, m); err != nil {
			log.Printf("[service] send failed for message %d: %v\n", m.ID, err)
			_ = s.repo.MarkAsFailed(ctx, m.ID, err.Error())
		}
	}
	return nil
}

func (s *Service) ListSentMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	return s.repo.ListSentMessages(ctx, limit, offset)
}

/* ------------ Webhook call ------------ */

type webhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type webhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

func (s *Service) sendMessage(ctx context.Context, m Message) error {
	// If no webhook configured, simulate success (useful for local dev).
	if s.webhookURL == "" || s.httpClient == nil {
		log.Printf("[service] (dry-run) would send to=%s content=%q\n", m.Recipient, m.Content)
		externalID := uuid.NewString()
		sentAt := time.Now()
		if err := s.repo.MarkAsSent(ctx, m.ID, externalID, sentAt); err != nil {
			return err
		}
		if s.cache != nil {
			_ = s.cache.CacheSentMessage(ctx, m.ID, externalID, sentAt)
		}
		return nil
	}

	payload := webhookRequest{
		To:      m.Recipient,
		Content: m.Content,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.authKey != "" {
		req.Header.Set("x-ins-auth-key", s.authKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var wr webhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return err
	}

	// webhook.site free tier: static messageId.
	// We generate our own UUID to ensure uniqueness.
	externalID := uuid.NewString()
	if wr.MessageID != "" && wr.MessageID != "static" {
		externalID = wr.MessageID
	}

	sentAt := time.Now()
	if err := s.repo.MarkAsSent(ctx, m.ID, externalID, sentAt); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.CacheSentMessage(ctx, m.ID, externalID, sentAt)
	}
	return nil
}
