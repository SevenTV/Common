package mutations

import (
	"sync"

	"github.com/seventv/common/mongo"
	"github.com/seventv/common/redis"
)

type Mutate struct {
	mongo mongo.Instance
	redis redis.Instance
	mx    map[string]*sync.Mutex
}

func New(mongoInst mongo.Instance, redisInst redis.Instance) *Mutate {
	return &Mutate{
		mongo: mongoInst,
		redis: redisInst,
		mx:    map[string]*sync.Mutex{},
	}
}
