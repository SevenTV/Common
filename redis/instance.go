package redis

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/seventv/common/sync_map"
	"github.com/seventv/common/utils"
	"go.uber.org/zap"
)

type Instance interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key Key) (string, error)
	Set(ctx context.Context, key Key, value interface{}) error
	SetEX(ctx context.Context, key Key, value interface{}, expiry time.Duration) error
	Exists(ctx context.Context, keys ...Key) (int, error)
	IncrBy(ctx context.Context, key Key, amount int) (int, error)
	DecrBy(ctx context.Context, key Key, amount int) (int, error)
	Expire(ctx context.Context, key Key, expiry time.Duration) error
	Del(ctx context.Context, keys ...Key) (int, error)
	TTL(ctx context.Context, key Key) (time.Duration, error)
	Pipeline(ctx context.Context) redis.Pipeliner
	Subscribe(ctx context.Context, ch chan string, subscribeTo ...Key)
	ComposeKey(svc string, args ...string) Key
	RawClient() *redis.Client
}

type redisInst struct {
	cl  *redis.Client
	sub *redis.PubSub

	subs sync_map.Map[Key, *subController]
}

type subController struct {
	evt   Key
	i     *uint64
	count *int64
	subs  sync_map.Map[uint64, chan string]
	inst  *redisInst
}

func (s *subController) Subscribe(ch chan string) func() {
	i := atomic.AddUint64(s.i, 1)
	atomic.AddInt64(s.count, 1)
	s.subs.Store(i, ch)
	return func() {
		if atomic.AddInt64(s.count, -1) == 0 {
			s.inst.subs.Delete(Key(s.evt.String()))
			if err := s.inst.sub.Unsubscribe(context.Background(), s.evt.String()); err != nil {
				zap.S().Errorw("failed to unsubscribe",
					"error", err,
				)
			}
		}
		s.subs.Delete(i)

	}
}

func (i *redisInst) Ping(ctx context.Context) error {
	return i.cl.Ping(ctx).Err()
}

func (i *redisInst) RawClient() *redis.Client {
	return i.cl
}

func (i *redisInst) ComposeKey(svc string, args ...string) Key {
	return Key(fmt.Sprintf("%s:%s", svc, strings.Join(args, ":")))
}

func (r *redisInst) Get(ctx context.Context, key Key) (string, error) {
	return r.RawClient().Get(ctx, string(key)).Result()
}

func (r *redisInst) Set(ctx context.Context, key Key, value interface{}) error {
	return r.RawClient().Set(ctx, string(key), value, 0).Err()
}

func (r *redisInst) SetEX(ctx context.Context, key Key, value interface{}, expiry time.Duration) error {
	return r.RawClient().SetEX(ctx, string(key), value, expiry).Err()
}

func (r *redisInst) Exists(ctx context.Context, keys ...Key) (int, error) {
	k := make([]string, len(keys))
	for i, v := range keys {
		k[i] = string(v)
	}
	i, err := r.RawClient().Exists(ctx, k...).Result()
	return int(i), err
}

func (r *redisInst) IncrBy(ctx context.Context, key Key, amount int) (int, error) {
	i, err := r.RawClient().IncrBy(ctx, string(key), int64(amount)).Result()
	return int(i), err
}

func (r *redisInst) DecrBy(ctx context.Context, key Key, amount int) (int, error) {
	i, err := r.RawClient().DecrBy(ctx, string(key), int64(amount)).Result()
	return int(i), err
}

func (r *redisInst) Expire(ctx context.Context, key Key, expiry time.Duration) error {
	return r.RawClient().Expire(ctx, string(key), expiry).Err()
}

func (r *redisInst) TTL(ctx context.Context, key Key) (time.Duration, error) {
	return r.RawClient().TTL(ctx, string(key)).Result()
}

func (r *redisInst) Del(ctx context.Context, keys ...Key) (int, error) {
	k := make([]string, len(keys))
	for i, v := range keys {
		k[i] = string(v)
	}
	i, err := r.RawClient().Del(ctx, k...).Result()
	return int(i), err
}

func (r *redisInst) Pipeline(ctx context.Context) redis.Pipeliner {
	return r.RawClient().Pipeline()
}

// Subscribe to a channel on Redis
func (r *redisInst) Subscribe(ctx context.Context, ch chan string, subscribeTo ...Key) {
	for _, e := range subscribeTo {
		sub, ok := r.subs.LoadOrStore(e, &subController{
			evt:   e,
			i:     utils.PointerOf(uint64(0)),
			count: utils.PointerOf(int64(0)),
			inst:  r,
		})
		if !ok {
			_ = r.sub.Subscribe(ctx, e.String())
		}
		defer sub.Subscribe(ch)()
	}

	<-ctx.Done()
}

type Key string

var Nil = redis.Nil

func (k Key) String() string {
	return string(k)
}
