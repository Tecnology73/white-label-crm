package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

var (
	Client *redis.Client
)

func NewConnection(opts *redis.Options) *redis.Client {
	Client = redis.NewClient(opts)
	if err := Client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("[initRedis] %v\n", err)
	}

	return Client
}
