package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func Setup(ctx context.Context, opt SetupOptions) {
	rc := redis.NewClient(&redis.Options{
		Addr: opt.URI,
		DB:   opt.DB,
	})

	if err := rc.Ping(ctx).Err(); err != nil {
		panic(err)
	} else {
		log.Info("redis, ok")
	}
}

type SetupOptions struct {
	URI string
	DB  int
}
