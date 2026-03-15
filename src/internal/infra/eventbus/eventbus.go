package eventbus

import (
	"context"
	"suscord/internal/domain/eventbus"
	"time"

	"go.uber.org/zap"
)

type EventBus struct {
	timeout  time.Duration
	logger   *zap.SugaredLogger
	handlers map[string][]eventbus.Handler
}

func NewEventBus(logger *zap.SugaredLogger, timeout time.Duration) *EventBus {
	return &EventBus{
		timeout:  timeout,
		logger:   logger,
		handlers: make(map[string][]eventbus.Handler),
	}
}

func (eb *EventBus) Subscribe(key string, handler eventbus.Handler) {
	if _, ok := eb.handlers[key]; ok {
		eb.handlers[key] = make([]eventbus.Handler, 0)
	}
	eb.handlers[key] = append(eb.handlers[key], handler)
}

func (eb *EventBus) Publish(key string, payload any) {
	var err error

	go func() {
		for _, handler := range eb.handlers[key] {
			ctx, cansel := context.WithTimeout(context.Background(), eb.timeout)
			defer cansel()
			if err = handler(ctx, payload); err != nil {
				eb.logger.Errorw("eventbus error: "+key, "err", err, "payload", payload)
			}
		}
	}()
}
