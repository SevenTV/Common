package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Instance interface {
	Ping(ctx context.Context) error
	RawClient() *redis.Client
	ComposeKey(svc, name string) Key
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

func (i *redisInst) ComposeKey(svc, name string) Key {
	return Key(fmt.Sprintf("7tv-%s:%s", svc, name))
}

type Key string

var Nil = redis.Nil

func (k Key) String() string {
	return string(k)
}
