package cache

import (
	"context"
	"log"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	InvalidateByPrefix(ctx context.Context, prefix string) error
}

// Remember and Forget treat the cache as an optimisation: a failed write or invalidation costs a slow
// read, never a failed request, so the error is logged rather than returned.
func Remember(ctx context.Context, c Cache, key string, value interface{}, expiration time.Duration) {
	if err := c.Set(ctx, key, value, expiration); err != nil {
		log.Printf("cache: set %s: %v", key, err)
	}
}

func Forget(ctx context.Context, c Cache, keys ...string) {
	if err := c.Delete(ctx, keys...); err != nil {
		log.Printf("cache: delete %v: %v", keys, err)
	}
}
