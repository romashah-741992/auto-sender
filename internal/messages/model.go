package messages

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"

	MaxContentLength = 160
)

type Message struct {
	ID                int64      `json:"id"`
	Recipient         string     `json:"recipient"`
	Content           string     `json:"content"`
	Status            Status     `json:"status"`
	ExternalMessageID *string    `json:"externalMessageId,omitempty"`
	SentAt            *time.Time `json:"sentAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
