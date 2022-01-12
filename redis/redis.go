package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func Setup(ctx context.Context, opt SetupOptions) (Instance, error) {
	var rc *redis.Client

	if len(opt.Addresses) == 0 {
		logrus.Fatal("you must provide at least one redis address")
	}

	if opt.Sentinel {
		rc = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       "master",
			SentinelAddrs:    opt.Addresses,
			SentinelUsername: opt.Username,
			SentinelPassword: opt.Password,
			DB:               opt.Database,
		})
	} else {
		rc = redis.NewClient(&redis.Options{
			Addr:     opt.Addresses[0],
			Username: opt.Username,
			Password: opt.Password,
			DB:       opt.Database,
		})
	}

	if err := rc.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logrus.Info("redis, ok")

	return &redisInst{cl: rc}, nil
}

type SetupOptions struct {
	Username string
	Password string
	Database int

	Addresses []string
	Sentinel  bool
}
