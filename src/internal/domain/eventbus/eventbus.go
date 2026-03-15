package eventbus

import "context"

type Handler func(ctx context.Context, payload any) error

type EventBus interface {
	Subscribe(key string, handler Handler)
	Publish(key string, payload any)
}
