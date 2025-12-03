package redis

import (
	"context"
	"fmt"
	"time"

	rds "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *rds.Client
}

func New(addr string) *Client {
	return &Client{
		rdb: rds.NewClient(&rds.Options{
			Addr: addr,
		}),
	}
}

func (c *Client) CacheSentMessage(ctx context.Context, id int64, externalID string, sentAt time.Time) error {
	key := fmt.Sprintf("message:%d", id)
	fields := map[string]string{
		"messageId": externalID,
		"sentAt":    sentAt.Format(time.RFC3339),
	}
	if err := c.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	return c.rdb.Expire(ctx, key, 24*time.Hour).Err()
}
