package server

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

func (r *RateLimiter) Allow(codeID string) bool {
	key := fmt.Sprintf("ratelimit:%s", codeID)
	now := time.Now().Unix()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	txp := r.client.TxPipeline()
	defer txp.Close()
	cmd1 := txp.HIncrBy(ctx, key, "count", 1)
	cmd2 := txp.Expire(ctx, key, r.window)
	cmd3 := txp.HGet(ctx, key, "timestamp")

	if _, err := txp.Exec(context.Background()); err != nil {
		log.Error("Transaction failed:", err)
	}
	count, err := cmd1.Result()
	if err != nil {
		fmt.Println("Error increment count:", err)
	}
	if _, err := cmd2.Result(); err != nil {
		fmt.Println("Error setting expiration:", err)
	}
	timestamp, err := cmd3.Int64()
	if err != nil {
		fmt.Println("Error get timestamp:", err)
	}

	if now-int64(r.window.Seconds()) > timestamp {
		r.client.HSet(ctx, key, "timestamp", now)
		r.client.HSet(ctx, key, "count", 1)
		return true
	}

	if count > int64(r.limit) {
		return false
	}

	return true
}
