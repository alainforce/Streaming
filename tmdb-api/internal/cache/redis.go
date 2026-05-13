// internal/cache/redis.go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates and validates a Redis client connection.
// Like pgxpool, we create this ONCE at startup and share it everywhere.
func NewRedisClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // No password in local dev. In prod, use env var.
		DB:           0,  // Redis has 16 logical DBs (0-15). 0 is the default.
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10, // Max simultaneous connections — mirrors our pgxpool setting.
	})

	// Validate the connection is reachable at startup.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: failed to connect to Redis: %w", err)
	}

	return client, nil
}
