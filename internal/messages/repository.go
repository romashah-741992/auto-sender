package messages

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Repository abstraction used by service & scheduler.
type Repository interface {
	GetPendingMessages(ctx context.Context, limit int) ([]Message, error)
	MarkAsSent(ctx context.Context, id int64, externalID string, sentAt time.Time) error
	MarkAsFailed(ctx context.Context, id int64, errMsg string) error
	ListSentMessages(ctx context.Context, limit, offset int) ([]Message, error)
}

/* ------------------ MySQL implementation ------------------ */

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) GetPendingMessages(ctx context.Context, limit int) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recipient, content, status, external_message_id, sent_at, created_at, updated_at
		FROM messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var status string
		var externalID sql.NullString
		var sentAt sql.NullTime

		if err := rows.Scan(
			&m.ID,
			&m.Recipient,
			&m.Content,
			&status,
			&externalID,
			&sentAt,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		m.Status = Status(status)
		if externalID.Valid {
			m.ExternalMessageID = &externalID.String
		}
		if sentAt.Valid {
			t := sentAt.Time
			m.SentAt = &t
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *SQLRepository) MarkAsSent(ctx context.Context, id int64, externalID string, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages
		SET status = 'sent',
		    external_message_id = ?,
		    sent_at = ?,
		    updated_at = ?
		WHERE id = ?`,
		externalID, sentAt, time.Now(), id)
	return err
}

func (r *SQLRepository) MarkAsFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages
		SET status = 'failed',
		    error_text = ?,
		    updated_at = ?
		WHERE id = ?`,
		errMsg, time.Now(), id)
	return err
}

func (r *SQLRepository) ListSentMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recipient, content, status, external_message_id, sent_at, created_at, updated_at
		FROM messages
		WHERE status = 'sent'
		ORDER BY sent_at DESC
		LIMIT ? OFFSET ?`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		var status string
		var externalID sql.NullString
		var sentAt sql.NullTime

		if err := rows.Scan(
			&m.ID,
			&m.Recipient,
			&m.Content,
			&status,
			&externalID,
			&sentAt,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		m.Status = Status(status)
		if externalID.Valid {
			m.ExternalMessageID = &externalID.String
		}
		if sentAt.Valid {
			t := sentAt.Time
			m.SentAt = &t
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

/* ------------------ In-memory implementation (for local dev) ------------------ */

type InMemoryRepository struct {
	mu       sync.Mutex
	messages []Message
	nextID   int64
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		messages: make([]Message, 0),
		nextID:   1,
	}
}

func (r *InMemoryRepository) AddMessage(recipient, content string) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	id := r.nextID
	r.nextID++

	msg := Message{
		ID:        id,
		Recipient: recipient,
		Content:   content,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.messages = append(r.messages, msg)
	return id
}

func (r *InMemoryRepository) GetPendingMessages(ctx context.Context, limit int) ([]Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Message, 0, limit)
	for _, m := range r.messages {
		if m.Status == StatusPending {
			result = append(result, m)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (r *InMemoryRepository) MarkAsSent(ctx context.Context, id int64, externalID string, sentAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.messages {
		if r.messages[i].ID == id {
			r.messages[i].Status = StatusSent
			r.messages[i].ExternalMessageID = &externalID
			r.messages[i].SentAt = &sentAt
			r.messages[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("message %d not found", id)
}

func (r *InMemoryRepository) MarkAsFailed(ctx context.Context, id int64, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.messages {
		if r.messages[i].ID == id {
			r.messages[i].Status = StatusFailed
			r.messages[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("message %d not found", id)
}

func (r *InMemoryRepository) ListSentMessages(ctx context.Context, limit, offset int) ([]Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sent := make([]Message, 0)
	for _, m := range r.messages {
		if m.Status == StatusSent {
			sent = append(sent, m)
		}
	}

	if offset >= len(sent) {
		return []Message{}, nil
	}
	end := offset + limit
	if end > len(sent) {
		end = len(sent)
	}
	return sent[offset:end], nil
}

// SeedDummyMessages is for dev runs without DB.
func SeedDummyMessages(r *InMemoryRepository) {
	r.AddMessage("+905551111111", "Insider - Project 1")
	r.AddMessage("+905552222222", "Insider - Project 2")
	r.AddMessage("+905553333333", "Another test message")
}
