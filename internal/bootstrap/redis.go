package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func InitiateRedis(cfg *config.Config) *Cache {
	if cfg.RedisHost == "" {
		log.Println("Redis host not configured. Running without cache.")
		return &Cache{}
	}

	port := cfg.RedisPort
	if port == "" {
		port = "6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, port),
		Password: cfg.RedisPassword,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v. Running without cache.", err)
		return &Cache{}
	}

	log.Println("Connected to Redis successfully.")
	return &Cache{client: client}
}

func (c *Cache) IsEnabled() bool {
	return c != nil && c.client != nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !c.IsEnabled() {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, expiration).Err()
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.IsEnabled() {
		return redis.Nil
	}
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if !c.IsEnabled() || len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) InvalidateByPrefix(ctx context.Context, prefix string) error {
	if !c.IsEnabled() || prefix == "" {
		return nil
	}

	iter := c.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	var keysToDelete []string

	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
		if len(keysToDelete) >= 100 {
			if err := c.client.Del(ctx, keysToDelete...).Err(); err != nil {
				return err
			}
			keysToDelete = keysToDelete[:0]
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keysToDelete) > 0 {
		return c.client.Del(ctx, keysToDelete...).Err()
	}

	return nil
}
