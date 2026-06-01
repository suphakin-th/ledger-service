package ports

import "context"

type EventBus interface {
	Publish(ctx context.Context, channel string, payload any) error
}
