package mutations

import (
	"sync"

	"github.com/seventv/common/mongo"
	"github.com/seventv/common/redis"
	"github.com/seventv/common/svc/s3"
)

type Mutate struct {
	mongo mongo.Instance
	redis redis.Instance
	s3    s3.Instance
	mx    map[string]*sync.Mutex
}

func New(opt InstanceOptions) *Mutate {
	return &Mutate{
		mongo: opt.Mongo,
		redis: opt.Redis,
		s3:    opt.S3,
		mx:    map[string]*sync.Mutex{},
	}
}

type InstanceOptions struct {
	Mongo mongo.Instance
	Redis redis.Instance
	S3    s3.Instance
}
