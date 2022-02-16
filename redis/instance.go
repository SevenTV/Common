package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Instance interface {
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
	ComposeKey(svc, name string) Key
	RawClient() *redis.Client
}

type redisInst struct {
	cl      *redis.Client
	sub     *redis.PubSub
	subsMtx sync.Mutex
	subs    map[Key][]*redisSub
}

type redisSub struct {
	ch chan string
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
func (inst *redisInst) Subscribe(ctx context.Context, ch chan string, subscribeTo ...Key) {
	inst.subsMtx.Lock()
	defer inst.subsMtx.Unlock()
	localSub := &redisSub{ch}
	for _, e := range subscribeTo {
		if _, ok := inst.subs[e]; !ok {
			_ = inst.sub.Subscribe(ctx, e.String())
		}
		inst.subs[e] = append(inst.subs[e], localSub)
	}

	go func() {
		<-ctx.Done()
		inst.subsMtx.Lock()
		defer inst.subsMtx.Unlock()
		for _, e := range subscribeTo {
			for i, v := range inst.subs[e] {
				if v == localSub {
					if i != len(inst.subs[e])-1 {
						inst.subs[e][i] = inst.subs[e][len(inst.subs[e])-1]
					}
					inst.subs[e] = inst.subs[e][:len(inst.subs[e])-1]
					if len(inst.subs[e]) == 0 {
						delete(inst.subs, e)
						if err := inst.sub.Unsubscribe(context.Background(), e.String()); err != nil {
							logrus.WithError(err).Error("failed to unsubscribe")
						}
					}
					break
				}
			}
		}
	}()
}

type Key string

var Nil = redis.Nil

func (k Key) String() string {
	return string(k)
}
