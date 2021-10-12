package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func Setup(ctx context.Context, opt SetupOptions) (Instance, error) {
	opts, err := redis.ParseURL(opt.URI)
	if err != nil {
		return nil, err
	}

	rc := redis.NewClient(opts)

	if err := rc.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logrus.Info("redis, ok")

	return &redisInst{cl: rc}, nil
}

type SetupOptions struct {
	URI string
}
