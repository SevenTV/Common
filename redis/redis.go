package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redis_sync "github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"go.uber.org/zap"
)

func Setup(ctx context.Context, opt SetupOptions) (Instance, error) {
	var rc *redis.Client

	if len(opt.Addresses) == 0 {
		return nil, fmt.Errorf("you must provide at least one redis address")
	}

	if opt.Sentinel {
		rc = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       opt.MasterName,
			SentinelAddrs:    opt.Addresses,
			SentinelUsername: opt.Username,
			SentinelPassword: opt.Password,
			Username:         opt.Username,
			Password:         opt.Password,
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

	inst := &redisInst{
		cl:  rc,
		sub: rc.Subscribe(context.Background()),
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorw("panic in subs",
					"error", err,
				)
			}
		}()
		ch := inst.sub.Channel()
		var msg *redis.Message
		for {
			msg = <-ch
			payload := msg.Payload // dont change we want to copy the memory due to concurrency.
			if subs, ok := inst.subs.Load(Key(msg.Channel)); ok {
				subs.subs.Range(func(key uint64, value chan string) bool {
					select {
					case value <- payload:
					default:
						zap.S().Warnw("channel blocked",
							"channel", msg.Channel,
						)
					}
					return true
				})
			}
		}
	}()

	if opt.EnableSync {
		pool := redis_sync.NewPool(rc)

		inst.sync = redsync.New(pool)
	}

	return inst, nil
}

type SetupOptions struct {
	MasterName string
	Username   string
	Password   string
	Database   int

	Addresses []string
	Sentinel  bool

	EnableSync bool
}
