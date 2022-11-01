package utils

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var (
	RDB *redis.Client
)

func InitKVEngine(ctx context.Context, addr, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}
