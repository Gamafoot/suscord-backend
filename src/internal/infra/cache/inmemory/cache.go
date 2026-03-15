package inmemory

import (
	"context"
	derr "suscord/internal/domain/errors"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache struct {
	cache *cache.Cache
}

func NewCache() *Cache {
	return &Cache{cache: cache.New(5*time.Minute, 10*time.Minute)}
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.cache.Set(key, value, ttl)
	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	value, ok := c.cache.Get(key)
	if !ok {
		return nil, derr.ErrKeyNotFound
	}
	return value, nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}
