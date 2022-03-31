package query

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/redis"
	"github.com/SevenTV/Common/structures/v3"
	"github.com/SevenTV/Common/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type Query struct {
	mongo mongo.Instance
	redis redis.Instance
	c     *cache.Cache
	mx    map[string]*sync.Mutex
}

func New(mongoInst mongo.Instance, redisInst redis.Instance) *Query {
	return &Query{
		mongo: mongoInst,
		redis: redisInst,
		c:     cache.New(time.Minute*1, time.Minute*5),
		mx:    map[string]*sync.Mutex{},
	}
}

func (q *Query) lock(tag string) *sync.Mutex {
	l, ok := q.mx[tag]
	if !ok {
		l = &sync.Mutex{}
		q.mx[tag] = l
	}
	l.Lock()
	return l
}

func (q *Query) key(tag string) redis.Key {
	return q.redis.ComposeKey("common", fmt.Sprintf("cache:%s", tag))
}

// getFromMemCache retrieve a cached item
func (q *Query) getFromMemCache(ctx context.Context, key redis.Key, i interface{}) bool {
	var (
		s   string
		err error
	)
	v, ok := q.c.Get(key.String())

	if ok {
		s = v.(string)
	} else {
		s, err = q.redis.Get(ctx, key)
	}
	if len(s) > 0 {
		if err := multierror.Append(err, json.Unmarshal(utils.S2B(s), i)).ErrorOrNil(); err != nil {
			if err != redis.Nil {
				logrus.WithError(err).WithField("key", key).Error("redis, failed to retrieve a cache query item")
			}
			return false
		} else {
			return true
		}
	}
	return false
}

// setInMemCache sets an item into the cache
func (q *Query) setInMemCache(ctx context.Context, key redis.Key, i interface{}, ex time.Duration) error {
	b, err := json.Marshal(i)
	if err == nil {
		s := utils.B2S(b)
		q.c.Add(key.String(), s, ex)
		if err = q.redis.SetEX(ctx, key, s, ex); err != nil {
			return err
		}
	}
	return nil
}

type QueryResult[T QueriableType] struct {
	items []*T
	err   error
}

type QueriableType interface {
	structures.User | structures.Emote | structures.EmoteSet | structures.Message | structures.Role
}

func (qr *QueryResult[T]) setItems(items []*T) *QueryResult[T] {
	qr.items = items
	return qr
}

func (qr *QueryResult[T]) setError(err error) *QueryResult[T] {
	qr.err = err
	return qr
}

func (qr *QueryResult[T]) Error() error {
	return qr.err
}

func (qr *QueryResult[T]) First() (*T, error) {
	if qr.err != nil {
		return nil, qr.err
	}
	if len(qr.items) == 0 {
		return nil, errors.ErrNoItems()
	}
	return qr.items[0], nil
}

func (qr *QueryResult[T]) Index(pos int) (*T, error) {
	if qr.err != nil {
		return nil, qr.err
	}
	if pos > len(qr.items)-1 {
		return nil, errors.ErrNoItems()
	}
	return qr.items[pos], nil
}

func (qr *QueryResult[T]) Last() (*T, error) {
	if qr.err != nil {
		return nil, qr.err
	}
	if len(qr.items) == 0 {
		return nil, errors.ErrNoItems()
	}
	return qr.items[len(qr.items)-1], nil
}

func (qr *QueryResult[T]) Items() ([]*T, error) {
	return qr.items, qr.err
}

func (qr *QueryResult[T]) Empty() bool {
	return len(qr.items) == 0
}
