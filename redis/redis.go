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

	inst := &redisInst{
		cl:   rc,
		sub:  rc.Subscribe(context.Background()),
		subs: map[Key][]*redisSub{},
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithField("err", err).Fatal("panic in subs")
			}
		}()
		ch := inst.sub.Channel()
		var msg *redis.Message
		for {
			msg = <-ch
			payload := msg.Payload // dont change we want to copy the memory due to concurrency.
			inst.subsMtx.Lock()
			for _, s := range inst.subs[Key(msg.Channel)] {
				select {
				case s.ch <- payload:
				default:
					logrus.Warn("channel blocked dropping message: ", msg.Channel)
				}
			}
			inst.subsMtx.Unlock()
		}
	}()

	return inst, nil
}

type SetupOptions struct {
	Username string
	Password string
	Database int

	Addresses []string
	Sentinel  bool
}
