package storage

import (
	"context"
	"time"
)

type CacheStorage interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Remove(ctx context.Context, key string) error
}
