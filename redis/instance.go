package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Instance interface {
	Ping(ctx context.Context) error
	RawClient() *redis.Client
}

type redisInst struct {
	cl *redis.Client
}

func (i *redisInst) Ping(ctx context.Context) error {
	return i.cl.Ping(ctx).Err()
}

func (i *redisInst) RawClient() *redis.Client {
	return i.cl
}
