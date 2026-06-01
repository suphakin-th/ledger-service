package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type EventBus struct {
	client *redis.Client
}

func NewEventBus(addr string) *EventBus {
	return &EventBus{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (b *EventBus) Publish(ctx context.Context, channel string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return b.client.Publish(ctx, channel, data).Err()
}

func (b *EventBus) Close() error {
	return b.client.Close()
}
